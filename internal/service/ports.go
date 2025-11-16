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
}
