package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
)

func TestTeamGet(t *testing.T) {
	port := os.Getenv("SERVICE_PORT")
	if port == "" {
		port = "8080"
	}

	baseURL := fmt.Sprintf("http://localhost:%s", port)

	t.Run("Get existing team with members", func(t *testing.T) {
		testTeam := Team{
			Name: "backend-team",
			Members: []TeamMember{
				{Id: "u100", Name: "Alice", IsActive: true},
				{Id: "u101", Name: "Bob", IsActive: true},
				{Id: "u102", Name: "Charlie", IsActive: false},
			},
		}

		teamJSON, err := json.Marshal(testTeam)
		if err != nil {
			t.Log("Failed to marshal team")
			t.FailNow()
		}

		resp, err := http.Post(baseURL+"/team/add", "application/json", bytes.NewBuffer(teamJSON))
		if err != nil {
			t.Logf("Failed to create team: %v", err)
			t.FailNow()
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Logf("Team creation should succeed, got status: %d", resp.StatusCode)
			t.FailNow()
		}

		resp, err = http.Get(baseURL + "/team/get?team_name=backend-team")
		if err != nil {
			t.Logf("Failed to get team: %v", err)
			t.FailNow()
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Logf("Team retrieval should succeed, got status: %d", resp.StatusCode)
			t.FailNow()
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Logf("Failed to read response body: %v", err)
			t.FailNow()
		}

		var retrievedTeam Team
		err = json.Unmarshal(bodyBytes, &retrievedTeam)
		if err != nil {
			t.Logf("Failed to unmarshal team response: %v", err)
			t.FailNow()
		}

		if testTeam.Name != retrievedTeam.Name {
			t.Logf("Team name should match: expected %s, got %s", testTeam.Name, retrievedTeam.Name)
			t.FailNow()
		}

		if len(retrievedTeam.Members) != 3 {
			t.Logf("Should have 3 team members, got %d", len(retrievedTeam.Members))
			t.FailNow()
		}

		expectedMembers := map[string]TeamMember{
			"u100": {Id: "u100", Name: "Alice", IsActive: true},
			"u101": {Id: "u101", Name: "Bob", IsActive: true},
			"u102": {Id: "u102", Name: "Charlie", IsActive: false},
		}

		for _, member := range retrievedTeam.Members {
			expected, exists := expectedMembers[member.Id]
			if !exists {
				t.Logf("Member %s should exist", member.Id)
				t.FailNow()
			}
			if expected != member {
				t.Logf("Member data should match for %s", member.Id)
				t.FailNow()
			}
		}
	})

	t.Run("Get non-existent team", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/team/get?team_name=non-existent-team")
		if err != nil {
			t.Logf("Failed to send request: %v", err)
			t.FailNow()
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Logf("Should return 404 for non-existent team, got %d", resp.StatusCode)
			t.FailNow()
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Logf("Failed to read response body: %v", err)
			t.FailNow()
		}

		var errorResp ErrorResponse
		err = json.Unmarshal(bodyBytes, &errorResp)
		if err != nil {
			t.Logf("Failed to unmarshal error response: %v", err)
			t.FailNow()
		}

		if errorResp.Error.Code != "NOT_FOUND" {
			t.Logf("Error code should be NOT_FOUND, got %s", errorResp.Error.Code)
			t.FailNow()
		}

		contains := false
		for i := 0; i <= len(errorResp.Error.Message)-len("non-existent-team"); i++ {
			if errorResp.Error.Message[i:i+len("non-existent-team")] == "non-existent-team" {
				contains = true
				break
			}
		}
		if !contains {
			t.Logf("Error message should contain team name: %s", errorResp.Error.Message)
			t.FailNow()
		}
	})

	t.Run("Get team without team_name parameter", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/team/get")
		if err != nil {
			t.Logf("Failed to send request: %v", err)
			t.FailNow()
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Logf("Should return 400 for missing team_name, got %d", resp.StatusCode)
			t.FailNow()
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Logf("Failed to read response body: %v", err)
			t.FailNow()
		}

		var errorResp ErrorResponse
		err = json.Unmarshal(bodyBytes, &errorResp)
		if err != nil {
			t.Logf("Failed to unmarshal error response: %v", err)
			t.FailNow()
		}

		if errorResp.Error.Code != "BAD_REQUEST" {
			t.Logf("Error code should be BAD_REQUEST, got %s", errorResp.Error.Code)
			t.FailNow()
		}

		contains := false
		for i := 0; i <= len(errorResp.Error.Message)-len("team_name"); i++ {
			if errorResp.Error.Message[i:i+len("team_name")] == "team_name" {
				contains = true
				break
			}
		}
		if !contains {
			t.Logf("Error message should mention team_name parameter: %s", errorResp.Error.Message)
			t.FailNow()
		}
	})

	t.Run("Get team with empty team_name parameter", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/team/get?team_name=")
		if err != nil {
			t.Logf("Failed to send request: %v", err)
			t.FailNow()
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Logf("Should return 400 for empty team_name, got %d", resp.StatusCode)
			t.FailNow()
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Logf("Failed to read response body: %v", err)
			t.FailNow()
		}

		var errorResp ErrorResponse
		err = json.Unmarshal(bodyBytes, &errorResp)
		if err != nil {
			t.Logf("Failed to unmarshal error response: %v", err)
			t.FailNow()
		}

		if errorResp.Error.Code != "BAD_REQUEST" {
			t.Logf("Error code should be BAD_REQUEST, got %s", errorResp.Error.Code)
			t.FailNow()
		}

		contains := false
		for i := 0; i <= len(errorResp.Error.Message)-len("team_name"); i++ {
			if errorResp.Error.Message[i:i+len("team_name")] == "team_name" {
				contains = true
				break
			}
		}
		if !contains {
			t.Logf("Error message should mention team_name parameter: %s", errorResp.Error.Message)
			t.FailNow()
		}
	})

	t.Run("Get team with special characters in name", func(t *testing.T) {
		specialTeamName := "team-with-dashes_and_underscores"

		testTeam := Team{
			Name: specialTeamName,
			Members: []TeamMember{
				{Id: "u200", Name: "Special User", IsActive: true},
			},
		}

		teamJSON, err := json.Marshal(testTeam)
		if err != nil {
			t.Log("Failed to marshal team")
			t.FailNow()
		}

		resp, err := http.Post(baseURL+"/team/add", "application/json", bytes.NewBuffer(teamJSON))
		if err != nil {
			t.Logf("Failed to create team: %v", err)
			t.FailNow()
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Logf("Team creation should succeed, got status: %d", resp.StatusCode)
			t.FailNow()
		}

		resp, err = http.Get(baseURL + "/team/get?team_name=" + specialTeamName)
		if err != nil {
			t.Logf("Failed to get team: %v", err)
			t.FailNow()
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Logf("Team retrieval should succeed, got status: %d", resp.StatusCode)
			t.FailNow()
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Logf("Failed to read response body: %v", err)
			t.FailNow()
		}

		var retrievedTeam Team
		err = json.Unmarshal(bodyBytes, &retrievedTeam)
		if err != nil {
			t.Logf("Failed to unmarshal team response: %v", err)
			t.FailNow()
		}

		if specialTeamName != retrievedTeam.Name {
			t.Logf("Team name with special characters should match: expected %s, got %s", specialTeamName, retrievedTeam.Name)
			t.FailNow()
		}

		if len(retrievedTeam.Members) != 1 {
			t.Logf("Should have 1 team member, got %d", len(retrievedTeam.Members))
			t.FailNow()
		}

		if retrievedTeam.Members[0].Id != "u200" {
			t.Logf("Member ID should match: expected u200, got %s", retrievedTeam.Members[0].Id)
			t.FailNow()
		}
	})

	t.Run("Get team with inactive members", func(t *testing.T) {
		teamName := "inactive-members-team"

		testTeam := Team{
			Name: teamName,
			Members: []TeamMember{
				{Id: "u300", Name: "Active User", IsActive: true},
				{Id: "u301", Name: "Inactive User", IsActive: false},
			},
		}

		teamJSON, err := json.Marshal(testTeam)
		if err != nil {
			t.Log("Failed to marshal team")
			t.FailNow()
		}

		resp, err := http.Post(baseURL+"/team/add", "application/json", bytes.NewBuffer(teamJSON))
		if err != nil {
			t.Logf("Failed to create team: %v", err)
			t.FailNow()
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Logf("Team creation should succeed, got status: %d", resp.StatusCode)
			t.FailNow()
		}

		resp, err = http.Get(baseURL + "/team/get?team_name=" + teamName)
		if err != nil {
			t.Logf("Failed to get team: %v", err)
			t.FailNow()
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Logf("Team retrieval should succeed, got status: %d", resp.StatusCode)
			t.FailNow()
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Logf("Failed to read response body: %v", err)
			t.FailNow()
		}

		var retrievedTeam Team
		err = json.Unmarshal(bodyBytes, &retrievedTeam)
		if err != nil {
			t.Logf("Failed to unmarshal team response: %v", err)
			t.FailNow()
		}

		if teamName != retrievedTeam.Name {
			t.Logf("Team name should match: expected %s, got %s", teamName, retrievedTeam.Name)
			t.FailNow()
		}

		if len(retrievedTeam.Members) != 2 {
			t.Logf("Should have both active and inactive members, got %d", len(retrievedTeam.Members))
			t.FailNow()
		}

		foundActive := false
		foundInactive := false
		for _, member := range retrievedTeam.Members {
			if member.Id == "u300" {
				foundActive = true
				if !member.IsActive {
					t.Logf("User u300 should be active")
					t.FailNow()
				}
			} else if member.Id == "u301" {
				foundInactive = true
				if member.IsActive {
					t.Logf("User u301 should be inactive")
					t.FailNow()
				}
			}
		}
		if !foundActive {
			t.Log("Should find active user")
			t.FailNow()
		}
		if !foundInactive {
			t.Log("Should find inactive user")
			t.FailNow()
		}
	})
}

func TestTeamGetConcurrent(t *testing.T) {
	port := os.Getenv("SERVICE_PORT")
	if port == "" {
		port = "8080"
	}

	baseURL := fmt.Sprintf("http://localhost:%s", port)

	teamName := "concurrent-team"
	testTeam := Team{
		Name: teamName,
		Members: []TeamMember{
			{Id: "u400", Name: "Concurrent User", IsActive: true},
		},
	}

	teamJSON, err := json.Marshal(testTeam)
	if err != nil {
		t.Log("Failed to marshal team")
		t.FailNow()
	}

	resp, err := http.Post(baseURL+"/team/add", "application/json", bytes.NewBuffer(teamJSON))
	if err != nil {
		t.Logf("Failed to create team: %v", err)
		t.FailNow()
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Logf("Team creation should succeed, got status: %d", resp.StatusCode)
		t.FailNow()
	}

	const concurrentRequests = 10
	results := make(chan error, concurrentRequests)

	for _ = range concurrentRequests {
		go func() {
			resp, err := http.Get(baseURL + "/team/get?team_name=" + teamName)
			if err != nil {
				results <- err
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				results <- fmt.Errorf("expected status 200, got %v", resp.StatusCode)
				return
			}

			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				results <- err
				return
			}

			var retrievedTeam Team
			if err := json.Unmarshal(bodyBytes, &retrievedTeam); err != nil {
				results <- err
				return
			}

			if retrievedTeam.Name != teamName {
				results <- fmt.Errorf("expected team name %s, got %s", teamName, retrievedTeam.Name)
				return
			}

			results <- nil
		}()
	}

	for _ = range concurrentRequests {
		err := <-results
		if err != nil {
			t.Logf("Concurrent request failed: %v", err)
			t.FailNow()
		}
	}
}
