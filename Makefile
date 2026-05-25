.PHONY: build test lint up down produce seed

build:
	go build -o bin/server ./cmd/server

test:
	go test ./... -race -count=1

lint:
	golangci-lint run ./...

up:
	docker compose up -d --build

down:
	docker compose down -v

produce:
	@echo '{"query":"кроссовки","user_id":"u1","timestamp":"$(shell date -u +%Y-%m-%dT%H:%M:%SZ)"}' | \
		docker compose exec -T kafka kafka-console-producer.sh \
		--broker-list localhost:9092 --topic search-events

seed:
	@for q in "кроссовки" "куртка" "платье" "телефон" "ноутбук" "сумка" "часы" "кроссовки" "кроссовки" "куртка"; do \
		echo "{\"query\":\"$$q\",\"user_id\":\"u$$RANDOM\",\"timestamp\":\"$(shell date -u +%Y-%m-%dT%H:%M:%SZ)\"}" | \
		docker compose exec -T kafka kafka-console-producer.sh \
		--broker-list localhost:9092 --topic search-events; \
	done