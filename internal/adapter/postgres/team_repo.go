package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raccoon00/avito-pr/internal/domain"
	"github.com/raccoon00/avito-pr/internal/service"
)

type PostgresTeamTable struct {
	Conn       *pgxpool.Conn
	TeamTable  string
	UsersTable string
}

func NewTeamRepo(
	conn *pgxpool.Conn,
	teamTable string,
	usersTable string,
) service.TeamRepository {
	return &PostgresTeamTable{Conn: conn, TeamTable: teamTable, UsersTable: usersTable}
}

func (t *PostgresTeamTable) Create(ctx context.Context, team domain.Team) error {
	_, err := t.Conn.Exec(
		ctx,
		"INSERT INTO $1 (team_name) VALUES ($2)",
		t.TeamTable,
		team.Name,
	)
	if errors.Is(err, pg)
}
