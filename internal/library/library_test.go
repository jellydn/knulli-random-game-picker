package library_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/jellydn/knulli-random-game-picker/internal/library"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"
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
	alias := filepath.Join(roms, "alias.rom")
	if err := os.Symlink(b, alias); err != nil {
		t.Fatal(err)
	}
	games[0]["path"] = alias
	lib, err = library.Discover(context.Background(), loc, fetch)
	if err != nil || len(lib.Games) != 0 {
		t.Fatalf("Symlink aliases across real systems must be excluded: %+v %v", lib, err)
	}
	if err := os.Remove(alias); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(hidden, alias); err != nil {
		t.Fatal(err)
	}
	lib, err = library.Discover(context.Background(), loc, fetch)
	if err != nil || len(lib.Games) != 1 || lib.Games[0].Path != b {
		t.Fatalf("Rescan reused stale symlink identity: %+v %v", lib, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	canceledFetch := func(ctx context.Context, route string) ([]byte, error) {
		body, err := fetch(ctx, route)
		if route == "/systems/custom/games" {
			cancel()
		}
		return body, err
	}
	if _, err := library.Discover(ctx, loc, canceledFetch); err != context.Canceled {
		t.Fatalf("Canceled scan processed the catalogue: %v", err)
	}
}

func TestDiscoveryExcludesApplicationSystems(t *testing.T) {
	root := t.TempDir()
	ids := []string{"snes", "tools", "Tools", "emulators", "Emulators", "ports"}
	config := "<systemList>"
	for _, id := range ids {
		systemRoot := filepath.Join(root, id)
		if err := os.Mkdir(systemRoot, 0700); err != nil {
			t.Fatal(err)
		}
		if id != "snes" {
			if err := os.WriteFile(filepath.Join(systemRoot, "gamelist.xml"), []byte("<invalid"), 0600); err != nil {
				t.Fatal(err)
			}
		}
		config += "<system><name>" + id + "</name><path>" + systemRoot + "</path></system>"
	}
	config += "</systemList>"
	if err := os.WriteFile(filepath.Join(root, "es_systems.cfg"), []byte(config), 0600); err != nil {
		t.Fatal(err)
	}
	names := []string{"game.sfc", "shared.sfc"}
	for _, id := range ids[1:] {
		names = append(names, id+"-PPSSPP.sh", id+"-ScummVM.sh")
	}
	for _, name := range names {
		if err := os.WriteFile(filepath.Join(root, name), nil, 0600); err != nil {
			t.Fatal(err)
		}
	}
	fetch := func(_ context.Context, route string) ([]byte, error) {
		if route == "/systems" {
			var systems []map[string]any
			for _, id := range ids {
				systems = append(systems, map[string]any{"name": id, "gamesystem": true})
			}
			return json.Marshal(systems)
		}
		if route == "/systems/snes/games" {
			return json.Marshal([]map[string]any{
				{"path": filepath.Join(root, "game.sfc"), "name": "Tools and Emulators", "favorite": true},
				{"path": filepath.Join(root, "shared.sfc"), "name": "Ambiguous game"},
			})
		}
		id := strings.TrimSuffix(strings.TrimPrefix(route, "/systems/"), "/games")
		return json.Marshal([]map[string]any{
			{"path": filepath.Join(root, id+"-PPSSPP.sh"), "favorite": "true"},
			{"path": filepath.Join(root, id+"-ScummVM.sh")},
			{"path": filepath.Join(root, "shared.sfc"), "hidden": true},
		})
	}
	lib, err := library.Discover(context.Background(), library.Locations{User: root}, fetch)
	if err != nil {
		t.Fatal(err)
	}
	if len(lib.Games) != 1 || lib.Games[0].Path != filepath.Join(root, "game.sfc") {
		t.Fatalf("Only the unambiguous real game must survive discovery: %+v", lib)
	}
	if len(lib.Warnings) != 1 || !strings.Contains(lib.Warnings[0], "Ambiguous paths") {
		t.Fatalf("Application gamelists must not be parsed: %v", lib.Warnings)
	}
	for _, id := range ids[1:] {
		if pool := library.Eligible(lib, id, false, nil); len(pool) != 0 {
			t.Fatalf("Application system %s remains eligible: %v", id, pool)
		}
	}
	if pool := library.Eligible(lib, "", true, nil); len(pool) != 1 || lib.Games[pool[0]].Title != "Tools and Emulators" {
		t.Fatalf("Favorites must retain the real game, not application entries: %v", pool)
	}
}

func TestEligibleRejectsCachedApplications(t *testing.T) {
	lib := library.Library{Games: []library.Game{
		{System: "tools", Path: "/relocated/Grout.sh", Favorite: true},
		{System: "emulators", Path: "/relocated/PPSSPP.sh", Favorite: true},
		{Path: "/userdata/roms/tools/RetSend/RetSend.sh", Favorite: true},
		{Path: "/userdata/roms/emulators/ScummVM.sh", Favorite: true},
		{Path: "/userdata/roms/tools/../emulators/PPSSPP.sh", Favorite: true},
		{Path: "/userdata/roms/toolshed/game.sh", Favorite: true},
		{Path: "/userdata/roms/emulators-extra/game.sh", Favorite: true},
		{System: "pc", Path: "/games/game.sh", Favorite: true},
		{System: "sh", Path: "/games/another.sh", Favorite: true},
		{System: "tool", Path: "/games/valid.sh", Favorite: true},
		{System: "ports", Path: "/games/port.sh", Favorite: true},
	}}
	pool := library.Eligible(lib, "", true, nil)
	if len(pool) != 6 {
		t.Fatalf("Expected six real/path-boundary controls, got %v", pool)
	}
	for i, index := range pool {
		if index != i+5 {
			t.Fatalf("Wrong eligibility at %d: %v", i, pool)
		}
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

func TestDiscoveryConfigMerge(t *testing.T) {
	root := t.TempDir()
	user, shared := filepath.Join(root, "user"), filepath.Join(root, "shared")
	for _, dir := range []string{user, shared} {
		if err := os.Mkdir(dir, 0700); err != nil {
			t.Fatal(err)
		}
	}
	rom := filepath.Join(root, "game.rom")
	if err := os.WriteFile(rom, nil, 0600); err != nil {
		t.Fatal(err)
	}
	write := func(path, label, systemPath string) {
		t.Helper()
		xml := "<systemList><system><name>fixture</name><fullname>" + label + "</fullname><path>" + systemPath + "</path></system></systemList>"
		if err := os.WriteFile(path, []byte(xml), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write(filepath.Join(shared, "es_systems.cfg"), "Shared base", root)
	write(filepath.Join(user, "es_systems.cfg"), "User base", root)
	fetch := func(_ context.Context, route string) ([]byte, error) {
		if route == "/systems" {
			return []byte(`[{"name":"fixture","gamesystem":true}]`), nil
		}
		return json.Marshal([]map[string]any{{"path": rom}})
	}
	check := func(want string) {
		t.Helper()
		lib, err := library.Discover(context.Background(), library.Locations{User: user, Shared: shared}, fetch)
		if err != nil || len(lib.Games) != 1 || lib.Games[0].SystemLabel != want || lib.Games[0].Path != rom {
			t.Fatalf("Expected %s with exact live path: %+v, %v", want, lib, err)
		}
	}
	check("User base")
	write(filepath.Join(user, "es_systems_custom.cfg"), "Custom base", root)
	check("Custom base")
	write(filepath.Join(user, "es_systems_z.cfg"), "User additive", "")
	check("User additive")
	// Preserve current merge order, not a claim about every firmware's policy:
	// sorted user additions then sorted shared additions, nonempty fields only.
	write(filepath.Join(shared, "es_systems_a.cfg"), "Shared additive", "")
	check("Shared additive")
	write(filepath.Join(shared, "es_systems_z.cfg"), "Last shared additive", "")
	check("Last shared additive")
}

func TestDiscoveryXMLLimit(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "es_systems.cfg")
	const limit = 32 << 20
	xml := "<systemList></systemList>"
	data := append([]byte(xml), bytes.Repeat([]byte(" "), limit-len(xml))...)
	for _, extra := range []string{"", " "} {
		if err := os.WriteFile(path, append(data, extra...), 0600); err != nil {
			t.Fatal(err)
		}
		fetched := false
		fetch := func(context.Context, string) ([]byte, error) {
			fetched = true
			return []byte("[]"), nil
		}
		_, err := library.Discover(context.Background(), library.Locations{User: root}, fetch)
		if extra == "" && (err != nil || !fetched) {
			t.Fatalf("Exact XML limit rejected: %v", err)
		}
		if extra != "" && (err == nil || fetched) {
			t.Fatalf("Oversized XML reached ES: %v", err)
		}
	}
}

// Host filesystem/JSON cost only: this does not measure ES or handheld latency.
func BenchmarkDiscover(b *testing.B) {
	for _, count := range []int{10000, 25000, 50000} {
		b.Run(fmt.Sprint(count), func(b *testing.B) {
			root := b.TempDir()
			config := "<systemList><system><name>fixture</name><path>" + root + "</path></system></systemList>"
			if err := os.WriteFile(filepath.Join(root, "es_systems.cfg"), []byte(config), 0600); err != nil {
				b.Fatal(err)
			}
			games := make([]map[string]any, count)
			for i := range games {
				path := filepath.Join(root, fmt.Sprintf("game-%d.rom", i))
				if err := os.WriteFile(path, nil, 0600); err != nil {
					b.Fatal(err)
				}
				games[i] = map[string]any{"path": path, "name": "Synthetic game"}
			}
			body, err := json.Marshal(games)
			if err != nil {
				b.Fatal(err)
			}
			fetch := func(_ context.Context, route string) ([]byte, error) {
				if route == "/systems" {
					return []byte(`[{"name":"fixture","gamesystem":true}]`), nil
				}
				return body, nil
			}
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				lib, err := library.Discover(context.Background(), library.Locations{User: root}, fetch)
				if err != nil || len(lib.Games) != count {
					b.Fatalf("got %d games: %v", len(lib.Games), err)
				}
			}
		})
	}
}
