package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

type MergePullRequestRequest struct {
	PullRequestID string `json:"pull_request_id"`
}

func TestMergePullRequest(t *testing.T) {
	port := os.Getenv("SERVICE_PORT")
	if port == "" {
		port = "8080"
	}

	baseURL := fmt.Sprintf("http://localhost:%s", port)

	t.Run("Successfully merge open PR", func(t *testing.T) {
		// Create a team and PR
		testTeam := Team{
			Name: "merge-team",
			Members: []TeamMember{
				{Id: "u8000", Name: "Alice", IsActive: true},
				{Id: "u8001", Name: "Bob", IsActive: true},
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
			PullRequestID:   "pr-merge-001",
			PullRequestName: "Feature to merge",
			AuthorID:        "u8000",
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

		// Verify initial status is OPEN
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

		if prResp.PR.Status != "OPEN" {
			t.Logf("PR status should be OPEN initially, got %s", prResp.PR.Status)
			t.FailNow()
		}

		// Merge the PR
		mergeReq := MergePullRequestRequest{
			PullRequestID: "pr-merge-001",
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

		bodyBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			t.Logf("Failed to read response body: %v", err)
			t.FailNow()
		}

		var mergeResp PullRequestResponse
		err = json.Unmarshal(bodyBytes, &mergeResp)
		if err != nil {
			t.Logf("Failed to unmarshal merge response: %v", err)
			t.FailNow()
		}

		// Verify PR is merged
		if mergeResp.PR.Status != "MERGED" {
			t.Logf("PR status should be MERGED, got %s", mergeResp.PR.Status)
			t.FailNow()
		}

		if mergeResp.PR.MergedAt == nil {
			t.Log("MergedAt timestamp should be set")
			t.FailNow()
		}

		// Verify merged timestamp is recent
		mergedTime, err := time.Parse(time.RFC3339, *mergeResp.PR.MergedAt)
		if err != nil {
			t.Logf("Failed to parse MergedAt timestamp: %v", err)
			t.FailNow()
		}

		if time.Since(mergedTime) > 5*time.Second {
			t.Log("MergedAt should be recent")
			t.FailNow()
		}

		// Verify reviewers are preserved
		if len(mergeResp.PR.AssignedReviewers) != len(prResp.PR.AssignedReviewers) {
			t.Logf("Reviewers count should remain the same after merge, got %d vs %d",
				len(mergeResp.PR.AssignedReviewers), len(prResp.PR.AssignedReviewers))
			t.FailNow()
		}
	})

	t.Run("Idempotent merge operation", func(t *testing.T) {
		// Create a team and PR
		testTeam := Team{
			Name: "idempotent-team",
			Members: []TeamMember{
				{Id: "u9000", Name: "Charlie", IsActive: true},
				{Id: "u9001", Name: "Diana", IsActive: true},
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
			PullRequestID:   "pr-merge-002",
			PullRequestName: "Idempotent test",
			AuthorID:        "u9000",
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

		// Merge the PR first time
		mergeReq := MergePullRequestRequest{
			PullRequestID: "pr-merge-002",
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
			t.Logf("First merge should succeed, got status: %d", resp.StatusCode)
			t.FailNow()
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Logf("Failed to read response body: %v", err)
			t.FailNow()
		}

		var firstMergeResp PullRequestResponse
		err = json.Unmarshal(bodyBytes, &firstMergeResp)
		if err != nil {
			t.Logf("Failed to unmarshal first merge response: %v", err)
			t.FailNow()
		}

		firstMergedAt := firstMergeResp.PR.MergedAt

		// Merge the same PR again (idempotent operation)
		resp, err = http.Post(baseURL+"/pullRequest/merge", "application/json", bytes.NewBuffer(reqJSON))
		if err != nil {
			t.Logf("Failed to merge PR again: %v", err)
			t.FailNow()
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Logf("Second merge should succeed, got status: %d", resp.StatusCode)
			t.FailNow()
		}

		bodyBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			t.Logf("Failed to read response body: %v", err)
			t.FailNow()
		}

		var secondMergeResp PullRequestResponse
		err = json.Unmarshal(bodyBytes, &secondMergeResp)
		if err != nil {
			t.Logf("Failed to unmarshal second merge response: %v", err)
			t.FailNow()
		}

		// Verify status remains MERGED
		if secondMergeResp.PR.Status != "MERGED" {
			t.Logf("PR status should remain MERGED, got %s", secondMergeResp.PR.Status)
			t.FailNow()
		}

		// Verify MergedAt timestamp is preserved (not updated)
		if secondMergeResp.PR.MergedAt == nil {
			t.Log("MergedAt should still be set")
			t.FailNow()
		}

		if *firstMergedAt != *secondMergeResp.PR.MergedAt {
			t.Logf("MergedAt timestamp should not change on subsequent merges, got %s vs %s",
				*firstMergedAt, *secondMergeResp.PR.MergedAt)
			t.FailNow()
		}
	})

	t.Run("Merge non-existent PR", func(t *testing.T) {
		mergeReq := MergePullRequestRequest{
			PullRequestID: "non-existent-pr-merge",
		}

		reqJSON, err := json.Marshal(mergeReq)
		if err != nil {
			t.Log("Failed to marshal merge request")
			t.FailNow()
		}

		resp, err := http.Post(baseURL+"/pullRequest/merge", "application/json", bytes.NewBuffer(reqJSON))
		if err != nil {
			t.Logf("Failed to send merge request: %v", err)
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
			Name: "merge-reassign-team",
			Members: []TeamMember{
				{Id: "u10000", Name: "Eve", IsActive: true},
				{Id: "u10001", Name: "Frank", IsActive: true},
				{Id: "u10002", Name: "Grace", IsActive: true},
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
			PullRequestID:   "pr-merge-reassign",
			PullRequestName: "Merge then reassign test",
			AuthorID:        "u10000",
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
			PullRequestID: "pr-merge-reassign",
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
			PullRequestID: "pr-merge-reassign",
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
}
