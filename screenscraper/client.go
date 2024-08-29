package screenscraper

import (
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
	DevID                       = "1234"
	DevPassword                 = "password"
	BaseURL                     = "https://www.screenscraper.fr/api2/jeuInfos.php"
	UnreadableBodyErr           = errors.New("unreadable body")
	EmptyBodyErr                = errors.New("empty body")
	GameNotFoundErr             = errors.New("game not found")
	APIClosedErr                = errors.New("API closed")
	HTTPRequestErr              = errors.New("error making HTTP request")
	Box2D             MediaType = "box-2D"
	Box3D             MediaType = "box-3D"
	MixV1             MediaType = "mixrbv1"
	MixV2             MediaType = "mixrbv2"
)

const MAX_FILE_SIZE_BYTES = 104857600 // 100MB

func FindGame(systemID string, romPath string) (Response, error) {
	var res Response

	u, err := url.Parse(BaseURL)
	if err != nil {
		return res, err
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

	resp, err := http.Get(u.String())
	if err != nil {
		return res, HTTPRequestErr
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return res, UnreadableBodyErr
	}

	s := string(body)
	switch {
	case strings.Contains(s, "API closed"):
		return res, APIClosedErr
	case strings.Contains(s, "Erreur"):
		return res, GameNotFoundErr
	case s == "":
		return res, EmptyBodyErr
	}

	if err := json.Unmarshal(body, &res); err != nil {
		return res, err
	}

	return res, nil
}

func DownloadMedia(medias []Media, mediaType MediaType, dest string) error {
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

	destDir := filepath.Dir(dest)
	if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	resp, err := http.Get(mediaURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	file, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
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
