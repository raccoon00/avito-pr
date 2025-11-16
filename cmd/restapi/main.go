package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/raccoon00/avito-pr/internal/config"
)

func main() {
	log.SetOutput(os.Stdout)

	url := config.Load().GetDBConnectionString()
	var conn *pgx.Conn
	var err error
	for _ = range 10 {
		conn, err = pgx.Connect(context.Background(), url)
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}
	if err != nil {
		log.Fatalf("Unable to connect to database on url %v, %v", url, err)
	}
	defer conn.Close(context.Background())

	conn.Ping(context.Background())
}
