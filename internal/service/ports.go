package service

import (
	"context"

	"github.com/raccoon00/avito-pr/internal/domain"
)

type TeamRepository interface {
	Create(ctx context.Context, team *domain.Team) (*domain.Team, error)
}
