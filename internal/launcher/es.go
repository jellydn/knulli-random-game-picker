// Package launcher contains the bounded, loopback-only ES IPC transport.
package launcher

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const endpoint = "127.0.0.1:1234"

// Request opens a fresh connection for every request. No proxy, redirect,
// automatic retry, DNS lookup or general-purpose external URL is permitted.
func Request(ctx context.Context, method, route, body string) ([]byte, error) {
	if (method != "GET" && method != "POST") || !strings.HasPrefix(route, "/") || strings.ContainsAny(route, "\r\n") || strings.ContainsRune(body, 0) {
		return nil, errors.New("Invalid ES IPC request")
	}
	conn, err := (&net.Dialer{Timeout: 5 * time.Second}).DialContext(ctx, "tcp4", endpoint)
	if err != nil {
		return nil, fmt.Errorf("ES local API unavailable: %w", err)
	}
	defer conn.Close()
	stop := context.AfterFunc(ctx, func() { _ = conn.Close() })
	defer stop()
	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
	req := &http.Request{Method: method, URL: &url.URL{Path: route}, Host: endpoint, Header: http.Header{"Content-Type": {"text/plain"}}, Body: io.NopCloser(strings.NewReader(body)), ContentLength: int64(len(body)), Close: true}
	// Preserve escaped API route segments instead of escaping percent signs twice.
	req.URL, err = url.ParseRequestURI(route)
	if err != nil {
		return nil, err
	}
	if err = req.Write(conn); err != nil {
		return nil, fmt.Errorf("ES request interrupted; launch outcome may be unknown, do not retry: %w", err)
	}
	response, err := http.ReadResponse(bufio.NewReaderSize(io.LimitReader(conn, (64<<20)+65536+1), 65536), req)
	if err != nil {
		return nil, fmt.Errorf("Invalid ES response; launch outcome may be unknown, do not retry: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("ES returned HTTP %d", response.StatusCode)
	}
	if len(response.TransferEncoding) != 0 || response.ContentLength < 0 || response.ContentLength > 64<<20 {
		return nil, errors.New("Unsupported ES framing or response size")
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, (64<<20)+1))
	if err != nil || int64(len(data)) != response.ContentLength {
		return nil, errors.New("Incomplete ES response; launch outcome may be unknown, do not retry")
	}
	return data, nil
}

func Fetch(ctx context.Context, route string) ([]byte, error) { return Request(ctx, "GET", route, "") }
func Submit(ctx context.Context, path string) error {
	_, err := Request(ctx, "POST", "/launch", path)
	return err
}
