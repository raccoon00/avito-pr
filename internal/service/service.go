package service

import (
	"context"

	"github.com/raccoon00/avito-pr/internal/domain"
)

type Service struct {
	TeamRepo TeamRepository
	UserRepo UserRepository
}

func CreateService(
	teamRepo TeamRepository,
	userRepo UserRepository,
) *Service {
	return &Service{
		TeamRepo: teamRepo,
		UserRepo: userRepo,
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
