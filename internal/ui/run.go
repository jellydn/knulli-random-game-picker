// Package ui uses gabagool for the console shell and settings, with a compact
// selector for the picker's four face-button actions and centered cover layout.
package ui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	gaba "github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool/constants"
	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool/i18n"
	"github.com/jellydn/knulli-random-game-picker/internal/firmware"
	"github.com/jellydn/knulli-random-game-picker/internal/input"
	"github.com/jellydn/knulli-random-game-picker/internal/library"
	"github.com/jellydn/knulli-random-game-picker/internal/picker"
	"github.com/veandco/go-sdl2/sdl"
)

type Scanner func(context.Context) (library.Library, error)
type Options struct {
	Profile            firmware.Profile
	Demo, Windowed     bool
	Width, Height      int
	Assets, Screenshot string
}
type scanResult struct {
	library library.Library
	err     error
}

// Run returns a revalidated game only after every SDL/controller resource closes.
// The caller owns submission, outside the UI lifetime.
func Run(ctx context.Context, options Options, state *picker.State, scan Scanner) (*library.Game, error) {
	mapping := input.New(options.Profile.SwapFaceButtons, false)
	gaba.SetInputMappingBytes(mapping.JSON())
	// Explicit JSON owns face buttons; do not let a toolkit environment override swap twice.
	_ = os.Unsetenv("FLIP_FACE_BUTTONS")
	// Share the packaged, licensed font; builds omit unused embedded CFW fonts.
	_ = os.Setenv("FALLBACK_FONT", filepath.Join(options.Assets, "DejaVuSans.ttf"))
	_ = i18n.InitI18NFromBytes(nil)
	initOptions := gaba.Options{WindowTitle: "KNULLI Random Game Picker", PrimaryThemeColorHex: 0x46dcae, LogPath: filepath.Join(options.Profile.StateDir, "ui.log")}
	if options.Demo {
		initOptions.LogPath = filepath.Join(os.TempDir(), "picker-demo-ui.log")
	}
	if options.Windowed {
		_ = os.Setenv("ENVIRONMENT", "DEV")
		_ = os.Setenv("WINDOW_WIDTH", fmt.Sprint(options.Width))
		_ = os.Setenv("WINDOW_HEIGHT", fmt.Sprint(options.Height))
	} else {
		initOptions.WindowOptions.FullscreenDesktop = true
	}
	_ = sdl.SetHint("SDL_GAMECONTROLLERCONFIG_FILE", options.Profile.ControllerDB)
	gaba.Init(initOptions)
	defer gaba.Close()
	if options.Windowed {
		gaba.GetWindow().Window.SetPosition(0, 0)
	}
	canvas, err := newCanvas(options.Assets)
	if err != nil {
		return nil, err
	}
	defer canvas.close()
	var pending <-chan scanResult
	var verifying, launch *library.Game
	dirty := true
	startScan := func() {
		state.Selected = -1
		state.Ready = false
		state.Operation = picker.Scanning
		state.Notice = "Reading EmulationStation's library..."
		ch := make(chan scanResult, 1)
		pending = ch
		go func() { lib, err := scan(ctx); ch <- scanResult{lib, err} }()
	}
	tick := func() {
		if pending == nil {
			return
		}
		select {
		case result := <-pending:
			dirty = true
			pending = nil
			canvas.resetArt()
			if verifying != nil {
				state.Operation = picker.Idle
				err := result.err
				if err == nil {
					err = picker.Revalidate(*verifying, result.library)
				}
				if err != nil {
					state.Selected = -1
					state.Ready = false
					state.Notice = err.Error()
				} else {
					launch = verifying
					state.View = picker.Exit
				}
				verifying = nil
			} else {
				state.Accept(result.library, result.err)
			}
		default:
		}
	}
	if !options.Demo {
		startScan()
	}
	for state.View != picker.Exit {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		tick()
		if state.View == picker.Settings {
			if err := showSettings(state, options.Profile.SwapFaceButtons, tick, startScan, options.Demo); err != nil {
				return nil, err
			}
			canvas.resetArt()
			dirty = true
			continue
		}
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			// Include window expose/resize events, not only semantic input.
			dirty = true
			// gabagool owns opened controllers; SDL's inherited database remains authoritative.
			switch e := event.(type) {
			case *sdl.ControllerDeviceEvent:
				if e.Type == sdl.CONTROLLERDEVICEADDED {
					if controller := sdl.GameControllerOpen(int(e.Which)); controller != nil {
						defer controller.Close()
					}
				}
			}
			action := mapping.Action(event)
			if action == input.Quit {
				state.View = picker.Exit
				launch = nil
				break
			}
			if state.View == picker.Exit || state.Operation == picker.Verifying {
				continue
			}
			if action == input.Back || action == input.Settings {
				state.View = picker.Settings
				state.Notice = ""
				break
			}
			if !state.Ready {
				continue
			}
			if action == input.Reroll {
				_ = state.Roll()
				canvas.resetArt()
			}
			if action == input.Confirm && state.Game() != nil {
				if options.Demo || !options.Profile.CanLaunch {
					state.Notice = "Launch disabled. No request was sent."
					continue
				}
				game := *state.Game()
				verifying = &game
				state.Operation = picker.Verifying
				state.Notice = "Checking launch eligibility..."
				ch := make(chan scanResult, 1)
				pending = ch
				go func() { lib, err := scan(ctx); ch <- scanResult{lib, err} }()
			}
		}
		if dirty {
			canvas.draw(state, options)
			dirty = false
		}
		if options.Screenshot != "" && state.Ready {
			if err := canvas.capture(options.Screenshot); err != nil {
				return nil, err
			}
			options.Screenshot = ""
		}
		time.Sleep(16 * time.Millisecond)
	}
	return launch, nil
}

func showSettings(state *picker.State, swap bool, tick, scan func(), demo bool) error {
	// List's shoulder callbacks let this screen cycle a row without changing
	// focus. Bind semantic Left/Right to those callbacks only in this view.
	gaba.SetInputMappingBytes(input.New(swap, true).JSON())
	gaba.SetFlipFaceButtons(false)
	defer func() { gaba.SetInputMappingBytes(input.New(swap, false).JSON()); gaba.SetFlipFaceButtons(false) }()
	window := gaba.GetWindow()
	_ = window.Renderer.SetLogicalSize(window.GetWidth(), window.GetHeight())
	items := make([]gaba.MenuItem, 4)
	update := func() {
		label := "All Systems"
		for _, system := range state.Systems() {
			if system.ID == state.Filters.System {
				label = system.Label
			}
		}
		onoff := func(value bool) string {
			if value {
				return "ON"
			}
			return "OFF"
		}
		items[0].Text = "System:  " + label
		items[1].Text = "Favorites only:  " + onoff(state.Filters.Favorites)
		items[2].Text = "Exclude last 10 picks:  " + onoff(state.ExcludeRecent)
		items[3].Text = "Reset recent selections"
	}
	update()
	options := gaba.DefaultListOptions("Settings / Filters", items)
	options.SelectedIndex = state.Focus
	options.UseSmallTitle = true
	options.ActionButton = input.Settings.VirtualButton()
	options.TertiaryActionButton = input.Quit.VirtualButton()
	options.MultiSelectConfirmButton = constants.VirtualButtonUnassigned
	options.OnSelect = func(index int, _ *gaba.MenuItem) { state.Focus = index; tick(); update() }
	cycle := func(direction int) func(int) int {
		return func(index int) int {
			tick()
			state.Focus = index
			if state.Ready {
				state.Cycle(direction)
			}
			update()
			return index
		}
	}
	options.OnL1 = cycle(-1)
	options.OnR1 = cycle(1)
	options.FooterHelpItems = []gaba.FooterHelpItem{{ButtonName: input.Confirm.VirtualButton().GetName(), HelpText: "Pick / toggle"}, {ButtonName: input.Settings.VirtualButton().GetName(), HelpText: "Rescan"}, {ButtonName: input.Back.VirtualButton().GetName(), HelpText: "Exit"}}
	result, err := gaba.List(options)
	tick()
	if gaba.IsCancelled(err) {
		state.View = picker.Exit
		return nil
	}
	if err != nil {
		return err
	}
	if result == nil || len(result.Selected) == 0 {
		state.View = picker.Exit
		return nil
	}
	state.Focus = result.Selected[0]
	switch result.Action {
	case gaba.ListActionTertiaryTriggered:
		state.View = picker.Exit
	case gaba.ListActionTriggered:
		if demo {
			state.Notice = "Demo mode: no device scan or launch"
		} else if state.Operation == picker.Idle {
			scan()
		}
	case gaba.ListActionSelected:
		if state.Ready {
			state.ConfirmSetting()
		}
	}
	return nil
}
