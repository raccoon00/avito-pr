package service

import (
	"context"
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
