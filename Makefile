include .env
export

DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@localhost:5432/$(DB_NAME)?sslmode=disable

migrate-up:
	migrate -database "$(DB_URL)" -path migrations up

migrate-down:
	migrate -database "$(DB_URL)" -path migrations down 1

run:
	go run ./cmd/app

build:
	go build -o bin/app ./cmd/app

start: build
	./bin/app