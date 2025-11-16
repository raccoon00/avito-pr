export CGO_ENABLED=0
export GOOS=linux

ENV_FILE?=.env

build:
	go build -ldflags="-s -w" -o bin/restapi ./cmd/restapi

run: build
	docker compose --env-file $(ENV_FILE) up --build -d
	docker compose logs -f

clean_down:
	docker compose down -v
