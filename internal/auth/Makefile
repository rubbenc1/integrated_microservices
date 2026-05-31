include .env

.PHONY: run migrate

run:
	go run ../../cmd/auth/main.go

migrate:
	docker run --rm --network host \
		-v $(CURDIR)/migrations:/migrations \
		migrate/migrate \
		-path=/migrations \
		-database "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:$(DATABASE_PORT)/$(POSTGRES_DB)?sslmode=disable" \
		up