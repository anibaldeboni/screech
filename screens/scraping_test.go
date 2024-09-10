package screens

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/anibaldeboni/screech/scraper"
)

func TestFindRoms(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T) string
		maxDepth  int
		expected  []string
		expectErr bool
	}{
		{
			name: "Basic test",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				_ = os.WriteFile(filepath.Join(dir, "game1.rom"), []byte{}, 0644)
				_ = os.WriteFile(filepath.Join(dir, "game2.rom"), []byte{}, 0644)
				_ = os.Mkdir(filepath.Join(dir, "subdir"), 0755)
				_ = os.WriteFile(filepath.Join(dir, "subdir", "game3.rom"), []byte{}, 0644)
				return dir
			},
			maxDepth: 2,
			expected: []string{"game1.rom", "game2.rom", "game3.rom"},
		},
		{
			name: "Skip hidden directories",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				_ = os.Mkdir(filepath.Join(dir, ".hidden"), 0755)
				_ = os.WriteFile(filepath.Join(dir, ".hidden", "game1.rom"), []byte{}, 0644)
				_ = os.WriteFile(filepath.Join(dir, "game2.rom"), []byte{}, 0644)
				return dir
			},
			maxDepth: 2,
			expected: []string{"game2.rom"},
		},
		{
			name: "Skip directories beyond maxDepth",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				_ = os.WriteFile(filepath.Join(dir, "game1.rom"), []byte{}, 0644)
				_ = os.Mkdir(filepath.Join(dir, "subdir"), 0755)
				_ = os.WriteFile(filepath.Join(dir, "subdir", "game2.rom"), []byte{}, 0644)
				return dir
			},
			maxDepth: 1,
			expected: []string{"game1.rom"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setup(t)
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			events := make(chan string, 10)
			roms := findRoms(ctx, events, dir, tt.maxDepth)

			var result []string
			for rom := range roms {
				result = append(result, rom)
			}

			if len(result) != len(tt.expected) {
				t.Fatalf("expected %d roms, got %d", len(tt.expected), len(result))
			}

			for i, rom := range result {
				if rom != tt.expected[i] {
					t.Errorf("expected rom %s, got %s", tt.expected[i], rom)
				}
			}

			if tt.expectErr {
				select {
				case event := <-events:
					if !strings.Contains(event, "Error reading directory") {
						t.Errorf("expected error event, got %s", event)
					}
				default:
					t.Error("expected error event, but no event received")
				}
			}
		})
	}
}

func TestWorker(t *testing.T) {
	tests := []struct {
		name                string
		roms                []string
		findGameFunc        func(ctx context.Context, systemID string, romPath string) (scraper.GameInfoResponse, error)
		downloadMediaFunc   func(context.Context, []scraper.Media, scraper.MediaType, string) error
		hasScrapedImageFunc func(string) bool
		expectedEvents      []string
		expectedCounts      counter
	}{
		{
			name: "Valid ROM",
			roms: []string{"game1.rom"},
			findGameFunc: func(ctx context.Context, systemID string, rom string) (scraper.GameInfoResponse, error) {
				return scraper.GameInfoResponse{}, nil
			},
			downloadMediaFunc: func(ctx context.Context, medias []scraper.Media, mediaType scraper.MediaType, dest string) error {
				return nil
			},
			hasScrapedImageFunc: func(rom string) bool {
				return false
			},
			expectedEvents: []string{"Scrapped game1"},
			expectedCounts: counter{success: newUint32(1), failed: newUint32(0), skipped: newUint32(0)},
		},
		{
			name: "Invalid ROM",
			roms: []string{"game1.txt"},
			findGameFunc: func(ctx context.Context, systemID string, rom string) (scraper.GameInfoResponse, error) {
				return scraper.GameInfoResponse{}, nil
			},
			downloadMediaFunc: func(ctx context.Context, medias []scraper.Media, mediaType scraper.MediaType, dest string) error {
				return nil
			},
			hasScrapedImageFunc: func(rom string) bool {
				return false
			},
			expectedEvents: []string{"Skipping game1.txt: invalid file"},
			expectedCounts: counter{success: newUint32(0), failed: newUint32(0), skipped: newUint32(1)},
		},
		{
			name: "Already scraped ROM",
			roms: []string{"game1.rom"},
			findGameFunc: func(ctx context.Context, systemID string, rom string) (scraper.GameInfoResponse, error) {
				return scraper.GameInfoResponse{}, nil
			},
			downloadMediaFunc: func(ctx context.Context, medias []scraper.Media, mediaType scraper.MediaType, dest string) error {
				return nil
			},
			hasScrapedImageFunc: func(rom string) bool {
				return true
			},
			expectedEvents: []string{"Skipping game1: image already scraped"},
			expectedCounts: counter{success: newUint32(0), failed: newUint32(0), skipped: newUint32(1)},
		},
		{
			name: "Scraping error",
			roms: []string{"game1.rom"},
			findGameFunc: func(ctx context.Context, systemID string, rom string) (scraper.GameInfoResponse, error) {
				return scraper.GameInfoResponse{}, errors.New("scraping error")
			},
			downloadMediaFunc: func(ctx context.Context, medias []scraper.Media, mediaType scraper.MediaType, dest string) error {
				return nil
			},
			hasScrapedImageFunc: func(rom string) bool {
				return false
			},
			expectedEvents: []string{"Error scraping game1: scraping error"},
			expectedCounts: counter{success: newUint32(0), failed: newUint32(1), skipped: newUint32(0)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			roms := make(chan string, len(tt.roms))
			for _, rom := range tt.roms {
				roms <- rom
			}
			close(roms)

			events := make(chan string, 10)
			var wg sync.WaitGroup
			wg.Add(1)

			count := counter{success: new(atomic.Uint32), failed: new(atomic.Uint32), skipped: new(atomic.Uint32)}

			originalFindGame := findGame
			originalDownloadMedia := downloadMedia
			originalHasScrapedImage := hasScrapedImage
			defer func() {
				findGame = originalFindGame
				downloadMedia = originalDownloadMedia
				hasScrapedImage = originalHasScrapedImage
			}()

			findGame = tt.findGameFunc
			downloadMedia = tt.downloadMediaFunc
			hasScrapedImage = tt.hasScrapedImageFunc

			go worker(ctx, &wg, roms, events, &count)

			wg.Wait()
			close(events)

			var resultEvents []string
			for event := range events {
				resultEvents = append(resultEvents, event)
			}

			if len(resultEvents) != len(tt.expectedEvents) {
				t.Fatalf("expected %d events, got %d", len(tt.expectedEvents), len(resultEvents))
			}

			for i, event := range resultEvents {
				if event != tt.expectedEvents[i] {
					t.Errorf("expected event %s, got %s", tt.expectedEvents[i], event)
				}
			}

			if count.success.Load() != tt.expectedCounts.success.Load() {
				t.Errorf("expected %d success, got %d", tt.expectedCounts.success.Load(), count.success.Load())
			}
			if count.failed.Load() != tt.expectedCounts.failed.Load() {
				t.Errorf("expected %d failed, got %d", tt.expectedCounts.failed.Load(), count.failed.Load())
			}
			if count.skipped.Load() != tt.expectedCounts.skipped.Load() {
				t.Errorf("expected %d skipped, got %d", tt.expectedCounts.skipped.Load(), count.skipped.Load())
			}
		})
	}
}

func TestBuildWorkerPool(t *testing.T) {
	tests := []struct {
		name                string
		roms                []string
		findGameFunc        func(ctx context.Context, systemID string, romPath string) (scraper.GameInfoResponse, error)
		downloadMediaFunc   func(context.Context, []scraper.Media, scraper.MediaType, string) error
		hasScrapedImageFunc func(string) bool
		expectedEvents      []string
		expectedCounts      counter
	}{
		{
			name: "All valid ROMs",
			roms: []string{"game1.rom", "game2.rom"},
			findGameFunc: func(ctx context.Context, systemID string, rom string) (scraper.GameInfoResponse, error) {
				return scraper.GameInfoResponse{}, nil
			},
			downloadMediaFunc: func(ctx context.Context, medias []scraper.Media, mediaType scraper.MediaType, dest string) error {
				return nil
			},
			hasScrapedImageFunc: func(rom string) bool {
				return false
			},
			expectedEvents: []string{"Scrapped game1", "Scrapped game2", "Scraping finished.", "Success: 2", "Failed: 0", "Skipped: 0"},
			expectedCounts: counter{success: newUint32(2), failed: newUint32(0), skipped: newUint32(0)},
		},
		{
			name: "Some invalid ROMs",
			roms: []string{"game1.rom", "game2.txt"},
			findGameFunc: func(ctx context.Context, systemID string, rom string) (scraper.GameInfoResponse, error) {
				return scraper.GameInfoResponse{}, nil
			},
			downloadMediaFunc: func(ctx context.Context, medias []scraper.Media, mediaType scraper.MediaType, dest string) error {
				return nil
			},
			hasScrapedImageFunc: func(rom string) bool {
				return false
			},
			expectedEvents: []string{"Scrapped game1", "Skipping game2.txt: invalid file", "Scraping finished.", "Success: 1", "Failed: 0", "Skipped: 1"},
			expectedCounts: counter{success: newUint32(1), failed: newUint32(0), skipped: newUint32(1)},
		},
		{
			name: "Already scraped ROMs",
			roms: []string{"game1.rom", "game2.rom"},
			findGameFunc: func(ctx context.Context, systemID string, rom string) (scraper.GameInfoResponse, error) {
				return scraper.GameInfoResponse{}, nil
			},
			downloadMediaFunc: func(ctx context.Context, medias []scraper.Media, mediaType scraper.MediaType, dest string) error {
				return nil
			},
			hasScrapedImageFunc: func(rom string) bool {
				return true
			},
			expectedEvents: []string{"Skipping game1: image already scraped", "Skipping game2: image already scraped", "Scraping finished.", "Success: 0", "Failed: 0", "Skipped: 2"},
			expectedCounts: counter{success: newUint32(0), failed: newUint32(0), skipped: newUint32(2)},
		},
		{
			name: "Scraping error",
			roms: []string{"game1.rom"},
			findGameFunc: func(ctx context.Context, systemID string, rom string) (scraper.GameInfoResponse, error) {
				return scraper.GameInfoResponse{}, errors.New("scraping error")
			},
			downloadMediaFunc: func(ctx context.Context, medias []scraper.Media, mediaType scraper.MediaType, dest string) error {
				return nil
			},
			hasScrapedImageFunc: func(rom string) bool {
				return false
			},
			expectedEvents: []string{"Error scraping game1: scraping error", "Scraping finished.", "Success: 0", "Failed: 1", "Skipped: 0"},
			expectedCounts: counter{success: newUint32(0), failed: newUint32(1), skipped: newUint32(0)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			roms := make(chan string, len(tt.roms))
			for _, rom := range tt.roms {
				roms <- rom
			}
			close(roms)

			events := make(chan string, 10)

			originalFindGame := findGame
			originalDownloadMedia := downloadMedia
			originalHasScrapedImage := hasScrapedImage
			defer func() {
				findGame = originalFindGame
				downloadMedia = originalDownloadMedia
				hasScrapedImage = originalHasScrapedImage
			}()

			findGame = tt.findGameFunc
			downloadMedia = tt.downloadMediaFunc
			hasScrapedImage = tt.hasScrapedImageFunc

			go buildWorkerPool(ctx, cancel, 2, roms, events)

			var resultEvents []string
			for event := range events {
				resultEvents = append(resultEvents, event)
			}

			if len(resultEvents) != len(tt.expectedEvents) {
				t.Fatalf("expected %d events, got %d", len(tt.expectedEvents), len(resultEvents))
			}

			for i, event := range resultEvents {
				if event != tt.expectedEvents[i] {
					t.Errorf("expected event %s, got %s", tt.expectedEvents[i], event)
				}
			}
		})
	}
}

func newUint32(val uint32) *atomic.Uint32 {
	v := atomic.Uint32{}
	v.Store(val)
	return &v
}
