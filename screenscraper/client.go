package screenscraper

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"

	"github.com/anibaldeboni/screech/config"
)

var (
	DevID       = "1234"
	DevPassword = "password"
	// url         = "https://www.screenscraper.fr/api2/jeuInfos.php?devid=${u%?}&devpassword=${p#??}&softname=crossmix&output=json&ssid=${userSS}&sspassword=${passSS}&sha1=&systemeid=${ssSystemID}&romtype=rom&romnom=${romNameTrimmed}.zip&romtaille=${rom_size}"
	BaseURL = "https://www.screenscraper.fr/api2/jeuInfos.php"
)

func makeHTTPRequest(systemID string, romPath string) (Response, error) {

	u, err := url.Parse(BaseURL)
	if err != nil {
		return Response{}, err
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
	q.Set("romnom", cleanRomName(filepath.Base(romPath))+".zip")
	q.Set("romtaille", fileSize(romPath))

	u.RawQuery = q.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return Response{}, err
	}
	defer resp.Body.Close()

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return Response{}, err
	}

	return response, nil
}

func SHA1Sum(filePath string) string {
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
func cleanRomName(romName string) string {
	re := regexp.MustCompile(` \([^()]*\)| \[[A-z0-9!+]*\]|\([^()]*\)|\[[A-z0-9!+]*\]`)
	return re.ReplaceAllString(romName, "")
}

func fileSize(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()
	fi, err := file.Stat()
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%d", fi.Size())
}
