.PHONY: migrate

migrate:
	tern migrate --migrations ./internal/store/pgstore/migrations --config ./internal/store/pgstore/migrations/tern.conf

run-api:
	air --build.cmd "go build -o ./bin/api ./cmd/api" --build.bin "./bin/api"