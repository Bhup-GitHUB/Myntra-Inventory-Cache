run:
	go run ./cmd/api

worker:
	go run ./cmd/worker

seed:
	go run ./cmd/seed

docker-up:
	docker compose up --build

docker-down:
	docker compose down -v

migrate-up:
	goose -dir migrations mysql "$${MYSQL_DSN}" up

test:
	go test ./...

load-read:
	./scripts/load-test-read.sh

load-checkout:
	./scripts/load-test-checkout.sh
