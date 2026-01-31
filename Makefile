.PHONY: test test-backend build build-backend build-frontend lint lint-frontend \
        dev-backend dev-frontend docker-up docker-down docker-build

# Default target
all: build

# Testing
test: test-backend

test-backend:
	cd backend && go test ./...

# Building
build: build-backend build-frontend

build-backend:
	cd backend && go build ./...

build-frontend:
	cd frontend && npm run build

# Linting
lint: lint-frontend

lint-frontend:
	cd frontend && npm run lint

# Local development
dev-backend:
	cd backend && go run cmd/api/main.go

dev-frontend:
	cd frontend && npm run dev

# Docker
docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-build:
	docker-compose build
