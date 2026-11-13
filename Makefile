# Загружаем переменные из файла crd.env
include .env
export

# Строка подключения для миграций (собирается из переменных)
DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@localhost:5432/$(DB_NAME)?sslmode=disable

# Применить все миграции
migrate-up:
	migrate -database "$(DB_URL)" -path migrations up

# Откатить последнюю миграцию
migrate-down:
	migrate -database "$(DB_URL)" -path migrations down 1

# Запуск приложения
run:
	go run ./cmd/app

# Сборка бинарника
build:
	go build -o bin/app ./cmd/app

# Запуск скомпилированного бинарника
start: build
	./bin/app