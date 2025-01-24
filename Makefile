postgres:
	docker run --name pg -p 5432:5432 -e POSTGRES_PASSWORD=postgres -d postgres

createdb:
	docker exec -it pg createdb --username=postgres --owner=postgres bank

dropdb:
	docker exec -it pg dropdb -U postgres bank

migrateup:
	migrate -path db/migration -database "postgresql://postgres:postgres@localhost:5432/bank?sslmode=disable" -verbose up

migrateup1:
	migrate -path db/migration -database "postgresql://postgres:postgres@localhost:5432/bank?sslmode=disable" -verbose up 1

migratedown:
	migrate -path db/migration -database "postgresql://postgres:postgres@localhost:5432/bank?sslmode=disable" -verbose down

migratedown1:
	migrate -path db/migration -database "postgresql://postgres:postgres@localhost:5432/bank?sslmode=disable" -verbose down 1

test:
	go test -v -cover ./...

runserver:
	go run main.go

mock:
	mockgen -destination db/mock/store.go github.com/valkyraycho/bank/db/sqlc Store

protoc:
	rm pb/*.go
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
    --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
    proto/*.proto

evans:
	evans --host localhost --port 8081 -r repl

.PHONY: postgres createdb dropdb migrateup migrateup1 migratedown migratedown1 test runserver mock protoc evans