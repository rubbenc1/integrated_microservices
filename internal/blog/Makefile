include .env

.PHONY: run migrate

migrate:
	docker run --rm --network host \
		-v $(CURDIR)/migrations:/migrations \
		migrate/migrate \
		-path=/migrations \
		-database "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:$(DATABASE_PORT)/$(POSTGRES_DB)?sslmode=disable" \
		up
run:
	go run ../../cmd/blog/main.go