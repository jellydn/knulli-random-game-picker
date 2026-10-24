package picker_test

import (
	"errors"
	"github.com/jellydn/knulli-random-game-picker/internal/library"
	"github.com/jellydn/knulli-random-game-picker/internal/picker"
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
