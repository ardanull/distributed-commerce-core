APP=distributed-commerce-core

bootstrap:
	go mod tidy

fmt:
	gofmt -w $(shell find . -type f -name '*.go')

test:
	go test ./...

race:
	go test -race ./...

up:
	docker compose up --build -d

down:
	docker compose down -v

migrate:
	bash scripts/migrate.sh

smoke:
	bash scripts/smoke.sh

zip:
	cd .. && zip -r $(APP).zip $(APP)
