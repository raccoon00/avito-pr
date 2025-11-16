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

type ReassignReviewerRequest struct {
	PullRequestID string `json:"pull_request_id"`
	OldUserID     string `json:"old_user_id"`
}

type ReassignReviewerResponse struct {
	PR struct {
		PullRequestID     string   `json:"pull_request_id"`
		PullRequestName   string   `json:"pull_request_name"`
		AuthorID          string   `json:"author_id"`
		Status            string   `json:"status"`
		AssignedReviewers []string `json:"assigned_reviewers"`
		CreatedAt         *string  `json:"createdAt,omitempty"`
		MergedAt          *string  `json:"mergedAt,omitempty"`
	} `json:"pr"`
	ReplacedBy string `json:"replaced_by"`
}

func TestReassignReviewer(t *testing.T) {
	port := os.Getenv("SERVICE_PORT")
	if port == "" {
		port = "8080"
	}

	baseURL := fmt.Sprintf("http://localhost:%s", port)

	t.Run("Successfully reassign reviewer", func(t *testing.T) {
		// Create a team with multiple active users
		testTeam := Team{
			Name: "reassign-team",
			Members: []TeamMember{
				{Id: "u5000", Name: "Alice", IsActive: true},
				{Id: "u5001", Name: "Bob", IsActive: true},
				{Id: "u5002", Name: "Charlie", IsActive: true},
				{Id: "u5003", Name: "David", IsActive: true},
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
			PullRequestID:   "pr-reassign-001",
			PullRequestName: "Reassign test PR",
			AuthorID:        "u5000",
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

		// Verify initial reviewers (should be Bob and Charlie)
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

		if len(prResp.PR.AssignedReviewers) != 2 {
			t.Logf("Should have 2 reviewers, got %d", len(prResp.PR.AssignedReviewers))
			t.FailNow()
		}

		// Reassign Bob to David
		reassignReq := ReassignReviewerRequest{
			PullRequestID: "pr-reassign-001",
			OldUserID:     prResp.PR.AssignedReviewers[0],
		}

		reqJSON, err = json.Marshal(reassignReq)
		if err != nil {
			t.Log("Failed to marshal reassign request")
			t.FailNow()
		}

		resp, err = http.Post(baseURL+"/pullRequest/reassign", "application/json", bytes.NewBuffer(reqJSON))
		if err != nil {
			t.Logf("Failed to reassign reviewer: %v", err)
			t.FailNow()
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Logf("Reassign should succeed, got status: %d", resp.StatusCode)
			t.FailNow()
		}

		bodyBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			t.Logf("Failed to read response body: %v", err)
			t.FailNow()
		}

		var reassignResp ReassignReviewerResponse
		err = json.Unmarshal(bodyBytes, &reassignResp)
		if err != nil {
			t.Logf("Failed to unmarshal reassign response: %v", err)
			t.FailNow()
		}

		// Verify reassignment
		if reassignResp.ReplacedBy == "" {
			t.Log("New reviewer ID should be returned")
			t.FailNow()
		}

		if reassignResp.ReplacedBy == reassignReq.OldUserID {
			t.Log("New reviewer should be different from old reviewer")
			t.FailNow()
		}

		// Verify the old reviewer is no longer assigned
		foundOldReviewer := false
		for _, reviewer := range reassignResp.PR.AssignedReviewers {
			if reviewer == reassignReq.OldUserID {
				foundOldReviewer = true
				break
			}
		}
		if foundOldReviewer {
			t.Log("Old reviewer should no longer be assigned")
			t.FailNow()
		}

		// Verify new reviewer is assigned
		foundNewReviewer := false
		for _, reviewer := range reassignResp.PR.AssignedReviewers {
			if reviewer == reassignResp.ReplacedBy {
				foundNewReviewer = true
				break
			}
		}
		if !foundNewReviewer {
			t.Log("New reviewer should be assigned")
			t.FailNow()
		}

		// Verify PR status is still OPEN
		if reassignResp.PR.Status != "OPEN" {
			t.Logf("PR status should remain OPEN, got %s", reassignResp.PR.Status)
			t.FailNow()
		}
	})

	t.Run("Reassign non-assigned reviewer", func(t *testing.T) {
		// Create a team and PR
		testTeam := Team{
			Name: "reassign-team-2",
			Members: []TeamMember{
				{Id: "u6000", Name: "Eve", IsActive: true},
				{Id: "u6001", Name: "Frank", IsActive: true},
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
			PullRequestID:   "pr-reassign-002",
			PullRequestName: "Another PR",
			AuthorID:        "u6000",
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

		// Try to reassign a user who is not assigned as reviewer
		reassignReq := ReassignReviewerRequest{
			PullRequestID: "pr-reassign-002",
			OldUserID:     "u9999", // Non-assigned user
		}

		reqJSON, err = json.Marshal(reassignReq)
		if err != nil {
			t.Log("Failed to marshal reassign request")
			t.FailNow()
		}

		resp, err = http.Post(baseURL+"/pullRequest/reassign", "application/json", bytes.NewBuffer(reqJSON))
		if err != nil {
			t.Logf("Failed to send reassign request: %v", err)
			t.FailNow()
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusConflict {
			t.Logf("Should return 409 for non-assigned reviewer, got %d", resp.StatusCode)
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

		if errorResp.Error.Code != "NOT_ASSIGNED" {
			t.Logf("Error code should be NOT_ASSIGNED, got %s", errorResp.Error.Code)
			t.FailNow()
		}
	})

	t.Run("Reassign reviewer from non-existent PR", func(t *testing.T) {
		reassignReq := ReassignReviewerRequest{
			PullRequestID: "non-existent-pr",
			OldUserID:     "some-user",
		}

		reqJSON, err := json.Marshal(reassignReq)
		if err != nil {
			t.Log("Failed to marshal reassign request")
			t.FailNow()
		}

		resp, err := http.Post(baseURL+"/pullRequest/reassign", "application/json", bytes.NewBuffer(reqJSON))
		if err != nil {
			t.Logf("Failed to send reassign request: %v", err)
			t.FailNow()
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Logf("Should return 404 for non-existent PR, got %d", resp.StatusCode)
			t.FailNow()
		}
	})

	t.Run("Reassign reviewer after merge should fail", func(t *testing.T) {
		// Create a team and PR
		testTeam := Team{
			Name: "merge-reassign-team-2",
			Members: []TeamMember{
				{Id: "u11000", Name: "Ivan", IsActive: true},
				{Id: "u11001", Name: "Julia", IsActive: true},
				{Id: "u11002", Name: "Kevin", IsActive: true},
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

		// Create PR
		createPRReq := CreatePullRequestRequest{
			PullRequestID:   "pr-reassign-after-merge",
			PullRequestName: "Reassign after merge test",
			AuthorID:        "u11000",
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

		// Get the assigned reviewers
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

		// Merge the PR
		mergeReq := MergePullRequestRequest{
			PullRequestID: "pr-reassign-after-merge",
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

		// Try to reassign reviewer after merge
		reassignReq := ReassignReviewerRequest{
			PullRequestID: "pr-reassign-after-merge",
			OldUserID:     prResp.PR.AssignedReviewers[0],
		}

		reqJSON, err = json.Marshal(reassignReq)
		if err != nil {
			t.Log("Failed to marshal reassign request")
			t.FailNow()
		}

		resp, err = http.Post(baseURL+"/pullRequest/reassign", "application/json", bytes.NewBuffer(reqJSON))
		if err != nil {
			t.Logf("Failed to send reassign request: %v", err)
			t.FailNow()
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusConflict {
			t.Logf("Should return 409 for reassign on merged PR, got %d", resp.StatusCode)
			t.FailNow()
		}

		bodyBytes, err = io.ReadAll(resp.Body)
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

		if errorResp.Error.Code != "PR_MERGED" {
			t.Logf("Error code should be PR_MERGED, got %s", errorResp.Error.Code)
			t.FailNow()
		}
	})

	t.Run("Reassign reviewer with no available candidates", func(t *testing.T) {
		// Create a team with only 2 users
		testTeam := Team{
			Name: "small-reassign-team",
			Members: []TeamMember{
				{Id: "u7000", Name: "Grace", IsActive: true},
				{Id: "u7001", Name: "Henry", IsActive: true},
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

		// Create PR by Grace
		createPRReq := CreatePullRequestRequest{
			PullRequestID:   "pr-reassign-003",
			PullRequestName: "Small team PR",
			AuthorID:        "u7000",
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

		// Get the assigned reviewer (should be Henry)
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

		if len(prResp.PR.AssignedReviewers) != 1 {
			t.Logf("Should have 1 reviewer, got %d", len(prResp.PR.AssignedReviewers))
			t.FailNow()
		}

		// Try to reassign the only reviewer - should fail as no other candidates
		reassignReq := ReassignReviewerRequest{
			PullRequestID: "pr-reassign-003",
			OldUserID:     prResp.PR.AssignedReviewers[0],
		}

		reqJSON, err = json.Marshal(reassignReq)
		if err != nil {
			t.Log("Failed to marshal reassign request")
			t.FailNow()
		}

		resp, err = http.Post(baseURL+"/pullRequest/reassign", "application/json", bytes.NewBuffer(reqJSON))
		if err != nil {
			t.Logf("Failed to send reassign request: %v", err)
			t.FailNow()
		}
		defer resp.Body.Close()

		bodyBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			t.Logf("Failed to read response body: %v", err)
			t.FailNow()
		}

		if resp.StatusCode != http.StatusConflict {
			t.Logf("Should return 409 for no available candidates, got %d", resp.StatusCode)
			t.Logf("Response body: \n%v\n", string(bodyBytes))
			t.FailNow()
		}

		var errorResp ErrorResponse
		err = json.Unmarshal(bodyBytes, &errorResp)
		if err != nil {
			t.Logf("Failed to unmarshal error response: %v", err)
			t.FailNow()
		}

		if errorResp.Error.Code != "NO_CANDIDATE" {
			t.Logf("Error code should be NO_CANDIDATE, got %s", errorResp.Error.Code)
			t.FailNow()
		}
	})
}
