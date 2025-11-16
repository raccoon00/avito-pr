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

type SetUserIsActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

type UserResponse struct {
	User struct {
		UserID   string `json:"user_id"`
		Username string `json:"username"`
		TeamName string `json:"team_name"`
		IsActive bool   `json:"is_active"`
	} `json:"user"`
}

func TestSetUserIsActive(t *testing.T) {
	port := os.Getenv("SERVICE_PORT")
	if port == "" {
		port = "8080"
	}

	baseURL := fmt.Sprintf("http://localhost:%s", port)

	t.Run("Set user active status to false", func(t *testing.T) {
		testTeam := Team{
			Name: "test-team-active",
			Members: []TeamMember{
				{Id: "u500", Name: "Test User", IsActive: true},
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

		setActiveReq := SetUserIsActiveRequest{
			UserID:   "u500",
			IsActive: false,
		}

		reqJSON, err := json.Marshal(setActiveReq)
		if err != nil {
			t.Log("Failed to marshal set active request")
			t.FailNow()
		}

		resp, err = http.Post(baseURL+"/users/setIsActive", "application/json", bytes.NewBuffer(reqJSON))
		if err != nil {
			t.Logf("Failed to set user active status: %v", err)
			t.FailNow()
		}
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Logf("Failed to read response body: %v", err)
			t.FailNow()
		}

		if resp.StatusCode != http.StatusOK {
			t.Logf("Set user active should succeed, got status: %d", resp.StatusCode)
			t.Logf("Response body:\n%v\n", string(bodyBytes))
			t.FailNow()
		}

		var userResp UserResponse
		err = json.Unmarshal(bodyBytes, &userResp)
		if err != nil {
			t.Logf("Failed to unmarshal user response: %v", err)
			t.FailNow()
		}

		if userResp.User.UserID != "u500" {
			t.Logf("User ID should be u500, got %s", userResp.User.UserID)
			t.FailNow()
		}

		if userResp.User.Username != "Test User" {
			t.Logf("Username should be 'Test User', got %s", userResp.User.Username)
			t.FailNow()
		}

		if userResp.User.TeamName != "test-team-active" {
			t.Logf("Team name should be 'test-team-active', got %s", userResp.User.TeamName)
			t.FailNow()
		}

		if userResp.User.IsActive != false {
			t.Logf("User should be inactive, got active: %v", userResp.User.IsActive)
			t.FailNow()
		}

		resp, err = http.Get(baseURL + "/team/get?team_name=test-team-active")
		if err != nil {
			t.Logf("Failed to get team: %v", err)
			t.FailNow()
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Logf("Team retrieval should succeed, got status: %d", resp.StatusCode)
			t.FailNow()
		}

		bodyBytes, err = io.ReadAll(resp.Body)
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

		if len(retrievedTeam.Members) != 1 {
			t.Logf("Should have 1 team member, got %d", len(retrievedTeam.Members))
			t.FailNow()
		}

		if retrievedTeam.Members[0].IsActive != false {
			t.Logf("User should be inactive in team, got active: %v", retrievedTeam.Members[0].IsActive)
			t.FailNow()
		}
	})

	t.Run("Set user active status to true", func(t *testing.T) {
		testTeam := Team{
			Name: "test-team-inactive",
			Members: []TeamMember{
				{Id: "u600", Name: "Inactive User", IsActive: false},
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

		setActiveReq := SetUserIsActiveRequest{
			UserID:   "u600",
			IsActive: true,
		}

		reqJSON, err := json.Marshal(setActiveReq)
		if err != nil {
			t.Log("Failed to marshal set active request")
			t.FailNow()
		}

		resp, err = http.Post(baseURL+"/users/setIsActive", "application/json", bytes.NewBuffer(reqJSON))
		if err != nil {
			t.Logf("Failed to set user active status: %v", err)
			t.FailNow()
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Logf("Set user active should succeed, got status: %d", resp.StatusCode)
			t.FailNow()
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Logf("Failed to read response body: %v", err)
			t.FailNow()
		}

		var userResp UserResponse
		err = json.Unmarshal(bodyBytes, &userResp)
		if err != nil {
			t.Logf("Failed to unmarshal user response: %v", err)
			t.FailNow()
		}

		if userResp.User.IsActive != true {
			t.Logf("User should be active, got active: %v", userResp.User.IsActive)
			t.FailNow()
		}
	})

	t.Run("Set active status for non-existent user", func(t *testing.T) {
		setActiveReq := SetUserIsActiveRequest{
			UserID:   "non-existent-user",
			IsActive: true,
		}

		reqJSON, err := json.Marshal(setActiveReq)
		if err != nil {
			t.Log("Failed to marshal set active request")
			t.FailNow()
		}

		resp, err := http.Post(baseURL+"/users/setIsActive", "application/json", bytes.NewBuffer(reqJSON))
		if err != nil {
			t.Logf("Failed to send request: %v", err)
			t.FailNow()
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Logf("Should return 404 for non-existent user, got %d", resp.StatusCode)
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
		for i := 0; i <= len(errorResp.Error.Message)-len("non-existent-user"); i++ {
			if errorResp.Error.Message[i:i+len("non-existent-user")] == "non-existent-user" {
				contains = true
				break
			}
		}
		if !contains {
			t.Logf("Error message should contain user ID: %s", errorResp.Error.Message)
			t.FailNow()
		}
	})

	t.Run("Set active status with missing user_id", func(t *testing.T) {
		setActiveReq := map[string]interface{}{
			"is_active": true,
		}

		reqJSON, err := json.Marshal(setActiveReq)
		if err != nil {
			t.Log("Failed to marshal set active request")
			t.FailNow()
		}

		resp, err := http.Post(baseURL+"/users/setIsActive", "application/json", bytes.NewBuffer(reqJSON))
		if err != nil {
			t.Logf("Failed to send request: %v", err)
			t.FailNow()
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Logf("Should return 400 for missing user_id, got %d", resp.StatusCode)
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
	})

	t.Run("Set active status with missing is_active", func(t *testing.T) {
		setActiveReq := map[string]interface{}{
			"user_id": "some-user",
		}

		reqJSON, err := json.Marshal(setActiveReq)
		if err != nil {
			t.Log("Failed to marshal set active request")
			t.FailNow()
		}

		resp, err := http.Post(baseURL+"/users/setIsActive", "application/json", bytes.NewBuffer(reqJSON))
		if err != nil {
			t.Logf("Failed to send request: %v", err)
			t.FailNow()
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Logf("Should return 400 for missing is_active, got %d", resp.StatusCode)
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
	})
}
