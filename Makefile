postgres:
	docker run --name pg -p 5432:5432 -e POSTGRES_PASSWORD=postgres -d postgres

createdb:
	docker exec -it pg createdb --username=postgres --owner=postgres bank

dropdb:
	docker exec -it pg dropdb -U postgres bank

migrateup:
	migrate -path db/migration -database "postgresql://postgres:postgres@localhost:5432/bank?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://postgres:postgres@localhost:5432/bank?sslmode=disable" -verbose down

test:
	go test -v -cover ./...

runserver:
	go run main.go

.PHONY: postgres createdb dropdb migrateup migratedown test runserver