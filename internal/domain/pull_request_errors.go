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
