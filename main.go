package main

import (
	_ "embed"
	"log"
	"os"
	"runtime/debug"

	"github.com/anibaldeboni/screech/config"
	"github.com/anibaldeboni/screech/input"
	"github.com/anibaldeboni/screech/screens"
	"github.com/anibaldeboni/screech/uilib"

	"github.com/veandco/go-sdl2/sdl"
)

//go:embed assets/Roboto-Condensed.ttf
var RobotoCondensed []byte

//go:embed assets/Roboto-BoldCondensed.ttf
var RobotBoldCondensed []byte

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Unhandled error: %v\n", r)
			log.Println("Stack trace:")
			debug.PrintStack()
			os.Exit(-1)
		}
	}()

	config.InitVars()

	if err := uilib.InitSDL(); err != nil {
		panic(err)
	}

	if err := uilib.InitTTF(); err != nil {
		panic(err)
	}

	if err := uilib.InitFont(RobotoCondensed, &config.BodyFont, 30); err != nil {
		panic(err)
	}

	if err := uilib.InitFont(RobotoCondensed, &config.ListFont, 30); err != nil {
		panic(err)
	}

	if err := uilib.InitFont(RobotBoldCondensed, &config.HeaderFont, 38); err != nil {
		panic(err)
	}

	window, err := sdl.CreateWindow("ScreechApp", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, config.ScreenWidth, config.ScreenHeight, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		panic(err)
	}
	defer renderer.Destroy()

	mainScreen, err := screens.NewMainScreen(renderer)
	if err != nil {
		panic(err)
	}

	scrapingScreen, err := screens.NewScrapingScreen(renderer)
	if err != nil {
		panic(err)
	}

	screensMap := map[string]func(){
		"main_screen":     mainScreen.Draw,
		"scraping_screen": scrapingScreen.Draw,
	}

	inputHandlers := map[string]func(input.InputEvent){
		"main_screen":     mainScreen.HandleInput,
		"scraping_screen": scrapingScreen.HandleInput,
	}

	input.StartListening()

	running := true
	for running {

		for {
			event := sdl.PollEvent()
			if event == nil {
				break
			}

			switch event.(type) {
			case *sdl.QuitEvent:
				running = false
			}
		}

		select {
		case inputEvent := <-input.InputChannel:
			if handler, ok := inputHandlers[config.CurrentScreen]; ok {
				handler(inputEvent)
			}
		default:
			// No event received
		}

		if drawFunc, ok := screensMap[config.CurrentScreen]; ok {
			drawFunc()
		}

		sdl.Delay(16)
	}
}
