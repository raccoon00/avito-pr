package service

import (
	"context"
	"slices"
	"time"

	"github.com/raccoon00/avito-pr/internal/domain"
)

type Service struct {
	TeamRepo TeamRepository
	UserRepo UserRepository
	PRRepo   PullRequestRepository
}

func CreateService(
	teamRepo TeamRepository,
	userRepo UserRepository,
	prRepo PullRequestRepository,
) *Service {
	return &Service{
		TeamRepo: teamRepo,
		UserRepo: userRepo,
		PRRepo:   prRepo,
	}
}

func (s *Service) AddTeam(ctx context.Context, team *domain.Team) (*domain.Team, error) {
	insertedTeam, err := s.TeamRepo.Create(ctx, team)
	return insertedTeam, err
}

func (s *Service) GetTeam(ctx context.Context, teamName string) (*domain.Team, error) {
	team, err := s.TeamRepo.Get(ctx, teamName)
	return team, err
}

func (s *Service) SetUserIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error) {
	user, err := s.UserRepo.SetIsActive(ctx, userID, isActive)
	return user, err
}

func (s *Service) ReassignReviewer(ctx context.Context, prID, oldUserID string) (*domain.PullRequest, string, error) {
	// Get the pull request
	pr, err := s.PRRepo.GetByID(ctx, prID)
	if err != nil {
		return nil, "", err
	}

	// Check if PR is merged
	if pr.Status == domain.PullRequestStatusMerged {
		return nil, "", &domain.PRMergedError{PullRequestID: prID}
	}

	// Check if old user is assigned as reviewer
	found := slices.Contains(pr.AssignedReviewers, oldUserID)
	if !found {
		return nil, "", &domain.ReviewerNotAssignedError{PullRequestID: prID, UserID: oldUserID}
	}

	// Get the old user to find their team
	oldUser, err := s.UserRepo.GetByID(ctx, oldUserID)
	if err != nil {
		return nil, "", &domain.UserNotFoundError{UserID: oldUserID}
	}

	// Get active team members excluding the old user and current reviewers
	availableReviewers, err := s.UserRepo.GetActiveTeamMembers(ctx, oldUser.Team, pr.AuthorID)
	if err != nil {
		return nil, "", err
	}

	// Filter out current reviewers
	var candidates []domain.User
	for _, candidate := range availableReviewers {
		isCurrentReviewer := slices.Contains(pr.AssignedReviewers, candidate.Id)
		if !isCurrentReviewer {
			candidates = append(candidates, candidate)
		}
	}

	if len(candidates) == 0 {
		return nil, "", &domain.NoReviewersAvailableError{TeamName: oldUser.Team}
	}

	// Select first available candidate
	newReviewer := candidates[0]

	// Replace old reviewer with new reviewer
	newReviewers := make([]string, len(pr.AssignedReviewers))
	for i, reviewer := range pr.AssignedReviewers {
		if reviewer == oldUserID {
			newReviewers[i] = newReviewer.Id
		} else {
			newReviewers[i] = reviewer
		}
	}

	// Update the pull request
	pr.AssignedReviewers = newReviewers
	updatedPR, err := s.PRRepo.Update(ctx, pr)
	if err != nil {
		return nil, "", err
	}

	return updatedPR, newReviewer.Id, nil
}

func (s *Service) MergePullRequest(ctx context.Context, prID string) (*domain.PullRequest, error) {
	// Get the pull request
	pr, err := s.PRRepo.GetByID(ctx, prID)
	if err != nil {
		return nil, err
	}

	// If already merged, return current state (idempotent)
	if pr.Status == domain.PullRequestStatusMerged {
		return pr, nil
	}

	// Update status to MERGED and set merged_at timestamp
	now := time.Now()
	pr.Status = domain.PullRequestStatusMerged
	pr.MergedAt = &now

	updatedPR, err := s.PRRepo.Update(ctx, pr)
	if err != nil {
		return nil, err
	}

	return updatedPR, nil
}

func (s *Service) CreatePullRequest(ctx context.Context, prID, prName, authorID string) (*domain.PullRequest, error) {
	// Check if PR already exists
	exists, err := s.PRRepo.Exists(ctx, prID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, &domain.PullRequestExistsError{PullRequestID: prID}
	}

	// Get author to find their team
	author, err := s.UserRepo.GetByID(ctx, authorID)
	if err != nil {
		return nil, &domain.AuthorNotFoundError{AuthorID: authorID}
	}

	// Get active team members excluding author
	reviewers, err := s.UserRepo.GetActiveTeamMembers(ctx, author.Team, authorID)
	if err != nil {
		return nil, err
	}

	// Нужно инициализировать пустым массивом, иначе gin будет считать
	// что ничего не было передано
	var assignedReviewers []string = []string{}
	maxReviewers := min(len(reviewers), 2)

	for i := range maxReviewers {
		assignedReviewers = append(assignedReviewers, reviewers[i].Id)
	}

	now := time.Now()
	pr := &domain.PullRequest{
		ID:                prID,
		Name:              prName,
		AuthorID:          authorID,
		Status:            domain.PullRequestStatusOpen,
		AssignedReviewers: assignedReviewers,
		CreatedAt:         &now,
		MergedAt:          nil,
	}

	return s.PRRepo.Create(ctx, pr)
}
