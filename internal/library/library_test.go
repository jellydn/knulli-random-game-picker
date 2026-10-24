package library_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jellydn/knulli-random-game-picker/internal/library"
	"math/rand/v2"
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoveryContract(t *testing.T) {
	root := t.TempDir()
	user := filepath.Join(root, "user")
	roms := filepath.Join(root, "relocated library")
	write := func(path, text string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	a := filepath.Join(roms, "odd 'name; $.rom")
	b := filepath.Join(roms, "second.rom")
	hidden := filepath.Join(roms, "hidden.rom")
	for _, path := range []string{a, b, hidden} {
		write(path, "fixture")
	}
	write(filepath.Join(root, "shared/es_systems.cfg"), "<systemList><system><name>wrong</name><path>/wrong</path></system></systemList>")
	write(filepath.Join(user, "es_systems_custom.cfg"), "<systemList><system><name>custom</name><fullname>Custom Console</fullname><path>"+roms+"</path></system></systemList>")
	write(filepath.Join(roms, "gamelist.xml"), "<gameList><game><path>./odd 'name; $.rom</path><name>Scraped</name><image>./images/art.png</image></game></gameList>")
	write(filepath.Join(roms, "images/art.png"), "art fixture")
	systems := []map[string]any{{"name": "custom", "gamesystem": "true"}, {"name": "favorites", "gamesystem": true, "collection": true}}
	games := []map[string]any{{"path": a, "name": "Live Title", "favorite": "true"}, {"path": b, "name": "Second"}, {"path": hidden, "hidden": true}, {"path": filepath.Join(roms, "missing.rom")}}
	fetch := func(_ context.Context, route string) ([]byte, error) {
		switch route {
		case "/systems":
			return json.Marshal(systems)
		case "/systems/custom/games":
			return json.Marshal(games)
		case "/systems/ports/games":
			return json.Marshal(games[:1])
		default:
			return nil, fmt.Errorf("Unexpected route %s", route)
		}
	}
	loc := library.Locations{User: user, Shared: filepath.Join(root, "shared"), Legacy: filepath.Join(root, "legacy"), Home: root}
	lib, err := library.Discover(context.Background(), loc, fetch)
	if err != nil {
		t.Fatal(err)
	}
	if len(lib.Games) != 2 || lib.Games[0].Title != "Live Title" || lib.Games[0].Image != filepath.Join(roms, "images/art.png") {
		t.Fatalf("Discovery: %+v", lib)
	}
	if len(library.Eligible(lib, "custom", true, nil)) != 1 || len(library.Eligible(lib, "custom", true, []string{a})) != 0 || len(library.Eligible(lib, "missing", false, nil)) != 0 {
		t.Fatal("Filters did not intersect")
	}
	systems = append(systems, map[string]any{"name": "ports", "gamesystem": true})
	lib, err = library.Discover(context.Background(), loc, fetch)
	if err != nil {
		t.Fatal(err)
	}
	if len(lib.Games) != 1 || lib.Games[0].Path != b {
		t.Fatalf("Duplicate in an unsupported real system was not excluded: %+v", lib)
	}
	write(filepath.Join(roms, "gamelist.xml"), "<bad")
	lib, err = library.Discover(context.Background(), loc, fetch)
	if err != nil || len(lib.Games) != 1 || len(lib.Warnings) == 0 {
		t.Fatalf("Artwork metadata failure must not hide live games: %+v %v", lib, err)
	}
}

func TestUniformGameSampling(t *testing.T) {
	lib := library.Library{}
	for i := 0; i < 100; i++ {
		system := "large"
		if i == 0 {
			system = "small"
		}
		lib.Games = append(lib.Games, library.Game{System: system})
	}
	rng := rand.New(rand.NewPCG(91, 23))
	pool := library.Eligible(lib, "", false, nil)
	small := 0
	for range 20000 {
		pick, err := library.Choose(pool, rng)
		if err != nil {
			t.Fatal(err)
		}
		if pick == 0 {
			small++
		}
	}
	if small < 100 || small > 300 {
		t.Fatalf("Expected title-uniform, not system-uniform distribution: %d", small)
	}
	if value, err := library.Choose([]int{72}, rng); err != nil || value != 72 {
		t.Fatal(value, err)
	}
	if _, err := library.Choose(nil, rng); err == nil {
		t.Fatal("Empty pool accepted")
	}
}
