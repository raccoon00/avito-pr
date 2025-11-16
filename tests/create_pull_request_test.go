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

type CreatePullRequestRequest struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
}

type PullRequestResponse struct {
	PR struct {
		PullRequestID     string   `json:"pull_request_id"`
		PullRequestName   string   `json:"pull_request_name"`
		AuthorID          string   `json:"author_id"`
		Status            string   `json:"status"`
		AssignedReviewers []string `json:"assigned_reviewers"`
		CreatedAt         *string  `json:"createdAt,omitempty"`
		MergedAt          *string  `json:"mergedAt,omitempty"`
	} `json:"pr"`
}

func TestCreatePullRequest(t *testing.T) {
	port := os.Getenv("SERVICE_PORT")
	if port == "" {
		port = "8080"
	}

	baseURL := fmt.Sprintf("http://localhost:%s", port)

	t.Run("Create PR with available reviewers", func(t *testing.T) {
		// First, create a team with multiple active users
		testTeam := Team{
			Name: "dev-team",
			Members: []TeamMember{
				{Id: "u1000", Name: "Alice", IsActive: true},
				{Id: "u1001", Name: "Bob", IsActive: true},
				{Id: "u1002", Name: "Charlie", IsActive: true},
				{Id: "u1003", Name: "David", IsActive: false}, // inactive user
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

		// Create PR by Alice
		createPRReq := CreatePullRequestRequest{
			PullRequestID:   "pr-001",
			PullRequestName: "Add feature X",
			AuthorID:        "u1000",
		}

		reqJSON, err := json.Marshal(createPRReq)
		if err != nil {
			t.Log("Failed to marshal create PR request")
			t.FailNow()
		}

		resp, err = http.Post(baseURL+"/pullRequest/create", "application/json", bytes.NewBuffer(reqJSON))
		if err != nil {
			t.Logf("Failed to create PR: %v", err)
			t.FailNow()
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Logf("PR creation should succeed, got status: %d", resp.StatusCode)
			t.FailNow()
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Logf("Failed to read response body: %v", err)
			t.FailNow()
		}

		var prResp PullRequestResponse
		err = json.Unmarshal(bodyBytes, &prResp)
		if err != nil {
			t.Logf("Failed to unmarshal PR response: %v", err)
			t.FailNow()
		}

		// Verify PR data
		if prResp.PR.PullRequestID != "pr-001" {
			t.Logf("PR ID should be pr-001, got %s", prResp.PR.PullRequestID)
			t.FailNow()
		}

		if prResp.PR.PullRequestName != "Add feature X" {
			t.Logf("PR name should be 'Add feature X', got %s", prResp.PR.PullRequestName)
			t.FailNow()
		}

		if prResp.PR.AuthorID != "u1000" {
			t.Logf("Author ID should be u1000, got %s", prResp.PR.AuthorID)
			t.FailNow()
		}

		if prResp.PR.Status != "OPEN" {
			t.Logf("PR status should be OPEN, got %s", prResp.PR.Status)
			t.FailNow()
		}

		// Should have 2 reviewers (Bob and Charlie, excluding Alice and David)
		if len(prResp.PR.AssignedReviewers) != 2 {
			t.Logf("Should have 2 reviewers, got %d", len(prResp.PR.AssignedReviewers))
			t.FailNow()
		}

		// Verify reviewers are from the same team and not the author
		expectedReviewers := map[string]bool{
			"u1001": true, // Bob
			"u1002": true, // Charlie
		}

		for _, reviewer := range prResp.PR.AssignedReviewers {
			if reviewer == "u1000" {
				t.Log("Author should not be assigned as reviewer")
				t.FailNow()
			}
			if reviewer == "u1003" {
				t.Log("Inactive user should not be assigned as reviewer")
				t.FailNow()
			}
			if !expectedReviewers[reviewer] {
				t.Logf("Unexpected reviewer: %s", reviewer)
				t.FailNow()
			}
		}

		if prResp.PR.CreatedAt == nil {
			t.Log("CreatedAt should be set")
			t.FailNow()
		}

		if prResp.PR.MergedAt != nil {
			t.Log("MergedAt should not be set for new PR")
			t.FailNow()
		}
	})

	t.Run("Create PR with only one available reviewer", func(t *testing.T) {
		// Create a team with only one other active user
		testTeam := Team{
			Name: "small-team",
			Members: []TeamMember{
				{Id: "u2000", Name: "Eve", IsActive: true},
				{Id: "u2001", Name: "Frank", IsActive: true},
				{Id: "u2002", Name: "Grace", IsActive: false}, // inactive
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

		// Create PR by Eve
		createPRReq := CreatePullRequestRequest{
			PullRequestID:   "pr-002",
			PullRequestName: "Fix bug Y",
			AuthorID:        "u2000",
		}

		reqJSON, err := json.Marshal(createPRReq)
		if err != nil {
			t.Log("Failed to marshal create PR request")
			t.FailNow()
		}

		resp, err = http.Post(baseURL+"/pullRequest/create", "application/json", bytes.NewBuffer(reqJSON))
		if err != nil {
			t.Logf("Failed to create PR: %v", err)
			t.FailNow()
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Logf("PR creation should succeed, got status: %d", resp.StatusCode)
			t.FailNow()
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Logf("Failed to read response body: %v", err)
			t.FailNow()
		}

		var prResp PullRequestResponse
		err = json.Unmarshal(bodyBytes, &prResp)
		if err != nil {
			t.Logf("Failed to unmarshal PR response: %v", err)
			t.FailNow()
		}

		// Should have only 1 reviewer (Frank, excluding Eve and Grace)
		if len(prResp.PR.AssignedReviewers) != 1 {
			t.Logf("Should have 1 reviewer, got %d", len(prResp.PR.AssignedReviewers))
			t.FailNow()
		}

		if prResp.PR.AssignedReviewers[0] != "u2001" {
			t.Logf("Reviewer should be u2001, got %s", prResp.PR.AssignedReviewers[0])
			t.FailNow()
		}
	})

	t.Run("Create PR with no available reviewers", func(t *testing.T) {
		// Create a team with only the author
		testTeam := Team{
			Name: "solo-team",
			Members: []TeamMember{
				{Id: "u3000", Name: "Solo", IsActive: true},
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

		// Create PR by Solo
		createPRReq := CreatePullRequestRequest{
			PullRequestID:   "pr-003",
			PullRequestName: "Solo work",
			AuthorID:        "u3000",
		}

		reqJSON, err := json.Marshal(createPRReq)
		if err != nil {
			t.Log("Failed to marshal create PR request")
			t.FailNow()
		}

		resp, err = http.Post(baseURL+"/pullRequest/create", "application/json", bytes.NewBuffer(reqJSON))
		if err != nil {
			t.Logf("Failed to create PR: %v", err)
			t.FailNow()
		}
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Logf("Failed to read response body: %v", err)
			t.FailNow()
		}

		if resp.StatusCode != http.StatusCreated {
			t.Logf("PR creation should succeed, got status: %d", resp.StatusCode)
			t.Logf("Response body:\n%v\n", string(bodyBytes))
			t.FailNow()
		}

		var prResp PullRequestResponse
		err = json.Unmarshal(bodyBytes, &prResp)
		if err != nil {
			t.Logf("Failed to unmarshal PR response: %v", err)
			t.FailNow()
		}

		// Should have 0 reviewers (only author in team)
		if len(prResp.PR.AssignedReviewers) != 0 {
			t.Logf("Should have 0 reviewers, got %d", len(prResp.PR.AssignedReviewers))
			t.FailNow()
		}
	})

	t.Run("Create PR with non-existent author", func(t *testing.T) {
		createPRReq := CreatePullRequestRequest{
			PullRequestID:   "pr-004",
			PullRequestName: "Invalid PR",
			AuthorID:        "non-existent-user",
		}

		reqJSON, err := json.Marshal(createPRReq)
		if err != nil {
			t.Log("Failed to marshal create PR request")
			t.FailNow()
		}

		resp, err := http.Post(baseURL+"/pullRequest/create", "application/json", bytes.NewBuffer(reqJSON))
		if err != nil {
			t.Logf("Failed to send request: %v", err)
			t.FailNow()
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Logf("Should return 404 for non-existent author, got %d", resp.StatusCode)
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
	})

	t.Run("Create duplicate PR", func(t *testing.T) {
		// First create a team
		testTeam := Team{
			Name: "dup-team",
			Members: []TeamMember{
				{Id: "u4000", Name: "User1", IsActive: true},
				{Id: "u4001", Name: "User2", IsActive: true},
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

		// Create first PR
		createPRReq := CreatePullRequestRequest{
			PullRequestID:   "pr-dup",
			PullRequestName: "First PR",
			AuthorID:        "u4000",
		}

		reqJSON, err := json.Marshal(createPRReq)
		if err != nil {
			t.Log("Failed to marshal create PR request")
			t.FailNow()
		}

		resp, err = http.Post(baseURL+"/pullRequest/create", "application/json", bytes.NewBuffer(reqJSON))
		if err != nil {
			t.Logf("Failed to create PR: %v", err)
			t.FailNow()
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Logf("First PR creation should succeed, got status: %d", resp.StatusCode)
			t.FailNow()
		}

		// Try to create duplicate PR
		resp, err = http.Post(baseURL+"/pullRequest/create", "application/json", bytes.NewBuffer(reqJSON))
		if err != nil {
			t.Logf("Failed to send duplicate request: %v", err)
			t.FailNow()
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusConflict {
			t.Logf("Should return 409 for duplicate PR, got %d", resp.StatusCode)
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

		if errorResp.Error.Code != "PR_EXISTS" {
			t.Logf("Error code should be PR_EXISTS, got %s", errorResp.Error.Code)
			t.FailNow()
		}
	})
}
