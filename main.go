package main

import (
	"fmt"
	"os"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

const (
	fontPath        = "./test.ttf"
	fontSize        = 36
	WinWidth  int32 = 1280
	WinHeight int32 = 720
	CenterX   int32 = WinWidth / 2
	CenterY   int32 = WinHeight / 2
)

type Text struct {
	Content string
	X, Y    int32
}

func clamp(value, min, max int32) int32 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func run() (err error) {
	var window *sdl.Window
	var font *ttf.Font
	var surface *sdl.Surface
	var text *sdl.Surface
	var content = "Hello, World!"

	if err = ttf.Init(); err != nil {
		return
	}
	defer ttf.Quit()

	if err = sdl.Init(sdl.INIT_GAMECONTROLLER); err != nil {
		return
	}
	defer sdl.Quit()

	// Create a window for us to draw the text on
	if window, err = sdl.CreateWindow("Drawing text", sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED, WinWidth, WinHeight, sdl.WINDOW_SHOWN); err != nil {
		return
	}

	window.SetAlwaysOnTop(true)
	window.SetKeyboardGrab(true)
	window.Raise()

	defer window.Destroy()

	if font, err = ttf.OpenFont(fontPath, fontSize); err != nil {
		return
	}
	defer font.Close()

	print := func(texts []Text) {
		if surface, err = window.GetSurface(); err != nil {
			return
		}
		surface.FillRect(nil, 0x000000)
		surface.Free()

		for _, t := range texts {
			if text, err = font.RenderUTF8Blended(t.Content, sdl.Color{R: 255, G: 120, B: 75, A: 255}); err != nil {
				return
			}
			defer text.Free()
			if err = text.Blit(nil, surface, &sdl.Rect{X: clamp(t.X-(text.W/2), 0, WinWidth), Y: clamp(t.Y-(text.H/2), 0, WinHeight), W: 0, H: 0}); err != nil {
				return
			}
		}

		// Update the window surface with what we have drawn
		window.UpdateSurface()
	}

	startTime := sdl.GetTicks64()
	// Run infinite loop until user closes the window
	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				running = false
				sdl.Quit()
			case *sdl.KeyboardEvent:
				if e.Type == sdl.KEYUP {
					content = sdl.GetKeyName(e.Keysym.Sym)
				}
			default:
				content = fmt.Sprintf("Unknown event: %T", e)
			}
		}

		elapsedTime := sdl.GetTicks64() - startTime

		print(
			[]Text{
				{content, CenterX, CenterY - 100},
				{fmt.Sprintf("Elapsed time: %d", elapsedTime), CenterX, CenterY},
			},
		)
		if elapsedTime >= 15000 {
			running = false
		}

		sdl.Delay(16)
	}

	return
}

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}
