package ui

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	gaba "github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"github.com/jellydn/knulli-random-game-picker/internal/input"
	"github.com/jellydn/knulli-random-game-picker/internal/picker"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

var white = sdl.Color{R: 229, G: 239, B: 242, A: 255}
var mint = sdl.Color{R: 70, G: 220, B: 174, A: 255}
var muted = sdl.Color{R: 156, G: 178, B: 186, A: 255}

type canvas struct {
	renderer    *sdl.Renderer
	width       int32
	body, title *ttf.Font
	art         *sdl.Texture
	artPath     string
}

func newCanvas(assets string) (*canvas, error) {
	body, err := ttf.OpenFont(filepath.Join(assets, "DejaVuSans.ttf"), 18)
	if err != nil {
		return nil, err
	}
	title, err := ttf.OpenFont(filepath.Join(assets, "DejaVuSans.ttf"), 26)
	if err != nil {
		body.Close()
		return nil, err
	}
	return &canvas{renderer: gaba.GetWindow().Renderer, body: body, title: title}, nil
}
func (c *canvas) close() {
	c.resetArt()
	c.body.Close()
	c.title.Close()
}
func (c *canvas) resetArt() {
	if c.art != nil {
		c.art.Destroy()
	}
	c.art = nil
	c.artPath = ""
}
func safeArt(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil || stat.Size() > 8<<20 {
		return false
	}
	config, format, err := image.DecodeConfig(f)
	return err == nil && (format == "png" || format == "jpeg") && config.Width > 0 && config.Height > 0 && config.Width <= 2048 && config.Height <= 2048
}
func (c *canvas) box(rect sdl.Rect, color sdl.Color) {
	_ = c.renderer.SetDrawColor(color.R, color.G, color.B, color.A)
	_ = c.renderer.FillRect(&rect)
}
func (c *canvas) text(text string, y, width int32, color sdl.Color, heading bool) {
	font := c.body
	if heading {
		font = c.title
	}
	// A centered one-line label with UTF-8-safe ellipsis, never scaled down.
	runes := []rune(strings.ReplaceAll(text, "\n", " "))
	for len(runes) > 0 {
		w, _, err := font.SizeUTF8(string(runes))
		if err != nil {
			return
		}
		if int32(w) <= width {
			break
		}
		runes = runes[:len(runes)-1]
		if len(runes) > 3 {
			runes = append(runes[:len(runes)-3], '.', '.', '.')
		}
	}
	if len(runes) == 0 {
		return
	}
	surface, err := font.RenderUTF8Blended(string(runes), color)
	if err != nil {
		return
	}
	defer surface.Free()
	texture, err := c.renderer.CreateTextureFromSurface(surface)
	if err != nil {
		return
	}
	defer texture.Destroy()
	_ = c.renderer.Copy(texture, nil, &sdl.Rect{X: (c.width - surface.W) / 2, Y: y, W: surface.W, H: surface.H})
}
func (c *canvas) draw(state *picker.State, options Options) {
	w, h, _ := c.renderer.GetOutputSize()
	c.width = 640
	if h > 0 {
		c.width = max(640, w*480/h)
	}
	_ = c.renderer.SetLogicalSize(c.width, 480)
	c.box(sdl.Rect{W: c.width, H: 480}, sdl.Color{R: 12, G: 22, B: 30, A: 255})
	c.text("RANDOM GAME PICKER", 16, c.width-48, white, true)
	badge := "KNULLI  /  START: QUIT"
	if options.Demo {
		badge = "DEMO  /  NO LAUNCH"
	}
	c.text(badge, 52, c.width-48, mint, false)
	game := state.Game()
	if game != nil {
		if game.Image != c.artPath {
			if c.art != nil {
				c.art.Destroy()
				c.art = nil
			}
			c.artPath = game.Image
			if safeArt(game.Image) {
				c.art, _ = img.LoadTexture(c.renderer, game.Image)
			}
		}
		art := sdl.Rect{X: (c.width - 175) / 2, Y: 90, W: 175, H: 175}
		c.box(art, sdl.Color{R: 22, G: 36, B: 47, A: 255})
		if c.art != nil {
			_, _, w, h, err := c.art.Query()
			if err == nil {
				scale := min(float64(art.W)/float64(w), float64(art.H)/float64(h))
				fw, fh := int32(float64(w)*scale), int32(float64(h)*scale)
				_ = c.renderer.Copy(c.art, nil, &sdl.Rect{X: art.X + (art.W-fw)/2, Y: art.Y + (art.H-fh)/2, W: fw, H: fh})
			}
		} else {
			c.text("?", 136, 170, mint, true)
			c.text("NO ARTWORK", 191, 170, muted, false)
		}
		// Two readable title lines retain more identity than a tiny scaled label.
		words := strings.Fields(game.Title)
		line := ""
		second := ""
		for i, word := range words {
			candidate := strings.TrimSpace(line + " " + word)
			measured, _, _ := c.title.SizeUTF8(candidate)
			if int32(measured) > c.width-48 && line != "" {
				second = strings.Join(words[i:], " ")
				break
			}
			line = candidate
		}
		c.text(line, 275, c.width-48, white, true)
		if second != "" {
			c.text(second, 306, c.width-48, white, true)
		}
		c.text(game.SystemLabel, 340, c.width-48, mint, false)
		label := fmt.Sprintf("%d eligible games", len(state.Pool()))
		if game.Favorite {
			label = "FAVORITE  /  " + label
		}
		c.text(label, 368, c.width-48, muted, false)
	} else {
		title := "NO ELIGIBLE GAMES"
		if !state.Ready {
			title = "LIBRARY UNAVAILABLE"
			if state.Operation == picker.Scanning {
				title = "READING LIBRARY"
			}
		}
		c.text(title, 164, c.width-48, white, true)
		c.text("Open Settings to change filters or rescan.", 219, c.width-48, muted, false)
	}
	c.text(state.Notice, 398, c.width-48, mint, false)
	c.box(sdl.Rect{Y: 427, W: c.width, H: 53}, sdl.Color{R: 22, G: 36, B: 47, A: 255})
	play := "Play"
	color := white
	if game == nil || !options.Profile.CanLaunch || options.Demo {
		play = "No launch"
		color = muted
	} else if state.Operation == picker.Verifying {
		play = "Checking"
		color = muted
	}
	// All selector hints use the same semantic action map as the event loop.
	hints := []struct {
		action input.Action
		text   string
		color  sdl.Color
	}{{input.Confirm, play, color}, {input.Reroll, "Reroll", white}, {input.Settings, "Settings", white}, {input.Back, "Filters", white}}
	for i, hint := range hints {
		if state.Operation == picker.Verifying || (hint.action == input.Reroll && (!state.Ready || len(state.Pool()) == 0)) {
			hint.color = muted
		}
		surface, err := c.body.RenderUTF8Blended(hint.action.VirtualButton().GetName()+" "+hint.text, hint.color)
		if err != nil {
			continue
		}
		texture, err := c.renderer.CreateTextureFromSurface(surface)
		if err == nil {
			cell := c.width / 4
			_ = c.renderer.Copy(texture, nil, &sdl.Rect{X: int32(i)*cell + (cell-surface.W)/2, Y: 444, W: surface.W, H: surface.H})
			texture.Destroy()
		}
		surface.Free()
	}
	gaba.GetWindow().Present()
}
func (c *canvas) capture(path string) error {
	w, h, err := c.renderer.GetOutputSize()
	if err != nil {
		return err
	}
	surface, err := sdl.CreateRGBSurfaceWithFormat(0, w, h, 32, uint32(sdl.PIXELFORMAT_RGBA32))
	if err != nil {
		return err
	}
	defer surface.Free()
	if err = c.renderer.ReadPixels(nil, surface.Format.Format, unsafe.Pointer(&surface.Pixels()[0]), int(surface.Pitch)); err != nil {
		return err
	}
	return img.SavePNG(surface, path)
}
