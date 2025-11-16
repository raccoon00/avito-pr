package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raccoon00/avito-pr/internal/domain"
	"github.com/raccoon00/avito-pr/internal/service"
)

type PostgresTeamTable struct {
	Conn       *pgxpool.Pool
	TeamTable  string
	UsersTable string
}

func NewTeamRepo(
	conn *pgxpool.Pool,
	teamTable string,
	usersTable string,
) service.TeamRepository {
	return &PostgresTeamTable{Conn: conn, TeamTable: teamTable, UsersTable: usersTable}
}

func (t *PostgresTeamTable) Create(ctx context.Context, team *domain.Team) (*domain.Team, error) {
	row := t.Conn.QueryRow(
		ctx,
		"INSERT INTO $1 (team_name) VALUES ($2) RETURNING team_name;",
		t.TeamTable,
		team.Name,
	)

	var team_name string
	err := row.Scan(&team_name)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "42710" { // duplicate_object
				return nil, &domain.TeamExistsError{TeamName: team.Name}
			}
		}
		return nil, fmt.Errorf("Unhandled error inserting team into Postgres Team table: %w", err)
	}
	newTeam := domain.Team{Name: team_name, Members: make([]domain.User, 0, len(team.Members))}

	for _, member := range team.Members {
		row := t.Conn.QueryRow(
			ctx,
			"INSERT INTO $1 (user_id, username, is_active, team_name) VALUES ($2, $3, $4, $5) RETURNING user_id, username, is_active, team_name;",
			t.UsersTable, member.Id, member.Name, member.IsActive, member.Team,
		)

		var user domain.User
		err := row.Scan(&user.Id, &user.Name, &user.IsActive, &user.Team)

		if err != nil {
			return &newTeam, fmt.Errorf("Unhandled error inerting user into Postgres Users table: %w", err)
		}

		newTeam.Members = append(newTeam.Members, user)
	}

	return &newTeam, nil
}
