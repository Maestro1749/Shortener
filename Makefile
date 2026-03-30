migrate-up:
	migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/ShortenerDB?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/ShortenerDB?sslmode=disable" down 1