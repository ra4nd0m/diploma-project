.PHONY: up down logs run

up:
	docker compose -p achievment-service up -d

down:
	docker compose -p achievment-service down

logs:
	docker compose -p achievment-service logs

run:
	export $(shell grep -v '^#' .env.local | xargs) && go run cmd/api/main.go