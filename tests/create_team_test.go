package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"testing"
)

func processAddResponse(
	t *testing.T,
	resp *http.Response,
	err error,
) (int, *Team, *ErrorResponse) {
	if err != nil {
		t.Logf("Error when sending post: %v\n", err)
		t.FailNow()
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Logf("Error when reading response body: %v\n", err)
		t.FailNow()
	}

	if resp.StatusCode == http.StatusCreated {
		var respTeam OKResponse
		err = json.Unmarshal(bodyBytes, &respTeam)
		if err != nil {
			t.Logf("Failed to unmarshal response body: %v\n%v\nTrying to unmarshal error...\n", err, bodyBytes)
			t.FailNow()
		}

		return resp.StatusCode, &respTeam.Team, nil
	} else {
		var errorResp ErrorResponse
		err = json.Unmarshal(bodyBytes, &errorResp)
		if err != nil {
			t.Log("Failed to unmarshal response, unknown format")
			t.FailNow()
		}
		return resp.StatusCode, nil, &errorResp
	}
}

func TestTeamAdd(t *testing.T) {
	port := os.Getenv("SERVICE_PORT")

	testTeam := Team{
		Name: "Team A",
		Members: []TeamMember{
			{Id: "u1", Name: "Bob", IsActive: true},
		},
	}

	team, err := json.Marshal(testTeam)
	if err != nil {
		t.Log("Failed to marshal team")
		t.FailNow()
	}

	resp, err := http.Post(fmt.Sprintf("http://localhost:%v/team/add", port), "application/json", bytes.NewBuffer(team))
	sc, respOk, respErr := processAddResponse(t, resp, err)

	if sc == http.StatusCreated && respOk == nil {
		t.Log("Unexpected status and response pair")
		t.FailNow()
	}

	if sc == http.StatusCreated {
		if !reflect.DeepEqual(testTeam, *respOk) {
			t.Log("The returned team is not equal to the requested")
			sentTeam, _ := json.MarshalIndent(testTeam, "", "  ")
			recTeam, _ := json.MarshalIndent(*respOk, "", "  ")
			t.Logf("\nsent:\n%v\n\nreceived:\n%v\n", string(sentTeam), string(recTeam))
			t.FailNow()
		}
	} else {
		recResp, _ := json.MarshalIndent(respErr, "", "  ")
		t.Logf("Expected sc == 201, got %v, respErr:\n%v\n", sc, string(recResp))
		t.FailNow()
	}

	resp, err = http.Post(fmt.Sprintf("http://localhost:%v/team/add", port), "application/json", bytes.NewBuffer(team))
	sc, _, _ = processAddResponse(t, resp, err)

	if sc != http.StatusBadRequest {
		t.Logf("Expected sc to be a Bad Request (400), got %v\n", sc)
	}
}
