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

type Team struct {
	Name    string       `json:"team_name"`
	Members []TeamMember `json:"members"`
}

type TeamMember struct {
	Id     string `json:"user_id"`
	Name   string `json:"username"`
	Active bool   `json:"is_active"`
}

type OKResponse struct {
	Team Team `json:"team"`
}

type ErrorResponse struct {
	Error   string `json:"code"`
	Message string `json:"message"`
}

func processAddResponse(
	t *testing.T,
	resp *http.Response,
	err error,
) (int, *OKResponse, *ErrorResponse) {
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

		return resp.StatusCode, &respTeam, nil
	} else {
		var respErr ErrorResponse
		err = json.Unmarshal(bodyBytes, &respErr)
		if err != nil {
			t.Log("Failed to unmarshal response, unknown format")
			t.FailNow()
		}
		return resp.StatusCode, nil, &respErr
	}
}

func TestTeamAdd(t *testing.T) {
	port := os.Getenv("SERVICE_PORT")

	testTeam := Team{
		Name: "Team A",
		Members: []TeamMember{
			{Id: "u1", Name: "Bob", Active: true},
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
		if !reflect.DeepEqual(testTeam, respOk.Team) {
			t.Log("The returned team is not equal to the requested")
			sentTeam, _ := json.MarshalIndent(testTeam, "", "  ")
			recTeam, _ := json.MarshalIndent(respOk.Team, "", "  ")
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
