package scraper

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestHandleResponse(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		body           string
		expectedError  error
		expectedResult string
	}{
		{
			name:           "StatusOK",
			statusCode:     http.StatusOK,
			body:           "OK",
			expectedError:  nil,
			expectedResult: "OK",
		},
		{
			name:           "StatusBadRequest",
			statusCode:     http.StatusBadRequest,
			body:           "nom du fichier rom",
			expectedError:  RomFileNameErr,
			expectedResult: "",
		},
		{
			name:           "StatusUnauthorized",
			statusCode:     http.StatusUnauthorized,
			body:           "Unauthorized",
			expectedError:  ServerOverloadedErr,
			expectedResult: "",
		},
		{
			name:           "StatusForbidden",
			statusCode:     http.StatusForbidden,
			body:           "Forbidden",
			expectedError:  DevLoginErr,
			expectedResult: "",
		},
		{
			name:           "StatusNotFound",
			statusCode:     http.StatusNotFound,
			body:           "Not Found",
			expectedError:  GameNotFoundErr,
			expectedResult: "",
		},
		{
			name:           "StatusLocked",
			statusCode:     http.StatusLocked,
			body:           "Locked",
			expectedError:  ServerLockedErr,
			expectedResult: "",
		},
		{
			name:           "StatusUpgradeRequired",
			statusCode:     http.StatusUpgradeRequired,
			body:           "Upgrade Required",
			expectedError:  AppHasBeenBlockedErr,
			expectedResult: "",
		},
		{
			name:           "StatusTooManyRequests",
			statusCode:     http.StatusTooManyRequests,
			body:           "Too Many Requests",
			expectedError:  TooManyRequestsErr,
			expectedResult: "",
		},
		{
			name:           "StatusRequestHeaderFieldsTooLarge",
			statusCode:     http.StatusRequestHeaderFieldsTooLarge,
			body:           "Request Header Fields Too Large",
			expectedError:  ToowManyUnknownRomsErr,
			expectedResult: "",
		},
		{
			name:           "StatusScrapeQuotaExceeded",
			statusCode:     430,
			body:           "Scrape Quota Exceeded",
			expectedError:  ScrapeQuotaErr,
			expectedResult: "",
		},
		{
			name:           "UnexpectedStatusCode",
			statusCode:     http.StatusInternalServerError,
			body:           "Internal Server Error",
			expectedError:  errors.New("unexpected status code: 500, response: Internal Server Error"),
			expectedResult: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := &http.Response{
				StatusCode: tt.statusCode,
				Body:       io.NopCloser(strings.NewReader(tt.body)),
			}

			result, err := handleResponse(res)
			if tt.expectedError != nil && err == nil {
				t.Errorf("expected error but got none")
			}
			if tt.expectedError == nil && err != nil {
				t.Errorf("did not expect error but got: %v", err)
			}
			if tt.expectedError != nil && err != nil && err.Error() != tt.expectedError.Error() {
				t.Errorf("expected error: %v, got: %v", tt.expectedError, err)
			}
			if string(result) != tt.expectedResult {
				t.Errorf("expected result: %s, got: %s", tt.expectedResult, string(result))
			}
		})
	}
}
