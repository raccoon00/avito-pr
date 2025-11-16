package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raccoon00/avito-pr/internal/domain"
	"github.com/raccoon00/avito-pr/internal/service"
)

type PostgresUserTable struct {
	Conn       *pgxpool.Pool
	UsersTable string
}

func NewUserRepo(
	conn *pgxpool.Pool,
	usersTable string,
) service.UserRepository {
	return &PostgresUserTable{Conn: conn, UsersTable: usersTable}
}

func (u *PostgresUserTable) SetIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error) {
	updateQuery := fmt.Sprintf(
		"UPDATE %s SET is_active = $1 WHERE user_id = $2 RETURNING user_id, username, is_active, team_name",
		u.UsersTable,
	)

	row := u.Conn.QueryRow(ctx, updateQuery, isActive, userID)

	var user domain.User
	err := row.Scan(&user.Id, &user.Name, &user.IsActive, &user.Team)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return &user, nil
}

func (u *PostgresUserTable) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	selectQuery := fmt.Sprintf(
		"SELECT user_id, username, is_active, team_name FROM %s WHERE user_id = $1",
		u.UsersTable,
	)

	row := u.Conn.QueryRow(ctx, selectQuery, userID)

	var user domain.User
	err := row.Scan(&user.Id, &user.Name, &user.IsActive, &user.Team)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return &user, nil
}
