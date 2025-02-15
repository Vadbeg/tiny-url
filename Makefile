build:
	docker compose build

run:
	docker compose up

local-frontend:
	cd frontend && go run main.go

local-backend:
	cd backend && go run main.go