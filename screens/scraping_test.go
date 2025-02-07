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

	"github.com/anibaldeboni/screech/config"
	"github.com/anibaldeboni/screech/scraper"
)

func TestFindRoms(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T) romDirSettings
		maxDepth  int
		expected  []string
		expectErr bool
	}{
		{
			name: "Basic test",
			setup: func(t *testing.T) romDirSettings {
				dir := t.TempDir()
				_ = os.WriteFile(filepath.Join(dir, "game1.rom"), []byte{}, 0644)
				_ = os.WriteFile(filepath.Join(dir, "game2.rom"), []byte{}, 0644)
				_ = os.Mkdir(filepath.Join(dir, "subdir"), 0755)
				_ = os.WriteFile(filepath.Join(dir, "subdir", "game3.rom"), []byte{}, 0644)
				return romDirSettings{
					DirName:    "roms",
					Path:       dir,
					OutputDir:  "output",
					SystemName: "system",
				}
			},
			maxDepth: 2,
			expected: []string{"game1.rom", "game2.rom", "game3.rom"},
		},
		{
			name: "Skip hidden directories",
			setup: func(t *testing.T) romDirSettings {
				dir := t.TempDir()
				_ = os.Mkdir(filepath.Join(dir, ".hidden"), 0755)
				_ = os.WriteFile(filepath.Join(dir, ".hidden", "game1.rom"), []byte{}, 0644)
				_ = os.WriteFile(filepath.Join(dir, "game2.rom"), []byte{}, 0644)
				return romDirSettings{
					DirName:    "roms",
					Path:       dir,
					OutputDir:  "output",
					SystemName: "system",
				}
			},
			maxDepth: 2,
			expected: []string{"game2.rom"},
		},
		{
			name: "Skip directories beyond maxDepth",
			setup: func(t *testing.T) romDirSettings {
				dir := t.TempDir()
				_ = os.WriteFile(filepath.Join(dir, "game1.rom"), []byte{}, 0644)
				_ = os.Mkdir(filepath.Join(dir, "subdir"), 0755)
				_ = os.WriteFile(filepath.Join(dir, "subdir", "game2.rom"), []byte{}, 0644)
				return romDirSettings{
					DirName:    "roms",
					Path:       dir,
					OutputDir:  "output",
					SystemName: "system",
				}
			},
			maxDepth: 1,
			expected: []string{"game1.rom"},
		},
		{
			name: "Directory does not exists",
			setup: func(t *testing.T) romDirSettings {
				return romDirSettings{
					DirName:    "roms",
					Path:       "/nonexistent",
					OutputDir:  "output",
					SystemName: "system",
				}
			},
			maxDepth:  2,
			expected:  []string{},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setup(t)
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			events := make(chan string, 10)
			roms := findRoms(ctx, events, []romDirSettings{dir}, tt.maxDepth)

			var result []string
			for rom := range roms {
				result = append(result, rom.Name)
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
					if !strings.Contains(event, "Directory /nonexistent does not exist") {
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
		roms                []Rom
		findGameFunc        func(ctx context.Context, systemID string, romPath string) (scraper.GameInfoResponse, error)
		downloadMediaFunc   func(context.Context, []scraper.Media, scraper.MediaType, string) error
		hasScrapedImageFunc func(string) bool
		expectedEvents      []string
		expectedCounts      counter
	}{
		{
			name: "Valid ROM",
			roms: []Rom{
				{Name: "game1.rom", Path: "game1.rom", OutputDir: "output", SystemID: "1"},
			},
			findGameFunc: func(ctx context.Context, systemID string, rom string) (scraper.GameInfoResponse, error) {
				return scraper.GameInfoResponse{}, nil
			},
			downloadMediaFunc: func(ctx context.Context, medias []scraper.Media, mediaType scraper.MediaType, dest string) error {
				return nil
			},
			hasScrapedImageFunc: func(rom string) bool {
				return false
			},
			expectedEvents: []string{"Scraped game1"},
			expectedCounts: counter{success: newUint32(1), failed: newUint32(0), skipped: newUint32(0)},
		},
		{
			name: "Invalid ROM",
			roms: []Rom{
				{Name: "game1.txt", Path: "game1.txt", OutputDir: "output", SystemID: "1"},
			},
			findGameFunc: func(ctx context.Context, systemID string, rom string) (scraper.GameInfoResponse, error) {
				return scraper.GameInfoResponse{}, nil
			},
			downloadMediaFunc: func(ctx context.Context, medias []scraper.Media, mediaType scraper.MediaType, dest string) error {
				return nil
			},
			hasScrapedImageFunc: func(rom string) bool {
				return false
			},
			expectedEvents: []string{},
			expectedCounts: counter{success: newUint32(0), failed: newUint32(0), skipped: newUint32(0)},
		},
		{
			name: "Already scraped ROM",
			roms: []Rom{
				{Name: "game1.rom", Path: "game1.rom", OutputDir: "output", SystemID: "1"},
			},
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
			roms: []Rom{
				{Name: "game1.rom", Path: "game1.rom", OutputDir: "output", SystemID: "1"},
			},
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

			roms := make(chan Rom, len(tt.roms))
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

			config.ExcludeExtensions = []string{".txt"}

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
		roms                []Rom
		findGameFunc        func(ctx context.Context, systemID string, romPath string) (scraper.GameInfoResponse, error)
		downloadMediaFunc   func(context.Context, []scraper.Media, scraper.MediaType, string) error
		hasScrapedImageFunc func(string) bool
		expectedEvents      []string
		expectedCounts      counter
	}{
		{
			name: "All valid ROMs",
			roms: []Rom{
				{Name: "game1.rom", Path: "game1.rom", OutputDir: "output", SystemID: "1"},
				{Name: "game2.rom", Path: "game2.rom", OutputDir: "output", SystemID: "1"},
			},
			findGameFunc: func(ctx context.Context, systemID string, rom string) (scraper.GameInfoResponse, error) {
				return scraper.GameInfoResponse{}, nil
			},
			downloadMediaFunc: func(ctx context.Context, medias []scraper.Media, mediaType scraper.MediaType, dest string) error {
				return nil
			},
			hasScrapedImageFunc: func(rom string) bool {
				return false
			},
			expectedEvents: []string{"Scraped game1", "Scraped game2", "Scraping finished.", "Success: 2", "Failed: 0", "Skipped: 0"},
			expectedCounts: counter{success: newUint32(2), failed: newUint32(0), skipped: newUint32(0)},
		},
		{
			name: "Some invalid ROMs",
			roms: []Rom{
				{Name: "game1.rom", Path: "game1.rom", OutputDir: "output", SystemID: "1"},
				{Name: "game2.txt", Path: "game2.txt", OutputDir: "output", SystemID: "1"},
			},
			findGameFunc: func(ctx context.Context, systemID string, rom string) (scraper.GameInfoResponse, error) {
				return scraper.GameInfoResponse{}, nil
			},
			downloadMediaFunc: func(ctx context.Context, medias []scraper.Media, mediaType scraper.MediaType, dest string) error {
				return nil
			},
			hasScrapedImageFunc: func(rom string) bool {
				return false
			},
			expectedEvents: []string{"Scraped game1", "Scraping finished.", "Success: 1", "Failed: 0", "Skipped: 0"},
			expectedCounts: counter{success: newUint32(1), failed: newUint32(0), skipped: newUint32(0)},
		},
		{
			name: "Already scraped ROMs",
			roms: []Rom{
				{Name: "game1.rom", Path: "game1.rom", OutputDir: "output", SystemID: "1"},
				{Name: "game2.rom", Path: "game2.rom", OutputDir: "output", SystemID: "1"},
			},
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
			roms: []Rom{
				{Name: "game1.rom", Path: "game1.rom", OutputDir: "output", SystemID: "1"},
			},
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

			roms := make(chan Rom, len(tt.roms))
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

			config.ExcludeExtensions = []string{".txt"}

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
