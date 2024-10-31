package scraper_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/anibaldeboni/screech/config"
	"github.com/anibaldeboni/screech/scraper"
)

func setupStubServer(t *testing.T, resp string) *httptest.Server {
	return httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/find-games":
				res := scraper.GameInfoResponse{}
				res.Response.Jeu.Medias = []scraper.Media{
					{
						URL:    "http://example.com",
						Type:   "box-3D",
						Region: "br",
					},
				}
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json")

				body, err := json.Marshal(res)
				if err != nil {
					t.Error(err)
				}
				_, _ = w.Write(body)
			case "/get-media":
				file, _ := os.Open("../assets/screenshot.png")
				defer file.Close()

				stat, _ := file.Stat()

				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "image/png")
				http.ServeContent(w, r, "screenshot.png", stat.ModTime(), file)
			case "/errors":
				w.WriteHeader(http.StatusBadRequest)
				w.Header().Set("Content-Type", "text/plain")
				_, _ = w.Write([]byte(resp))
			default:
				t.Errorf("Invalid route: %s", r.URL.Path)
			}
		}))
}

func TestFindGame(t *testing.T) {
	server := setupStubServer(t, "")
	defer server.Close()

	scraper.BaseURL = server.URL + "/find-games"
	res, err := scraper.FindGame(context.Background(), "4", "game")
	if err != nil {
		t.Error(err)
	}

	if res.Response.Jeu.Medias[0].URL != "http://example.com" {
		t.Errorf("Expected http://example.com, got %s", res.Response.Jeu.Medias[0].URL)
	}
}

func TestFindGameCancelContext(t *testing.T) {
	server := setupStubServer(t, "")
	defer server.Close()

	scraper.BaseURL = server.URL + "/find-games"
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := scraper.FindGame(ctx, "4", "game")

	if !errors.Is(err, scraper.HTTPRequestAbortedErr) {
		t.Errorf("Expected HTTP Request Aborted error, got %v", err)
	}
}

func TestFindGameResponseErrors(t *testing.T) {
	tests := []struct {
		name string
		resp string
		err  error
	}{
		{
			name: "HTTP Request Error",
			resp: "API closed",
			err:  scraper.APIClosedErr,
		},
		{
			name: "Game Not Found",
			resp: "Erreur",
			err:  scraper.GameNotFoundErr,
		},
		{
			name: "Empty Body",
			resp: "",
			err:  scraper.EmptyBodyErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := setupStubServer(t, tt.resp)
			defer server.Close()

			scraper.BaseURL = server.URL + "/errors"
			_, err := scraper.FindGame(context.Background(), "4", "game")
			if !errors.Is(err, tt.err) {
				t.Errorf("Expected %v, got %v", tt.err, err)
			}

			if err == nil {
				t.Error("Expected error, got nil")
			}
		})
	}
}

func TestDownloadMedia(t *testing.T) {
	server := setupStubServer(t, "")

	defer server.Close()

	scraper.BaseURL = server.URL + "/get-media"
	config.Media.Regions = []string{"br"}
	err := scraper.DownloadMedia(
		context.Background(),
		[]scraper.Media{
			{
				URL:    server.URL + "/get-media",
				Type:   "box-3D",
				Region: "br",
			},
		}, scraper.Box3D, "screenshot.png")

	os.Remove("screenshot.png")

	if err != nil {
		t.Error(err)
	}
}

func TestDownloadMediaCancelContext(t *testing.T) {
	server := setupStubServer(t, "")

	defer server.Close()

	scraper.BaseURL = server.URL + "/get-media"
	config.Media.Regions = []string{"br"}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := scraper.DownloadMedia(
		ctx,
		[]scraper.Media{
			{
				URL:    server.URL + "/get-media",
				Type:   "box-3D",
				Region: "br",
			},
		}, scraper.Box3D, "screenshot.png")

	if !errors.Is(err, scraper.HTTPRequestAbortedErr) {
		t.Errorf("Expected HTTP Request Aborted error, got %v", err)
	}
}

func TestDownloadMediaInvalidRegion(t *testing.T) {
	server := setupStubServer(t, "")

	defer server.Close()

	scraper.BaseURL = server.URL + "/get-media"
	config.Media.Regions = []string{"ar"}
	err := scraper.DownloadMedia(
		context.Background(),
		[]scraper.Media{
			{
				URL:    server.URL + "/get-media",
				Type:   "box-3D",
				Region: "br",
			},
		}, scraper.Box3D, "screenshot.png")

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if !strings.Contains(err.Error(), "media not found for regions") {
		t.Errorf("Expected media not found for regions error, got %s", err.Error())
	}
}

func TestDownloadMediaIgnoringMissingRegion(t *testing.T) {
	server := setupStubServer(t, "")

	defer server.Close()

	scraper.BaseURL = server.URL + "/get-media"
	config.Media.Regions = []string{"ar"}
	config.Media.IgnoreMissingRegion = true
	err := scraper.DownloadMedia(
		context.Background(),
		[]scraper.Media{
			{
				URL:    server.URL + "/get-media",
				Type:   "box-3D",
				Region: "br",
			},
		}, scraper.Box3D, "screenshot.png")

	os.Remove("screenshot.png")

	if err != nil {
		t.Errorf("Expected no error, got %s", err.Error())
	}
}

func TestDownloadMediaInvalidMediaType(t *testing.T) {
	server := setupStubServer(t, "")

	defer server.Close()

	scraper.BaseURL = server.URL + "/get-media"
	config.Media.Regions = []string{"br"}
	err := scraper.DownloadMedia(
		context.Background(),
		[]scraper.Media{
			{
				URL:    server.URL + "/get-media",
				Type:   "box-3D",
				Region: "br",
			},
		}, scraper.MediaType("invalid-media"), "screenshot.png")

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if !errors.Is(err, scraper.UnknownMediaTypeErr) {
		t.Errorf("Expected Unknown Media Type error, got %v", err)
	}
}
