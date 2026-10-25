package launcher_test

import (
	"context"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/jellydn/knulli-random-game-picker/internal/launcher"
)

func TestSubmitCancellationAfterServerAcceptsDoesNotRetry(t *testing.T) {
	// Use the real endpoint: an injectable test transport would not exercise
	// the fixed-loopback contract. Do not run this alongside the Python fixture.
	listener, err := net.Listen("tcp4", "127.0.0.1:1234")
	if err != nil {
		t.Fatal(err)
	}
	requests := make(chan string, 8)
	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read submitted body: %v", err)
		}
		requests <- r.Method + " " + r.URL.RequestURI() + " " + string(data)
		<-r.Context().Done()
	})}
	defer server.Close()
	go server.Serve(listener)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	path := "/roms/日本語 'exact'; %2F.rom"
	go func() { done <- launcher.Submit(ctx, path) }()
	select {
	case got := <-requests:
		if want := "POST /launch " + path; got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server did not receive launch")
	}
	cancel()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("canceled unacknowledged launch reported success")
		}
	case <-time.After(time.Second):
		t.Fatal("cancellation did not interrupt response read")
	}
	select {
	case got := <-requests:
		t.Fatalf("launch retried: %q", got)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestInvalidRequestRejectedBeforeDial(t *testing.T) {
	for _, input := range []struct{ method, route, body string }{
		{"DELETE", "/launch", ""},
		{"POST", "http://example.com/launch", "rom"},
		{"POST", "/launch\r\nX-Injected: yes", "rom"},
		{"POST", "/launch", "rom\x00suffix"},
	} {
		_, err := launcher.Request(context.Background(), input.method, input.route, input.body)
		if err == nil || err.Error() != "Invalid ES IPC request" {
			t.Fatalf("invalid request %+v reached transport: %v", input, err)
		}
	}
}
