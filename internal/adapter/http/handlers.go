package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

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

type ErrorBody struct {
	Code    ErrorCode `json:"code" binding:"required"`
	Message string    `json:"message" binding:"required"`
}

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

func (s *GinService) TeamAdd(c *gin.Context) {
	ctx := context.Background()

	var team Team
	if err := c.ShouldBind(&team); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: ErrorBody{
			Code:    BAD_REQUEST,
			Message: err.Error(),
		}})
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
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: ErrorBody{
				Code:    TEAM_EXISTS,
				Message: fmt.Sprintf("Team %s already exists", newTeam.Name),
			}})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: ErrorBody{
				Code:    UNHANDLED_SERVER_ERROR,
				Message: err.Error(),
			}})
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

func (s *GinService) TeamGet(c *gin.Context) {
	ctx := context.Background()

	teamName := c.Query("team_name")
	if teamName == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: ErrorBody{
			Code:    BAD_REQUEST,
			Message: "team_name query parameter is required",
		}})
		return
	}

	team, err := s.srv.GetTeam(ctx, teamName)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: ErrorBody{
			Code:    NOT_FOUND,
			Message: fmt.Sprintf("Team %s not found", teamName),
		}})
		return
	}

	responseTeam := Team{
		Name:    team.Name,
		Members: make([]Member, 0, len(team.Members)),
	}

	for _, member := range team.Members {
		responseTeam.Members = append(responseTeam.Members, Member{
			Id:       member.Id,
			Name:     member.Name,
			IsActive: member.IsActive,
		})
	}

	c.JSON(http.StatusOK, responseTeam)
}

type SetUserIsActiveRequest struct {
	UserID string `json:"user_id" binding:"required"`
	// Ссылка на bool потому что gin не понимает разницу между
	// false и отсутствующим значением
	IsActive *bool `json:"is_active" binding:"required"`
}

type UserResponse struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

func (s *GinService) SetUserIsActive(c *gin.Context) {
	ctx := context.Background()

	var req SetUserIsActiveRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: ErrorBody{
			Code:    BAD_REQUEST,
			Message: err.Error(),
		}})
		return
	}

	user, err := s.srv.SetUserIsActive(ctx, req.UserID, *req.IsActive)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: ErrorBody{
			Code:    NOT_FOUND,
			Message: fmt.Sprintf("User %s not found", req.UserID),
		}})
		return
	}

	responseUser := UserResponse{
		UserID:   user.Id,
		Username: user.Name,
		TeamName: user.Team,
		IsActive: user.IsActive,
	}

	c.JSON(http.StatusOK, gin.H{
		"user": responseUser,
	})
}

type CreatePullRequestRequest struct {
	PullRequestID   string `json:"pull_request_id" binding:"required"`
	PullRequestName string `json:"pull_request_name" binding:"required"`
	AuthorID        string `json:"author_id" binding:"required"`
}

type PullRequestResponse struct {
	PullRequestID     string   `json:"pull_request_id"`
	PullRequestName   string   `json:"pull_request_name"`
	AuthorID          string   `json:"author_id"`
	Status            string   `json:"status"`
	AssignedReviewers []string `json:"assigned_reviewers"`
	CreatedAt         *string  `json:"createdAt,omitempty"`
	MergedAt          *string  `json:"mergedAt,omitempty"`
}

func (s *GinService) CreatePullRequest(c *gin.Context) {
	ctx := context.Background()

	var req CreatePullRequestRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: ErrorBody{
			Code:    BAD_REQUEST,
			Message: err.Error(),
		}})
		return
	}

	pr, err := s.srv.CreatePullRequest(ctx, req.PullRequestID, req.PullRequestName, req.AuthorID)
	if err != nil {
		var prExistsErr *domain.PullRequestExistsError
		var authorNotFoundErr *domain.AuthorNotFoundError
		var teamNotFoundErr *domain.TeamNotFoundError

		if errors.As(err, &prExistsErr) {
			c.JSON(http.StatusConflict, ErrorResponse{Error: ErrorBody{
				Code:    PR_EXISTS,
				Message: fmt.Sprintf("PR id %s already exists", req.PullRequestID),
			}})
		} else if errors.As(err, &authorNotFoundErr) || errors.As(err, &teamNotFoundErr) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: ErrorBody{
				Code:    NOT_FOUND,
				Message: err.Error(),
			}})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: ErrorBody{
				Code:    UNHANDLED_SERVER_ERROR,
				Message: err.Error(),
			}})
		}
		return
	}

	responsePR := PullRequestResponse{
		PullRequestID:     pr.ID,
		PullRequestName:   pr.Name,
		AuthorID:          pr.AuthorID,
		Status:            string(pr.Status),
		AssignedReviewers: pr.AssignedReviewers,
	}

	if pr.CreatedAt != nil {
		createdAtStr := pr.CreatedAt.Format(time.RFC3339)
		responsePR.CreatedAt = &createdAtStr
	}
	if pr.MergedAt != nil {
		mergedAtStr := pr.MergedAt.Format(time.RFC3339)
		responsePR.MergedAt = &mergedAtStr
	}

	c.JSON(http.StatusCreated, gin.H{
		"pr": responsePR,
	})
}
