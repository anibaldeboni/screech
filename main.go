package main

import (
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/debug"

	"github.com/anibaldeboni/screech/config"
	"github.com/anibaldeboni/screech/input"
	"github.com/anibaldeboni/screech/screens"
	"github.com/anibaldeboni/screech/uilib"
	"github.com/anibaldeboni/screech/version"

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

	fmt.Println(version.Short())

	flag.StringVar(&config.ConfigFile, "config", "screech.yaml", "Path to the configuration file")
	flag.Parse()

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

	if err := uilib.InitFont(RobotoCondensed, &config.LongTextFont, 20); err != nil {
		panic(err)
	}

	if err := uilib.InitFont(RobotBoldCondensed, &config.HeaderFont, 38); err != nil {
		panic(err)
	}

	window, err := sdl.CreateWindow("ScreechApp-"+config.Version, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, config.ScreenWidth, config.ScreenHeight, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = window.Destroy()
	}()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = renderer.Destroy()
	}()

	homeScreen, err := screens.NewHomeScreen(renderer)
	if err != nil {
		panic(err)
	}

	scrapingScreen, err := screens.NewScrapingScreen(renderer)
	if err != nil {
		panic(err)
	}

	screensMap := map[string]func(){
		"home_screen":     homeScreen.Draw,
		"scraping_screen": scrapingScreen.Draw,
	}

	inputHandlers := map[string]func(input.UserInputEvent){
		"home_screen":     homeScreen.HandleInput,
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
		case inputEvent := <-input.UserInputChannel:
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
