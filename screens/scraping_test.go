package screens

import (
	"context"
	"errors"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/anibaldeboni/screech/scraper"
)

type mockDirEntry struct {
	name  string
	isDir bool
}

func (m mockDirEntry) Name() string               { return m.name }
func (m mockDirEntry) IsDir() bool                { return m.isDir }
func (m mockDirEntry) Type() fs.FileMode          { return 0 }
func (m mockDirEntry) Info() (fs.FileInfo, error) { return nil, nil }

func mockWalkDir(romDir string, maxSubDirs int, paths []string) func(string, fs.WalkDirFunc) error {
	return func(root string, fn fs.WalkDirFunc) error {
		for _, path := range paths {
			depth := len(strings.Split(path, string(filepath.Separator))) - len(strings.Split(romDir, string(filepath.Separator)))
			if depth > maxSubDirs {
				continue
			}
			entry := mockDirEntry{name: filepath.Base(path), isDir: strings.HasSuffix(path, "/")}
			if err := fn(path, entry, nil); err != nil {
				if err == filepath.SkipDir {
					continue
				}
				return err
			}
		}
		return nil
	}
}

func TestFindRoms(t *testing.T) {
	tests := []struct {
		name       string
		romDir     string
		maxSubDirs int
		paths      []string
		expected   []string
	}{
		{
			name:       "Basic test",
			romDir:     "/roms",
			maxSubDirs: 2,
			paths:      []string{"/roms/game1.rom", "/roms/game2.rom", "/roms/subdir/game3.rom"},
			expected:   []string{"game1.rom", "game2.rom", "game3.rom"},
		},
		{
			name:       "Skip hidden directories",
			romDir:     "/roms",
			maxSubDirs: 2,
			paths:      []string{"/roms/.hidden/", "/roms/game2.rom"},
			expected:   []string{"game2.rom"},
		},
		{
			name:       "Skip directories beyond maxSubDirs",
			romDir:     "/roms",
			maxSubDirs: 1,
			paths:      []string{"/roms/game1.rom", "/roms/subdir/game2.rom"},
			expected:   []string{"game1.rom"},
		},
	}

	originalWalkDir := walkDir
	defer func() {
		walkDir = originalWalkDir
	}()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			walkDir = mockWalkDir(tt.romDir, tt.maxSubDirs, tt.paths)

			events := make(chan string, 10)
			roms := findRoms(ctx, events, tt.romDir, tt.maxSubDirs)

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
		})
	}
}

type mockScreenscraper struct {
	findGameFunc      func(ctx context.Context, systemID int, rom string) (*scraper.GameInfoResponse, error)
	downloadMediaFunc func(ctx context.Context, medias []scraper.Media, mediaType scraper.MediaType, dest string) error
}

func (m *mockScreenscraper) FindGame(ctx context.Context, systemID int, rom string) (*scraper.GameInfoResponse, error) {
	return m.findGameFunc(ctx, systemID, rom)
}

func (m *mockScreenscraper) DownloadMedia(ctx context.Context, medias []scraper.Media, mediaType scraper.MediaType, dest string) error {
	return m.downloadMediaFunc(ctx, medias, mediaType, dest)
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
