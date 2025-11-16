package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raccoon00/avito-pr/internal/domain"
	"github.com/raccoon00/avito-pr/internal/service"
)

type PostgresPullRequestTable struct {
	Conn    *pgxpool.Pool
	PRTable string
}

func NewPullRequestRepo(
	conn *pgxpool.Pool,
	prTable string,
) service.PullRequestRepository {
	return &PostgresPullRequestTable{Conn: conn, PRTable: prTable}
}

func (p *PostgresPullRequestTable) Create(ctx context.Context, pr *domain.PullRequest) (*domain.PullRequest, error) {
	insertQuery := fmt.Sprintf(
		"INSERT INTO %s (pull_request_id, pull_request_name, author_id, status, assigned_reviewers, created_at, merged_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING pull_request_id, pull_request_name, author_id, status, assigned_reviewers, created_at, merged_at",
		p.PRTable,
	)

	row := p.Conn.QueryRow(
		ctx,
		insertQuery,
		pr.ID, pr.Name, pr.AuthorID, string(pr.Status), pr.AssignedReviewers, pr.CreatedAt, pr.MergedAt,
	)

	var createdPR domain.PullRequest
	var status string
	err := row.Scan(
		&createdPR.ID,
		&createdPR.Name,
		&createdPR.AuthorID,
		&status,
		&createdPR.AssignedReviewers,
		&createdPR.CreatedAt,
		&createdPR.MergedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" { // unique_violation
				return nil, &domain.PullRequestExistsError{PullRequestID: pr.ID}
			}
		}
		return nil, fmt.Errorf("unhandled error inserting PR into Postgres PR table: %w", err)
	}

	createdPR.Status = domain.PullRequestStatus(status)
	return &createdPR, nil
}

func (p *PostgresPullRequestTable) GetByID(ctx context.Context, prID string) (*domain.PullRequest, error) {
	selectQuery := fmt.Sprintf(
		"SELECT pull_request_id, pull_request_name, author_id, status, assigned_reviewers, created_at, merged_at FROM %s WHERE pull_request_id = $1",
		p.PRTable,
	)

	row := p.Conn.QueryRow(ctx, selectQuery, prID)

	var pr domain.PullRequest
	var status string
	err := row.Scan(
		&pr.ID,
		&pr.Name,
		&pr.AuthorID,
		&status,
		&pr.AssignedReviewers,
		&pr.CreatedAt,
		&pr.MergedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("pull request not found: %w", err)
		}
		return nil, fmt.Errorf("error getting pull request: %w", err)
	}

	pr.Status = domain.PullRequestStatus(status)
	return &pr, nil
}

func (p *PostgresPullRequestTable) Exists(ctx context.Context, prID string) (bool, error) {
	selectQuery := fmt.Sprintf(
		"SELECT EXISTS(SELECT 1 FROM %s WHERE pull_request_id = $1)",
		p.PRTable,
	)

	row := p.Conn.QueryRow(ctx, selectQuery, prID)

	var exists bool
	err := row.Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error checking if PR exists: %w", err)
	}

	return exists, nil
}

func (p *PostgresPullRequestTable) Update(ctx context.Context, pr *domain.PullRequest) (*domain.PullRequest, error) {
	updateQuery := fmt.Sprintf(
		"UPDATE %s SET pull_request_name = $1, author_id = $2, status = $3, assigned_reviewers = $4, created_at = $5, merged_at = $6 WHERE pull_request_id = $7 RETURNING pull_request_id, pull_request_name, author_id, status, assigned_reviewers, created_at, merged_at",
		p.PRTable,
	)

	row := p.Conn.QueryRow(
		ctx,
		updateQuery,
		pr.Name, pr.AuthorID, string(pr.Status), pr.AssignedReviewers, pr.CreatedAt, pr.MergedAt, pr.ID,
	)

	var updatedPR domain.PullRequest
	var status string
	err := row.Scan(
		&updatedPR.ID,
		&updatedPR.Name,
		&updatedPR.AuthorID,
		&status,
		&updatedPR.AssignedReviewers,
		&updatedPR.CreatedAt,
		&updatedPR.MergedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("error updating pull request: %w", err)
	}

	updatedPR.Status = domain.PullRequestStatus(status)
	return &updatedPR, nil
}
