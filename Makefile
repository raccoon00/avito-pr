export CGO_ENABLED=0
export GOOS=linux

ENV_FILE?=.env

build:
	go build -ldflags="-s -w" -o bin/restapi ./cmd/restapi

run: build
	docker compose --env-file $(ENV_FILE) up --build -d

logs:
	docker compose logs -f

test: run
	go test -count=1 ./tests

auto_test: test clean_down

clean_down:
	docker compose --env-file $(ENV_FILE) down -v
