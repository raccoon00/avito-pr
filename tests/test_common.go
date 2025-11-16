package tests

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

type Team struct {
	Name    string       `json:"team_name"`
	Members []TeamMember `json:"members"`
}

type TeamMember struct {
	Id       string `json:"user_id"`
	Name     string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type OKResponse struct {
	Team Team `json:"team"`
}

type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type SimpleErrorResponse struct {
	Error   string `json:"code"`
	Message string `json:"message"`
}

func processResponse(t *testing.T, resp *http.Response, err error) (int, *Team, *ErrorResponse) {
	if err != nil {
		t.Logf("Error when sending request: %v\n", err)
		t.FailNow()
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Logf("Error when reading response body: %v\n", err)
		t.FailNow()
	}

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		var team Team
		err = json.Unmarshal(bodyBytes, &team)
		if err != nil {
			t.Logf("Failed to unmarshal team response: %v\nResponse body: %v\n", err, string(bodyBytes))
			t.FailNow()
		}
		return resp.StatusCode, &team, nil
	} else {
		var errorResp ErrorResponse
		err = json.Unmarshal(bodyBytes, &errorResp)
		if err != nil {
			// Try simple error format for backward compatibility
			var simpleErr SimpleErrorResponse
			if simpleErrErr := json.Unmarshal(bodyBytes, &simpleErr); simpleErrErr == nil {
				errorResp.Error.Code = simpleErr.Error
				errorResp.Error.Message = simpleErr.Message
			} else {
				t.Logf("Failed to unmarshal error response: %v\nResponse body: %v\n", err, string(bodyBytes))
				t.FailNow()
			}
		}
		return resp.StatusCode, nil, &errorResp
	}
}

func assertEqual(t *testing.T, expected, actual interface{}, message string) {
	if expected != actual {
		t.Logf("%s: expected %v, got %v", message, expected, actual)
		t.FailNow()
	}
}

func assertTrue(t *testing.T, condition bool, message string) {
	if !condition {
		t.Logf("%s: condition is false", message)
		t.FailNow()
	}
}

func assertLen(t *testing.T, slice interface{}, expectedLen int, message string) {
	switch s := slice.(type) {
	case []TeamMember:
		if len(s) != expectedLen {
			t.Logf("%s: expected length %d, got %d", message, expectedLen, len(s))
			t.FailNow()
		}
	case []interface{}:
		if len(s) != expectedLen {
			t.Logf("%s: expected length %d, got %d", message, expectedLen, len(s))
			t.FailNow()
		}
	default:
		t.Logf("assertLen: unsupported type %T", slice)
		t.FailNow()
	}
}

func assertContains(t *testing.T, str, substr string, message string) {
	contains := false
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			contains = true
			break
		}
	}
	if !contains {
		t.Logf("%s: string '%s' does not contain '%s'", message, str, substr)
		t.FailNow()
	}
}
