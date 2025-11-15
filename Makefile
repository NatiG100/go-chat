.PHONY: run migrate deps

run:
	go run ./cmd/server
deps:
	go mod tidy
migrate:
	powershell -Command "Get-Content migrations/001_init.sql | docker exec -i go-chat-postgres-1 psql \"postgres://postgres:password@localhost:5432/chatdb?sslmode=disable\""