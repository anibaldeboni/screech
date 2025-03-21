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
	targetSystems []romDirSettings
)

type ScrapingScreen struct {
	ctx         context.Context
	renderer    *sdl.Renderer
	textView    *components.TextView
	cancel      context.CancelFunc
	initialized bool
}

type Rom struct {
	Name,
	Path,
	OutputDir,
	SystemID string
}

type counter struct {
	success, failed, skipped *atomic.Uint32
}

func NewScrapingScreen(renderer *sdl.Renderer) (*ScrapingScreen, error) {
	return &ScrapingScreen{
		renderer: renderer,
	}, nil
}

func SetScraping() {
	scraping = false
}

func SetTargetSystems(systems []romDirSettings) {
	targetSystems = systems
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

func (s *ScrapingScreen) HandleInput(event input.UserInputEvent) {
	switch event.KeyCode {
	case "DOWN":
		s.textView.ScrollDown(1)
	case "UP":
		s.textView.ScrollUp(1)
	case "B":
		if errors.Is(s.ctx.Err(), context.Canceled) {
			config.CurrentScreen = "home_screen"
			s.initialized = false
		} else {
			s.cancel()
		}
	}
}

func (s *ScrapingScreen) Draw() {
	s.InitScraping()

	_ = s.renderer.SetDrawColor(0, 0, 0, 255) // Background color
	_ = s.renderer.Clear()

	var header string
	if len(targetSystems) > 1 {
		header = "Scraping multiple systems"
	} else {
		header = "Scraping " + targetSystems[0].SystemName
	}

	uilib.RenderTexture(s.renderer, config.UiBackground, "Q2", "Q4")
	uilib.RenderTexture(s.renderer, config.UiOverlay, "Q2", "Q4")
	uilib.DrawText(
		s.renderer,
		header,
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
	roms := findRoms(s.ctx, events, targetSystems, config.MaxScanDepth)

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

func findRoms(ctx context.Context, events chan<- string, romDirs []romDirSettings, maxDepth int) <-chan Rom {
	roms := make(chan Rom, 15)

	go func() {
		defer close(roms)
		for _, romDir := range romDirs {
			if exists, err := dirExists(romDir.Path); err != nil {
				events <- fmt.Sprintf("Error checking directory: %v", err)
				return
			} else if !exists {
				events <- fmt.Sprintf("Directory %s does not exist", romDir.Path)
				return
			}
			err := filepath.WalkDir(
				romDir.Path,
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
							depth, err := calculateDepth(romDir.Path, path)
							if err != nil {
								events <- fmt.Sprintf("Error getting relative path: %v", err)
								return nil
							}

							if depth > maxDepth-1 || strings.HasPrefix(filepath.Base(path), ".") {
								return filepath.SkipDir
							}
							events <- "Walking " + romDir.Path
						} else {
							roms <- Rom{
								Name:      filepath.Base(path),
								Path:      path,
								OutputDir: romDir.OutputDir,
								SystemID:  romDir.SystemID,
							}
						}
						return nil
					}
				})
			if err != nil && !errors.Is(err, context.Canceled) {
				events <- fmt.Sprintf("Error walking the path: %v", err)
			}
		}
	}()

	return roms
}

func buildWorkerPool(ctx context.Context, cancel context.CancelFunc, workers int, roms <-chan Rom, events chan<- string) {
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
	roms <-chan Rom,
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
			if isInvalidRom(rom.Name) {
				continue
			}

			romName := strings.TrimSuffix(rom.Name, filepath.Ext(rom.Name))
			scrapeFile := filepath.Join(config.ScrapedImgDir(rom.OutputDir), romName+".png")

			if hasScrapedImage(scrapeFile) {
				if !config.IgnoreSkippedRomMessage {
					events <- fmt.Sprintf("Skipping %s: image already scraped", romName)
				}
				count.skipped.Add(1)
				continue
			}

			if res, err := findGame(ctx, rom.SystemID, rom.Name); err != nil {
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
					events <- "Scraped " + romName
					count.success.Add(1)
				}
			}
		}
	}
}
