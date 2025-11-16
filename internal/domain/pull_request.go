package domain

import "time"

type PullRequestStatus string

const (
	PullRequestStatusOpen   PullRequestStatus = "OPEN"
	PullRequestStatusMerged PullRequestStatus = "MERGED"
)

type PullRequest struct {
	ID                string
	Name              string
	AuthorID          string
	Status            PullRequestStatus
	AssignedReviewers []string
	CreatedAt         *time.Time
	MergedAt          *time.Time
}
