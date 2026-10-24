package input_test

import (
	"github.com/jellydn/knulli-random-game-picker/internal/input"
	"github.com/veandco/go-sdl2/sdl"
	"testing"
)

func TestKnulliAndXboxLabels(t *testing.T) {
	for _, swap := range []bool{false, true} {
		mapping := input.New(swap, false)
		buttons := []uint8{0, 1, 2, 3}
		if swap {
			buttons = []uint8{1, 0, 3, 2}
		}
		for i, want := range []input.Action{input.Confirm, input.Back, input.Reroll, input.Settings} {
			event := &sdl.ControllerButtonEvent{Type: sdl.CONTROLLERBUTTONDOWN, Button: buttons[i]}
			if got := mapping.Action(event); got != want {
				t.Fatalf("swap=%v button=%d got=%d want=%d", swap, buttons[i], got, want)
			}
			event.Type = sdl.CONTROLLERBUTTONUP
			if mapping.Action(event) != input.None {
				t.Fatal("Release must not trigger the next view")
			}
		}
		if mapping.Action(&sdl.KeyboardEvent{Type: sdl.KEYDOWN, Keysym: sdl.Keysym{Sym: sdl.K_RETURN}}) != input.Confirm {
			t.Fatal("Enter changed with controller layout")
		}
	}
}
