ENV_FILE?=.env
include $(ENV_FILE)
export

export CGO_ENABLED=0
export GOOS=linux


build:
	go build -ldflags="-s -w" -o bin/restapi ./cmd/restapi

run: build
	docker compose --env-file $(ENV_FILE) up --build -d

logs:
	docker compose logs -f

test: run
	go test -count=1 ./tests

auto_test: clean_down test

clean_down:
	docker compose --env-file $(ENV_FILE) down -v
