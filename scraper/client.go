package scraper

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
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
	UnknownMediaTypeErr             = errors.New("unknown media type, choose among box-2D, box-3D, mixrbv1, mixrbv2")
	Box2D                 MediaType = "box-2D"
	Box3D                 MediaType = "box-3D"
	MixV1                 MediaType = "mixrbv1"
	MixV2                 MediaType = "mixrbv2"
)

const maxFileSizeBytes = 104857600 // 100MB

func FindGame(ctx context.Context, systemID string, romPath string) (GameInfoResponse, error) {
	var result GameInfoResponse

	res, err := get(ctx, parseFindGameURL(systemID, romPath))
	if err != nil {
		return result, err
	}

	if err := json.Unmarshal(res, &result); err != nil {
		return result, fmt.Errorf("failed to unmarshal JSON: %w response: %s", err, string(res))
	}

	return result, nil
}

func DownloadMedia(ctx context.Context, medias []Media, mediaType MediaType, dest string) error {
	if err := checkDestination(dest); err != nil {
		return err
	}

	if err := checkMediaType(mediaType); err != nil {
		return err
	}

	mediaURL, err := findMediaURLByRegion(medias, mediaType)
	if err != nil {
		return err
	}

	mediaURL, err = addWHToMediaURL(mediaURL)
	if err != nil {
		return err
	}

	res, err := get(ctx, mediaURL)
	if err != nil {
		return err
	}

	if err := saveToDisk(dest, res); err != nil {
		return err
	}

	return nil
}

func parseFindGameURL(systemID, romPath string) string {
	u, _ := url.Parse(BaseURL)
	q := u.Query()
	q.Set("devid", DevID)
	q.Set("devpassword", DevPassword)
	q.Set("softname", "screech")
	q.Set("output", "json")
	q.Set("ssid", config.Username)
	q.Set("sspassword", config.Password)
	q.Set("sha1", SHA1Sum(romPath))
	q.Set("systemeid", systemID)
	q.Set("romtype", "rom")
	q.Set("romnom", cleanRomName(romPath)+".zip")
	q.Set("romtaille", strconv.FormatInt(fileSize(romPath), 10))
	u.RawQuery = q.Encode()
	return u.String()
}

func filterMediasByType(medias []Media, mediaType MediaType) []Media {
	var filtered []Media
	for _, media := range medias {
		if media.Type == string(mediaType) {
			filtered = append(filtered, media)
		}
	}
	return filtered
}

func findMediaURLByRegion(medias []Media, mediaType MediaType) (mediaURL string, err error) {
	mediasByType := filterMediasByType(medias, mediaType)
	if len(mediasByType) == 0 {
		return mediaURL, fmt.Errorf("media not found for type: %s", mediaType)
	}

findmedia:
	for _, r := range config.Media.Regions {
		for _, media := range mediasByType {
			if media.Region == r {
				mediaURL = media.URL
				break findmedia
			}
		}
	}

	if mediaURL == "" {
		if config.Media.IgnoreMissingRegion {
			mediaURL = mediasByType[0].URL
		} else {
			return mediaURL, fmt.Errorf("media not found for regions: %s", config.Media.Regions)
		}
	}

	return mediaURL, nil
}

func addWHToMediaURL(mediaURL string) (string, error) {
	u, err := url.Parse(mediaURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse media URL: %w", err)
	}
	q := u.Query()
	q.Set("maxwidth", strconv.Itoa(config.Boxart.Width))
	q.Set("maxheight", strconv.Itoa(config.Boxart.Height))
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func checkMediaType(mediaType MediaType) error {
	switch mediaType {
	case Box2D, Box3D, MixV1, MixV2:
		return nil
	default:
		return UnknownMediaTypeErr
	}
}

func checkDestination(dest string) error {
	if _, err := os.Stat(dest); err == nil {
		return fmt.Errorf("destination file already exists: %s", dest)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check if destination file exists: %w", err)
	}

	return nil
}

func get(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil, HTTPRequestAbortedErr
		}
		return nil, HTTPRequestErr
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Join(UnreadableBodyErr, fmt.Errorf("error: %w", err))
	}

	s := string(body)
	switch {
	case strings.Contains(s, "API closed"):
		return nil, APIClosedErr
	case strings.Contains(s, "Erreur"):
		return nil, errors.Join(GameNotFoundErr, fmt.Errorf(": %s", s))
	case s == "":
		return nil, EmptyBodyErr
	}

	return body, nil
}

func saveToDisk(dest string, file []byte) error {
	if err := os.MkdirAll(filepath.Dir(dest), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	if err := os.WriteFile(dest, file, os.ModePerm); err != nil {
		return fmt.Errorf("failed to write file to disk: %w", err)
	}

	return nil
}

func SHA1Sum(filePath string) string {
	if fileSize(filePath) > maxFileSizeBytes {
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

	return hex.EncodeToString(hash.Sum(nil))
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
