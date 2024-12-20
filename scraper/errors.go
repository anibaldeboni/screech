package scraper

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"unicode"
)

var (
	ServerLockedErr        = errors.New("server is locked")
	ServerOverloadedErr    = errors.New("The server is overloaded")
	TooManyRequestsErr     = errors.New("too many requests")
	ToowManyUnknownRomsErr = errors.New("Member has scraped too many unrecognized roms (see F.A.Q)")
	AppHasBeenBlockedErr   = errors.New("the app has been blocked")
	GameNotFoundErr        = errors.New("game not found")
	ScrapeQuotaErr         = errors.New("scrape quota has been exceeded for today")
	UnreadableBodyErr      = errors.New("unreadable body")
	RomFileNameErr         = errors.New("rom file name is not correct")
	DevLoginErr            = errors.New("Screech cannot access the server")
	BadRequestErr          = errors.New("bad request")
)

func handleResponse(res *http.Response) ([]byte, error) {
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Join(UnreadableBodyErr, err)
	}

	switch res.StatusCode {
	case http.StatusOK:
		return body, nil
	case http.StatusBadRequest:
		return nil, handleBadRequest(string(body))
	case http.StatusUnauthorized:
		return nil, ServerOverloadedErr
	case http.StatusForbidden:
		return nil, DevLoginErr
	case http.StatusNotFound:
		return nil, GameNotFoundErr
	case http.StatusLocked:
		return nil, ServerLockedErr
	case http.StatusUpgradeRequired:
		return nil, AppHasBeenBlockedErr
	case http.StatusTooManyRequests:
		return nil, TooManyRequestsErr
	case http.StatusRequestHeaderFieldsTooLarge:
		return nil, ToowManyUnknownRomsErr
	case 430:
		return nil, ScrapeQuotaErr

	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, response: %s", res.StatusCode, string(body))
	}

	return body, nil
}

func handleBadRequest(message string) error {
	if strings.Contains(message, "nom du fichier rom") {
		return RomFileNameErr
	}

	if isFieldError(message) {
		field, err := extractMiddleText(message)
		if err != nil {
			return errors.New("unknow field error")
		}
		return fmt.Errorf("field error: %s", field)
	}

	return errors.Join(BadRequestErr, errors.New(message))
}

func normalizeString(s string) string {
	var normalized []rune
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsSpace(r) {
			normalized = append(normalized, unicode.ToLower(r))
		}
	}
	return string(normalized)
}

func extractMiddleText(s string) (string, error) {
	normalized := normalizeString(s)
	re := regexp.MustCompile(`(?i)champ\s+(.*?)\s+errone`)
	matches := re.FindStringSubmatch(normalized)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1]), nil
	}
	return "", errors.New("string does not contain the required text")
}

func isFieldError(s string) bool {
	normalized := normalizeString(s)
	re := regexp.MustCompile(`(?i)champ\s+.*?\s+errone`)
	return re.MatchString(normalized)
}
