package screens

import (
	"hello/components"
	"hello/config"
	"hello/input"
	"hello/uilib"

	"github.com/veandco/go-sdl2/sdl"
)

type ScrapingScreen struct {
	renderer      *sdl.Renderer
	textComponent *components.TextComponent
	initialized   bool
}

func NewScrapingScreen(renderer *sdl.Renderer) (*ScrapingScreen, error) {
	return &ScrapingScreen{
		renderer: renderer,
	}, nil
}

func (s *ScrapingScreen) InitOverview() {
	if s.initialized {
		return
	}

	s.textComponent = components.NewTextComponent(s.renderer, "Scraping "+config.CurrentSystem, config.LongTextFont, 18, 1200)
	s.initialized = true
}

func (m *ScrapingScreen) HandleInput(event input.InputEvent) {
	switch event.KeyCode {
	case "B":
		config.CurrentTester = ""
		config.CurrentScreen = "main_screen"
		m.initialized = false
	}
}

func (s *ScrapingScreen) Draw() {
	s.InitOverview()

	s.renderer.SetDrawColor(0, 0, 0, 255) // Background color
	s.renderer.Clear()

	uilib.RenderTexture(s.renderer, config.UiBackground, "Q2", "Q4")

	uilib.RenderTexture(s.renderer, config.UiOverlay, "Q2", "Q4")

	uilib.DrawText(s.renderer, "Scraper", sdl.Point{X: 25, Y: 25}, config.Colors.PRIMARY, config.HeaderFont)

	s.textComponent = components.NewTextComponent(s.renderer, "Scraping "+config.ScrapedImgDir(), config.LongTextFont, 18, 1200)
	s.textComponent.Draw(config.Colors.WHITE)

	uilib.RenderTexture(s.renderer, config.UiControls, "Q3", "Q4")

	s.renderer.Present()
}
