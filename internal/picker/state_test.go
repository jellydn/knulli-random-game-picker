package picker_test

import (
	"errors"
	"github.com/jellydn/knulli-random-game-picker/internal/library"
	"github.com/jellydn/knulli-random-game-picker/internal/picker"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStartupFiltersFocusAndRecent(t *testing.T) {
	lib := library.Library{Games: []library.Game{{System: "other", Path: "/other", Title: "Not favorite"}, {System: "snes", Path: "/snes", Favorite: true}}}
	saved := picker.Filters{System: "other", Favorites: true}
	state := picker.New(saved)
	state.Notice = "Reading EmulationStation's library..."
	state.Accept(lib, nil)
	if state.Game() != nil || state.Filters != saved {
		t.Fatal("Startup ignored a saved filter")
	}
	if state.Notice != "" {
		t.Fatal("Completed empty scan still reports loading")
	}
	state.View = picker.Settings
	state.Focus = 2
	state.Accept(lib, nil)
	if state.View != picker.Settings || state.Focus != 2 {
		t.Fatal("Discovery stole focus")
	}
	state.Filters.Favorites = false
	if err := state.Roll(); err != nil || state.Game().Path != "/other" {
		t.Fatal("Saved system was not used")
	}
	state.ExcludeRecent = true
	if err := state.Roll(); err == nil {
		t.Fatal("Recent exclusion silently relaxed")
	}
	if state.Game() != nil {
		t.Fatal("Empty pool retained a playable selection")
	}
	state.Focus = 3
	state.Cycle(1)
	if err := state.Roll(); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := picker.Save(dir, state.Filters); err != nil {
		t.Fatal(err)
	}
	if got := picker.Load(dir); got != state.Filters {
		t.Fatal(got)
	}
	state.Accept(library.Library{}, errors.New("offline"))
	if state.Ready || state.Game() != nil {
		t.Fatal("Failed discovery retained playable state")
	}
	state.Accept(lib, nil)
	if !state.Ready {
		t.Fatal("Rescan did not recover")
	}
}
func TestRevalidationRequiresSystemAndExactPath(t *testing.T) {
	game := library.Game{System: "snes", Path: "/games/a 'quoted'; $.rom"}
	for _, candidate := range []library.Game{{System: "other", Path: game.Path}, {System: game.System, Path: "/games/a.rom"}} {
		if picker.Revalidate(game, library.Library{Games: []library.Game{candidate}}) == nil {
			t.Fatal("Accepted a different launch identity")
		}
	}
	if err := picker.Revalidate(game, library.Library{Games: []library.Game{game}}); err != nil {
		t.Fatal(err)
	}
}

func TestLoadInvalidSettingsFallsBackWithoutPartialFilters(t *testing.T) {
	for name, data := range map[string]string{
		"syntax":     `{"system":"snes",`,
		"field type": `{"system":"snes","favorites":"yes"}`,
		"top-level":  `[]`,
		"oversized":  `{"system":"snes"}` + strings.Repeat(" ", 32768),
	} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, "settings.json"), []byte(data), 0600); err != nil {
				t.Fatal(err)
			}
			if got := picker.Load(dir); got != (picker.Filters{}) {
				t.Fatalf("corrupt settings leaked partial filters: %+v", got)
			}
		})
	}
	if got := picker.Load(t.TempDir()); got != (picker.Filters{}) {
		t.Fatalf("missing settings: %+v", got)
	}
	dir := t.TempDir()
	valid := `{"system":"snes","favorites":true}`
	if err := os.WriteFile(filepath.Join(dir, "settings.json"), []byte(valid+strings.Repeat(" ", 32768-len(valid))), 0600); err != nil {
		t.Fatal(err)
	}
	if got := picker.Load(dir); got != (picker.Filters{System: "snes", Favorites: true}) {
		t.Fatalf("settings at exact size limit rejected: %+v", got)
	}
}

func TestSaveFailuresPreserveSettingsAndAllowRecovery(t *testing.T) {
	for _, failure := range []string{"write", "rename"} {
		t.Run(failure, func(t *testing.T) {
			dir := t.TempDir()
			original := picker.Filters{System: "snes", Favorites: true}
			if err := picker.Save(dir, original); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(dir, "settings.json")
			before, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			blocker := filepath.Join(dir, "settings.json.tmp")
			if failure == "rename" {
				// Preserve the old bytes in a directory to deterministically reject rename,
				// even when the tests are run as root.
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
				blocker = path
			}
			if err := os.Mkdir(blocker, 0700); err != nil {
				t.Fatal(err)
			}
			if failure == "rename" {
				if err := os.WriteFile(filepath.Join(path, "preserved"), before, 0600); err != nil {
					t.Fatal(err)
				}
			}
			updated := picker.Filters{System: "日本語 / exact", Favorites: false}
			if err := picker.Save(dir, updated); err == nil {
				t.Fatal("save reported success despite blocked replacement")
			}
			checkPath := path
			if failure == "rename" {
				checkPath = filepath.Join(path, "preserved")
			}
			if after, err := os.ReadFile(checkPath); err != nil || string(after) != string(before) {
				t.Fatalf("failed save altered existing bytes: %q, %v", after, err)
			}
			if err := os.RemoveAll(blocker); err != nil {
				t.Fatal(err)
			}
			if err := picker.Save(dir, updated); err != nil {
				t.Fatal(err)
			}
			if got := picker.Load(dir); got != updated {
				t.Fatalf("save did not recover: %+v", got)
			}
		})
	}
}
