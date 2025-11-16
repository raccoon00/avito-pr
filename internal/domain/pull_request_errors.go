package domain

import "fmt"

type PullRequestExistsError struct {
	PullRequestID string
}

func (e *PullRequestExistsError) Error() string {
	return fmt.Sprintf("Pull request %s already exists", e.PullRequestID)
}

type AuthorNotFoundError struct {
	AuthorID string
}

func (e *AuthorNotFoundError) Error() string {
	return fmt.Sprintf("Author %s not found", e.AuthorID)
}

type TeamNotFoundError struct {
	TeamName string
}

func (e *TeamNotFoundError) Error() string {
	return fmt.Sprintf("Team %s not found", e.TeamName)
}

type NoReviewersAvailableError struct {
	TeamName string
}

func (e *NoReviewersAvailableError) Error() string {
	return fmt.Sprintf("No reviewers available in team %s", e.TeamName)
}

type PRMergedError struct {
	PullRequestID string
}

func (e *PRMergedError) Error() string {
	return fmt.Sprintf("cannot reassign on merged PR %s", e.PullRequestID)
}

type ReviewerNotAssignedError struct {
	PullRequestID string
	UserID        string
}

func (e *ReviewerNotAssignedError) Error() string {
	return fmt.Sprintf("reviewer %s is not assigned to PR %s", e.UserID, e.PullRequestID)
}

type UserNotFoundError struct {
	UserID string
}

func (e *UserNotFoundError) Error() string {
	return fmt.Sprintf("user %s not found", e.UserID)
}
