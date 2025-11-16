package service

import (
	"context"

	"github.com/raccoon00/avito-pr/internal/domain"
)

type Service struct {
	TeamRepo TeamRepository
}

func CreateService(
	teamRepo TeamRepository,
) *Service {
	return &Service{
		TeamRepo: teamRepo,
	}
}

func (s *Service) AddTeam(ctx context.Context, team *domain.Team) (*domain.Team, error) {
	insertedTeam, err := s.TeamRepo.Create(ctx, team)
	return insertedTeam, err
}
