package screens

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/anibaldeboni/screech/components"
	"github.com/anibaldeboni/screech/config"
	"github.com/anibaldeboni/screech/input"
	"github.com/anibaldeboni/screech/screenscraper"
	"github.com/anibaldeboni/screech/uilib"

	"github.com/veandco/go-sdl2/sdl"
)

var scraping bool

type ScrapingScreen struct {
	renderer    *sdl.Renderer
	textView    *components.TextView
	initialized bool
}

func NewScrapingScreen(renderer *sdl.Renderer) (*ScrapingScreen, error) {
	return &ScrapingScreen{
		renderer: renderer,
	}, nil
}

func (s *ScrapingScreen) InitScraping() {
	if s.initialized {
		return
	}
	s.textView = components.NewTextView(s.renderer, 18)
	s.initialized = true
}

func SetScraping() {
	scraping = false
}

func (m *ScrapingScreen) HandleInput(event input.InputEvent) {
	switch event.KeyCode {
	case "DOWN":
		m.textView.ScrollDown(1)
	case "UP":
		m.textView.ScrollUp(1)
	case "B":
		config.CurrentScreen = "main_screen"
		m.initialized = false
	}
}

func (s *ScrapingScreen) Draw() {
	s.InitScraping()

	s.renderer.SetDrawColor(0, 0, 0, 255) // Background color
	s.renderer.Clear()

	uilib.RenderTexture(s.renderer, config.UiBackground, "Q2", "Q4")
	uilib.RenderTexture(s.renderer, config.UiOverlay, "Q2", "Q4")
	uilib.DrawText(
		s.renderer,
		"Scraping "+config.CurrentSystem,
		sdl.Point{X: 25, Y: 25},
		config.Colors.PRIMARY, config.HeaderFont,
	)

	s.textView.Draw(config.Colors.WHITE)

	uilib.RenderTexture(s.renderer, config.UiControls, "Q3", "Q4")

	s.renderer.Present()
	s.scrape()
}

func hasScrapedImage(scrapeFile string) bool {
	_, err := os.Stat(scrapeFile)
	if os.IsNotExist(err) {
		return false
	}

	return true
}

func isValidRom(rom string) bool {
	invalidExts := []string{".cue", ".m3u", ".jpg", ".png", ".img", ".sub", ".db", ".xml", ".txt", ".dat"}
	ext := filepath.Ext(rom)
	for _, invalidExt := range invalidExts {
		if strings.EqualFold(ext, invalidExt) {
			return false
		}
	}

	return true
}

func isInvalidRom(rom string) bool {
	return !isValidRom(rom)
}

func download(ch chan string) {
	var success int
	var failed int
	var skipped int

	defer close(ch)
	romDir := filepath.Join(config.Roms, config.CurrentSystem)
	dirEntries, err := os.ReadDir(romDir)
	if err != nil {
		ch <- fmt.Sprintf("Error reading directory: %v", err)
		return
	}

	for _, entry := range dirEntries {
		if entry.IsDir() {
			continue
		}

		rom := entry.Name()

		if isInvalidRom(rom) {
			ch <- fmt.Sprintf("Skipping %s", rom)
			skipped++
			continue
		}
		scrapeFile := filepath.Join(config.ScrapedImgDir(), strings.ReplaceAll(rom, filepath.Ext(rom), ".png"))
		if hasScrapedImage(scrapeFile) {
			ch <- fmt.Sprintf("Skipping %s", rom)
			skipped++
			continue
		}

		ch <- fmt.Sprintf("Scraping %s", rom)
		if res, err := screenscraper.FindGame(config.SystemsIDs[config.CurrentSystem], rom); err != nil {
			ch <- fmt.Sprintf("Error scraping %s: %v", rom, err)
			failed++
		} else {
			if err := screenscraper.DownloadMedia(res.Response.Jeu.Medias, screenscraper.MediaType(config.Media.Type), scrapeFile); err != nil {
				ch <- fmt.Sprintf("Error: %v", err)
				failed++
			} else {
				ch <- fmt.Sprintf("Scrapped %s", filepath.Base(scrapeFile))
				success++
			}
		}
	}

	ch <- fmt.Sprintf("Scraping finished. Success: %d, Failed: %d, Skipped: %d", success, failed, skipped)
}

func (s *ScrapingScreen) scrape() {
	if scraping {
		return
	} else {
		scraping = true
	}
	ch := make(chan string)

	go download(ch)
	go func(ch chan string) {
		for msg := range ch {
			s.textView.AddText(msg)
		}
	}(ch)
}
