package app

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raccoon00/avito-pr/internal/adapter/http"
	"github.com/raccoon00/avito-pr/internal/adapter/postgres"
	"github.com/raccoon00/avito-pr/internal/config"
	"github.com/raccoon00/avito-pr/internal/service"
)

func Run() {
	cfg := config.Load()
	ctx_root := context.Background()

	conn, err := connectToPostgres(ctx_root, cfg)
	if err != nil {
		log.Fatalf("Could not connect to database %v", err)
	}
	defer conn.Close()

	team_repo := postgres.NewTeamRepo(conn, "teams", "users")
	user_repo := postgres.NewUserRepo(conn, "users")
	srv := service.CreateService(team_repo, user_repo)

	http.Run(srv)
}

func connectToPostgres(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	conn, err := pgxpool.New(ctx, cfg.GetDBConnectionString())
	return conn, err
}
