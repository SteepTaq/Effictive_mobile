.PHONY: build run clean infra-up infra-down migrate-up migrate-down migrate-reset

NAME=subscriptions
MAIN=cmd/service/main.go
MIGRATE=cmd/migrate/main.go

build:
	go build -o $(NAME) $(MAIN)

run: build
	./$(NAME)

clean:
	rm -f $(NAME)

infra-up:
	docker-compose up -d

infra-down:
	docker-compose down

infra-build:
	docker-compose up --build -d

migrate-up:
	go run $(MIGRATE) -action=up

migrate-down:
	go run $(MIGRATE) -action=down

migrate-reset:
	$(MAKE) migrate-down
	$(MAKE) migrate-up


