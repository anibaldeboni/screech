package screenscraper

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/anibaldeboni/screech/config"
)

type MediaType string

var (
	DevID                           = "1234"
	DevPassword                     = "password"
	BaseURL                         = "https://www.screenscraper.fr/api2/jeuInfos.php"
	UnreadableBodyErr               = errors.New("unreadable body")
	EmptyBodyErr                    = errors.New("empty body")
	GameNotFoundErr                 = errors.New("game not found")
	APIClosedErr                    = errors.New("API closed")
	HTTPRequestErr                  = errors.New("error making HTTP request")
	HTTPRequestAbortedErr           = errors.New("request aborted")
	Box2D                 MediaType = "box-2D"
	Box3D                 MediaType = "box-3D"
	MixV1                 MediaType = "mixrbv1"
	MixV2                 MediaType = "mixrbv2"
)

const MAX_FILE_SIZE_BYTES = 104857600 // 100MB

func FindGame(ctx context.Context, systemID string, romPath string) (Response, error) {
	var result Response

	u, err := url.Parse(BaseURL)
	if err != nil {
		return result, err
	}

	q := u.Query()
	q.Set("devid", DevID)
	q.Set("devpassword", DevPassword)
	q.Set("softname", "crossmix")
	q.Set("output", "json")
	q.Set("ssid", config.Username)
	q.Set("sspassword", config.Password)
	q.Set("sha1", SHA1Sum(romPath))
	q.Set("systemeid", systemID)
	q.Set("romtype", "rom")
	q.Set("romnom", cleanRomName(romPath)+".zip")
	q.Set("romtaille", fmt.Sprintf("%d", fileSize(romPath)))

	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return result, HTTPRequestErr
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return result, HTTPRequestAbortedErr
		}
		return result, HTTPRequestErr
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return result, UnreadableBodyErr
	}

	s := string(body)
	switch {
	case strings.Contains(s, "API closed"):
		return result, APIClosedErr
	case strings.Contains(s, "Erreur"):
		return result, GameNotFoundErr
	case s == "":
		return result, EmptyBodyErr
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return result, err
	}

	return result, nil
}

func DownloadMedia(ctx context.Context, medias []Media, mediaType MediaType, dest string) error {
	var mediaURL string

	if _, err := os.Stat(dest); err == nil {
		return fmt.Errorf("destination file already exists: %s", dest)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check if destination file exists: %w", err)
	}

findmedia:
	for _, r := range config.GameRegions {
		for _, media := range medias {
			if media.Type == string(mediaType) && media.Region == r {
				mediaURL = media.URL
				break findmedia
			}
		}
	}

	if mediaURL == "" {
		return errors.New("media not found")
	}

	u, err := url.Parse(mediaURL)
	if err != nil {
		return fmt.Errorf("failed to parse media URL: %w", err)
	}
	q := u.Query()
	q.Set("maxwidth", fmt.Sprintf("%d", config.Thumbnail.Width))
	q.Set("maxheight", fmt.Sprintf("%d", config.Thumbnail.Height))
	u.RawQuery = q.Encode()

	destDir := filepath.Dir(dest)
	if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download media: %w", err)
	}
	defer res.Body.Close()

	file, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := io.Copy(file, res.Body); err != nil {
		return err
	}

	return nil
}

func CRC32Sum(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()

	hash := crc32.NewIEEE()
	if _, err := io.Copy(hash, file); err != nil {
		return ""
	}

	return fmt.Sprintf("%x", hash.Sum(nil))
}

func SHA1Sum(filePath string) string {
	if fileSize(filePath) > MAX_FILE_SIZE_BYTES {
		return ""
	}

	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()

	hash := sha1.New()
	if _, err := io.Copy(hash, file); err != nil {
		return ""
	}

	return fmt.Sprintf("%x", hash.Sum(nil))
}

func cleanRomName(file string) string {
	fileName := filepath.Base(file)

	return cleanSpaces(
		regexp.
			MustCompile(`(\.nkit|!|&|Disc |Rev |-|\s*\([^()]*\)|\s*\[[^\[\]]*\])`).
			ReplaceAllString(
				strings.TrimSuffix(fileName, filepath.Ext(fileName)),
				" ",
			),
	)
}

func cleanSpaces(input string) string {
	return strings.TrimSpace(
		regexp.
			MustCompile(`\s+`).
			ReplaceAllString(input, " "),
	)
}

func fileSize(filePath string) int64 {
	file, err := os.Open(filePath)
	if err != nil {
		return 0
	}
	defer file.Close()
	fi, err := file.Stat()
	if err != nil {
		return 0
	}
	return fi.Size()
}
