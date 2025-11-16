package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/raccoon00/avito-pr/internal/domain"
	"github.com/raccoon00/avito-pr/internal/service"
)

type ErrorCode string

const (
	TEAM_EXISTS  ErrorCode = "TEAM_EXISTS"
	PR_EXISTS    ErrorCode = "PR_EXISTS"
	PR_MERGED    ErrorCode = "PR_MERGED"
	NOT_ASSIGNED ErrorCode = "NOT_ASSIGNED"
	NO_CANDIDATE ErrorCode = "NO_CANDIDATE"
	NOT_FOUND    ErrorCode = "NOT_FOUND"

	BAD_REQUEST            ErrorCode = "BAD_REQUEST"
	UNHANDLED_SERVER_ERROR ErrorCode = "UNHANDLED_SERVER_ERROR"
)

type GinService struct {
	srv *service.Service
}

type Team struct {
	Name    string   `json:"team_name" binding:"required"`
	Members []Member `json:"members" binding:"required"`
}

type Member struct {
	Id       string `json:"user_id" binding:"required"`
	Name     string `json:"username" binding:"required"`
	IsActive bool   `json:"is_active" binding:"required"`
}

type ErrorResponse struct {
	Code    ErrorCode `json:"code" binding:"required"`
	Message string    `json:"message" binding:"required"`
}

func (s *GinService) TeamAdd(c *gin.Context) {
	ctx := context.Background()

	var team Team
	if err := c.ShouldBind(&team); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    BAD_REQUEST,
			Message: err.Error(),
		})
		return
	}

	newTeam := domain.Team{Name: team.Name, Members: make([]domain.User, 0, len(team.Members))}
	for _, member := range team.Members {
		newTeam.Members = append(newTeam.Members, domain.User{Id: member.Id, Name: member.Name, Team: team.Name, IsActive: member.IsActive})
	}

	insertedTeam, err := s.srv.AddTeam(ctx, &newTeam)
	if err != nil {
		var errTeamExits *domain.TeamExistsError
		if errors.As(err, &errTeamExits) {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Code:    TEAM_EXISTS,
				Message: fmt.Sprintf("Team %s already exists", newTeam.Name),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Code:    UNHANDLED_SERVER_ERROR,
				Message: err.Error(),
			})
		}
		return
	}

	createdTeam := Team{
		Name:    insertedTeam.Name,
		Members: make([]Member, 0, len(insertedTeam.Members)),
	}

	for _, member := range insertedTeam.Members {
		createdTeam.Members = append(createdTeam.Members, Member{
			Id:       member.Id,
			Name:     member.Name,
			IsActive: member.IsActive,
		})
	}

	c.JSON(http.StatusCreated, gin.H{
		"team": createdTeam,
	})
}
