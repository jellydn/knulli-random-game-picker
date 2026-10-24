// Package input maps inherited SDL logical buttons to app actions and gabagool virtual buttons.
package input

import (
	"encoding/json"
	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool/constants"
	"github.com/veandco/go-sdl2/sdl"
)

type Action int

const (
	None Action = iota
	Confirm
	Back
	Reroll
	Settings
	Up
	Down
	Left
	Right
	Quit
)

type Mapping struct {
	Keyboard   map[int]int `json:"keyboard_map"`
	Controller map[int]int `json:"controller_button_map"`
}

func New(swap, settings bool) Mapping {
	c := constants.VirtualButtonA
	b := constants.VirtualButtonB
	x := constants.VirtualButtonX
	y := constants.VirtualButtonY
	if swap {
		c, b = b, c
		x, y = y, x
	}
	left, right := constants.VirtualButtonLeft, constants.VirtualButtonRight
	if settings {
		left, right = constants.VirtualButtonL1, constants.VirtualButtonR1
	}
	return Mapping{
		Keyboard: map[int]int{int(sdl.K_RETURN): int(constants.VirtualButtonA), int(sdl.K_ESCAPE): int(constants.VirtualButtonB), int(sdl.K_x): int(constants.VirtualButtonX), int(sdl.K_y): int(constants.VirtualButtonY),
			int(sdl.K_UP): int(constants.VirtualButtonUp), int(sdl.K_DOWN): int(constants.VirtualButtonDown), int(sdl.K_LEFT): int(left), int(sdl.K_RIGHT): int(right), int(sdl.K_F10): int(constants.VirtualButtonStart)},
		Controller: map[int]int{int(sdl.CONTROLLER_BUTTON_A): int(c), int(sdl.CONTROLLER_BUTTON_B): int(b), int(sdl.CONTROLLER_BUTTON_X): int(x), int(sdl.CONTROLLER_BUTTON_Y): int(y),
			int(sdl.CONTROLLER_BUTTON_DPAD_UP): int(constants.VirtualButtonUp), int(sdl.CONTROLLER_BUTTON_DPAD_DOWN): int(constants.VirtualButtonDown), int(sdl.CONTROLLER_BUTTON_DPAD_LEFT): int(left), int(sdl.CONTROLLER_BUTTON_DPAD_RIGHT): int(right), int(sdl.CONTROLLER_BUTTON_START): int(constants.VirtualButtonStart)},
	}
}
func (m Mapping) JSON() []byte { data, _ := json.Marshal(m); return data }

// VirtualButton is also the source for UI labels, independent of physical layout.
func (a Action) VirtualButton() constants.VirtualButton {
	switch a {
	case Confirm:
		return constants.VirtualButtonA
	case Back:
		return constants.VirtualButtonB
	case Reroll:
		return constants.VirtualButtonX
	case Settings:
		return constants.VirtualButtonY
	case Up:
		return constants.VirtualButtonUp
	case Down:
		return constants.VirtualButtonDown
	case Left:
		return constants.VirtualButtonLeft
	case Right:
		return constants.VirtualButtonRight
	case Quit:
		return constants.VirtualButtonStart
	}
	return constants.VirtualButtonUnassigned
}

func (m Mapping) Action(event sdl.Event) Action {
	var virtual int
	switch e := event.(type) {
	case *sdl.KeyboardEvent:
		if e.Type != sdl.KEYDOWN || e.Repeat != 0 {
			return None
		}
		virtual = m.Keyboard[int(e.Keysym.Sym)]
	case *sdl.ControllerButtonEvent:
		if e.Type != sdl.CONTROLLERBUTTONDOWN {
			return None
		}
		virtual = m.Controller[int(e.Button)]
	case *sdl.QuitEvent:
		return Quit
	}
	for action := Confirm; action <= Quit; action++ {
		if action.VirtualButton() == constants.VirtualButton(virtual) {
			return action
		}
	}
	return None
}
