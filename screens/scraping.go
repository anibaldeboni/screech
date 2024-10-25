package screens

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/anibaldeboni/screech/components"
	"github.com/anibaldeboni/screech/config"
	"github.com/anibaldeboni/screech/input"
	"github.com/anibaldeboni/screech/scraper"
	"github.com/anibaldeboni/screech/uilib"

	"github.com/veandco/go-sdl2/sdl"
)

var (
	scraping        bool
	findGame        = scraper.FindGame
	downloadMedia   = scraper.DownloadMedia
	hasScrapedImage = func(scrapeFile string) bool {
		_, err := os.Stat(scrapeFile)
		return !os.IsNotExist(err)
	}
)

type counter struct {
	success, failed, skipped *atomic.Uint32
}

type DirWalker func(string, fs.WalkDirFunc) error

type ScrapingScreen struct {
	ctx         context.Context
	renderer    *sdl.Renderer
	textView    *components.TextView
	cancel      context.CancelFunc
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
	s.textView = components.NewTextView(
		s.renderer,
		components.TextViewSize{Width: 100, Height: 18},
		sdl.Point{X: 45, Y: 95},
	)
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
		if errors.Is(m.ctx.Err(), context.Canceled) {
			config.CurrentScreen = "main_screen"
			m.initialized = false
		} else {
			m.cancel()
		}
	}
}

func (s *ScrapingScreen) Draw() {
	s.InitScraping()

	_ = s.renderer.SetDrawColor(0, 0, 0, 255) // Background color
	_ = s.renderer.Clear()

	uilib.RenderTexture(s.renderer, config.UiBackground, "Q2", "Q4")
	uilib.RenderTexture(s.renderer, config.UiOverlay, "Q2", "Q4")
	uilib.DrawText(
		s.renderer,
		"Scraping "+config.SystemsNames[config.CurrentSystem],
		sdl.Point{X: 25, Y: 25},
		config.Colors.PRIMARY, config.HeaderFont,
	)

	s.textView.Draw(config.Colors.WHITE)

	uilib.RenderTexture(s.renderer, config.UiControls, "Q3", "Q4")

	s.renderer.Present()
	s.scrape()
}

func isInvalidRom(rom string) bool {
	return slices.Contains(config.ExcludeExtensions, filepath.Ext(rom))
}

func (s *ScrapingScreen) scrape() {
	if scraping {
		return
	} else {
		scraping = true
	}
	s.ctx, s.cancel = context.WithCancel(context.Background())
	events := make(chan string)
	roms := findRoms(s.ctx, events, filepath.Join(config.Roms, config.CurrentSystem), config.MaxScanDepth)

	if roms == nil {
		events <- "Scraping aborted!"
		s.cancel()
		return
	}

	go buildWorkerPool(s.ctx, s.cancel, config.Threads, roms, events)

	go func(ch <-chan string) {
		for msg := range ch {
			s.textView.AddText(msg)
		}
	}(events)
}

func findRelativePath(romDir string, system string) string {
	splittedPath := strings.Split(romDir, string(filepath.Separator))
	systemIndex := slices.Index(splittedPath, system)

	return filepath.Join(splittedPath[systemIndex:]...)
}

func calculateDepth(rootDir, targetDir string) (int, error) {
	relativePath, err := filepath.Rel(rootDir, targetDir)
	if err != nil {
		return -1, err
	}
	if relativePath == "." {
		return 0, nil
	}
	return len(strings.Split(relativePath, string(filepath.Separator))), nil
}

func dirExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

func findRoms(ctx context.Context, events chan<- string, romDir string, maxDepth int) <-chan string {
	roms := make(chan string, 15)

	go func() {
		defer close(roms)
		if exists, err := dirExists(romDir); err != nil {
			events <- fmt.Sprintf("Error checking directory: %v", err)
			return
		} else if !exists {
			events <- fmt.Sprintf("Directory %s does not exist", romDir)
			return
		}
		err := filepath.WalkDir(
			romDir,
			func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					events <- fmt.Sprintf("Error reading directory: %v", err)
					return err
				}

				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					if d.IsDir() {
						depth, err := calculateDepth(romDir, path)
						if err != nil {
							events <- fmt.Sprintf("Error getting relative path: %v", err)
							return nil
						}

						if depth > maxDepth-1 || strings.HasPrefix(filepath.Base(path), ".") {
							return filepath.SkipDir
						}
						events <- "Scanning " + findRelativePath(path, config.CurrentSystem)
					} else {
						roms <- filepath.Base(path)
					}
					return nil
				}
			})
		if err != nil && err != context.Canceled {
			events <- fmt.Sprintf("Error walking the path: %v", err)
		}
	}()

	return roms
}

func buildWorkerPool(ctx context.Context, cancel context.CancelFunc, workers int, roms <-chan string, events chan<- string) {
	var (
		success, failed, skipped atomic.Uint32
		wg                       sync.WaitGroup
	)

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go worker(ctx, &wg, roms, events, &counter{&success, &failed, &skipped})
	}

	go func() {
		wg.Wait()
		defer close(events)
		var completionMsg string
		if errors.Is(ctx.Err(), context.Canceled) {
			completionMsg = "Scraping aborted!"
		} else {
			completionMsg = "Scraping finished."
			cancel()
		}

		events <- completionMsg
		events <- fmt.Sprintf("Success: %d", success.Load())
		events <- fmt.Sprintf("Failed: %d", failed.Load())
		events <- fmt.Sprintf("Skipped: %d", skipped.Load())
	}()
}

func worker(
	ctx context.Context,
	wg *sync.WaitGroup,
	roms <-chan string,
	events chan<- string,
	count *counter,
) {
	defer wg.Done()

download:
	for rom := range roms {
		select {
		case <-ctx.Done():
			break download
		default:
			romName := strings.TrimSuffix(rom, filepath.Ext(rom))

			if isInvalidRom(rom) {
				events <- fmt.Sprintf("Skipping %s: invalid file", rom)
				count.skipped.Add(1)
				continue
			}
			scrapeFile := filepath.Join(config.ScrapedImgDir(), romName+".png")
			if hasScrapedImage(scrapeFile) {
				events <- fmt.Sprintf("Skipping %s: image already scraped", romName)
				count.skipped.Add(1)
				continue
			}

			if res, err := findGame(ctx, config.SystemsIDs[config.CurrentSystem], rom); err != nil {
				if errors.Is(err, scraper.HTTPRequestAbortedErr) {
					break download
				}
				events <- fmt.Sprintf("Error scraping %s: %v", romName, err)
				count.failed.Add(1)
			} else {

				if err := downloadMedia(ctx, res.Response.Jeu.Medias, scraper.MediaType(config.Media.Type), scrapeFile); err != nil {
					events <- fmt.Sprintf("Error scraping %s: %v", romName, err)
					count.failed.Add(1)
					if errors.Is(err, scraper.UnknownMediaTypeErr) {
						break download
					}
				} else {
					events <- "Scrapped " + romName
					count.success.Add(1)
				}
			}
		}
	}
}
