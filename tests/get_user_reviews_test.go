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

type GetUserReviewsResponse struct {
	UserID       string `json:"user_id"`
	PullRequests []struct {
		PullRequestID   string `json:"pull_request_id"`
		PullRequestName string `json:"pull_request_name"`
		AuthorID        string `json:"author_id"`
		Status          string `json:"status"`
	} `json:"pull_requests"`
}

func TestGetUserReviews(t *testing.T) {
	port := os.Getenv("SERVICE_PORT")
	if port == "" {
		port = "8080"
	}

	baseURL := fmt.Sprintf("http://localhost:%s", port)

	t.Run("Get reviews for user with multiple PRs", func(t *testing.T) {
		// Create a team with multiple users
		testTeam := Team{
			Name: "review-team",
			Members: []TeamMember{
				{Id: "u12000", Name: "Alice", IsActive: true},
				{Id: "u12001", Name: "Bob", IsActive: true},
				{Id: "u12002", Name: "Charlie", IsActive: true},
				{Id: "u12003", Name: "David", IsActive: true},
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

		// Create multiple PRs where Bob is assigned as reviewer
		prsToCreate := []struct {
			prID   string
			prName string
			author string
		}{
			{"pr-review-001", "Feature A", "u12000"},
			{"pr-review-002", "Feature B", "u12002"},
			{"pr-review-003", "Feature C", "u12003"},
		}

		for _, prData := range prsToCreate {
			createPRReq := CreatePullRequestRequest{
				PullRequestID:   prData.prID,
				PullRequestName: prData.prName,
				AuthorID:        prData.author,
			}

			reqJSON, err := json.Marshal(createPRReq)
			if err != nil {
				t.Log("Failed to marshal create PR request")
				t.FailNow()
			}

			resp, err := http.Post(baseURL+"/pullRequest/create", "application/json", bytes.NewBuffer(reqJSON))
			if err != nil {
				t.Logf("Failed to create PR: %v", err)
				t.FailNow()
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusCreated {
				t.Logf("PR creation should succeed, got status: %d", resp.StatusCode)
				t.FailNow()
			}
		}

		// Get reviews for Bob
		resp, err = http.Get(baseURL + "/users/getReview?user_id=u12001")
		if err != nil {
			t.Logf("Failed to get user reviews: %v", err)
			t.FailNow()
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Logf("Get user reviews should succeed, got status: %d", resp.StatusCode)
			t.FailNow()
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Logf("Failed to read response body: %v", err)
			t.FailNow()
		}

		var reviewsResp GetUserReviewsResponse
		err = json.Unmarshal(bodyBytes, &reviewsResp)
		if err != nil {
			t.Logf("Failed to unmarshal reviews response: %v", err)
			t.FailNow()
		}

		// Verify response
		if reviewsResp.UserID != "u12001" {
			t.Logf("User ID should be u12001, got %s", reviewsResp.UserID)
			t.FailNow()
		}

		// Bob should be assigned to all 3 PRs (as one of the 2 reviewers)
		if len(reviewsResp.PullRequests) != 3 {
			t.Logf("Should have 3 PRs assigned to Bob, got %d", len(reviewsResp.PullRequests))
			t.FailNow()
		}

		// Verify PR data
		expectedPRs := map[string]bool{
			"pr-review-001": true,
			"pr-review-002": true,
			"pr-review-003": true,
		}

		for _, pr := range reviewsResp.PullRequests {
			if !expectedPRs[pr.PullRequestID] {
				t.Logf("Unexpected PR ID in response: %s", pr.PullRequestID)
				t.FailNow()
			}
			if pr.Status != "OPEN" {
				t.Logf("PR status should be OPEN, got %s", pr.Status)
				t.FailNow()
			}
		}
	})

	t.Run("Get reviews for user with mixed status PRs", func(t *testing.T) {
		// Create a team
		testTeam := Team{
			Name: "mixed-status-team",
			Members: []TeamMember{
				{Id: "u13000", Name: "Eve", IsActive: true},
				{Id: "u13001", Name: "Frank", IsActive: true},
				{Id: "u13002", Name: "Grace", IsActive: true},
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

		// Create PRs for Frank
		createPRReq := CreatePullRequestRequest{
			PullRequestID:   "pr-mixed-001",
			PullRequestName: "Mixed status PR",
			AuthorID:        "u13000",
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

		// Merge one of the PRs
		mergeReq := MergePullRequestRequest{
			PullRequestID: "pr-mixed-001",
		}

		reqJSON, err = json.Marshal(mergeReq)
		if err != nil {
			t.Log("Failed to marshal merge request")
			t.FailNow()
		}

		resp, err = http.Post(baseURL+"/pullRequest/merge", "application/json", bytes.NewBuffer(reqJSON))
		if err != nil {
			t.Logf("Failed to merge PR: %v", err)
			t.FailNow()
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Logf("Merge should succeed, got status: %d", resp.StatusCode)
			t.FailNow()
		}

		// Get reviews for Frank
		resp, err = http.Get(baseURL + "/users/getReview?user_id=u13001")
		if err != nil {
			t.Logf("Failed to get user reviews: %v", err)
			t.FailNow()
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Logf("Get user reviews should succeed, got status: %d", resp.StatusCode)
			t.FailNow()
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Logf("Failed to read response body: %v", err)
			t.FailNow()
		}

		var reviewsResp GetUserReviewsResponse
		err = json.Unmarshal(bodyBytes, &reviewsResp)
		if err != nil {
			t.Logf("Failed to unmarshal reviews response: %v", err)
			t.FailNow()
		}

		// Frank should have the merged PR in his reviews
		if len(reviewsResp.PullRequests) != 1 {
			t.Logf("Should have 1 PR assigned to Frank, got %d", len(reviewsResp.PullRequests))
			t.FailNow()
		}

		if reviewsResp.PullRequests[0].Status != "MERGED" {
			t.Logf("PR status should be MERGED, got %s", reviewsResp.PullRequests[0].Status)
			t.FailNow()
		}
	})

	t.Run("Get reviews for user with no assigned PRs", func(t *testing.T) {
		// Create a team
		testTeam := Team{
			Name: "no-reviews-team",
			Members: []TeamMember{
				{Id: "u14000", Name: "Henry", IsActive: true},
				{Id: "u14001", Name: "Ivy", IsActive: true},
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

		// Get reviews for Ivy (no PRs created yet)
		resp, err = http.Get(baseURL + "/users/getReview?user_id=u14001")
		if err != nil {
			t.Logf("Failed to get user reviews: %v", err)
			t.FailNow()
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Logf("Get user reviews should succeed, got status: %d", resp.StatusCode)
			t.FailNow()
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Logf("Failed to read response body: %v", err)
			t.FailNow()
		}

		var reviewsResp GetUserReviewsResponse
		err = json.Unmarshal(bodyBytes, &reviewsResp)
		if err != nil {
			t.Logf("Failed to unmarshal reviews response: %v", err)
			t.FailNow()
		}

		if reviewsResp.UserID != "u14001" {
			t.Logf("User ID should be u14001, got %s", reviewsResp.UserID)
			t.FailNow()
		}

		if len(reviewsResp.PullRequests) != 0 {
			t.Logf("Should have 0 PRs assigned to Ivy, got %d", len(reviewsResp.PullRequests))
			t.FailNow()
		}
	})

	t.Run("Get reviews for non-existent user", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/users/getReview?user_id=non-existent-user")
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
	})

	t.Run("Get reviews without user_id parameter", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/users/getReview")
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
}
