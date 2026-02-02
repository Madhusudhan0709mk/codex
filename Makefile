.PHONY: bootstrap up down

bootstrap:
	@echo "No-op bootstrap. Ensure Docker is running."

up:
	docker compose -f infra/docker/docker-compose.yml up --build

down:
	docker compose -f infra/docker/docker-compose.yml down
