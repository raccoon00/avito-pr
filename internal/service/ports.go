package service

import (
	"context"

	"github.com/raccoon00/avito-pr/internal/domain"
)

type TeamRepository interface {
	Create(ctx context.Context, team *domain.Team) (*domain.Team, error)
	Get(ctx context.Context, teamName string) (*domain.Team, error)
}

type UserRepository interface {
	SetIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error)
	GetByID(ctx context.Context, userID string) (*domain.User, error)
	GetActiveTeamMembers(ctx context.Context, teamName string, excludeUserID string) ([]domain.User, error)
}

type PullRequestRepository interface {
	Create(ctx context.Context, pr *domain.PullRequest) (*domain.PullRequest, error)
	GetByID(ctx context.Context, prID string) (*domain.PullRequest, error)
	Exists(ctx context.Context, prID string) (bool, error)
	Update(ctx context.Context, pr *domain.PullRequest) (*domain.PullRequest, error)
	GetByReviewer(ctx context.Context, userID string) ([]domain.PullRequest, error)
}
