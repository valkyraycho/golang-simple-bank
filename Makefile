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
	mockgen -package mockdb -destination db/mock/store.go github.com/valkyraycho/bank/db/sqlc Store
	mockgen -package mockwk -destination worker/mock/distributor.go github.com/valkyraycho/bank/worker TaskDistributor

protoc:
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
    --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=pb --grpc-gateway_opt=paths=source_relative \
	--openapiv2_out=docs/swagger --openapiv2_opt=allow_merge=true,merge_file_name=simple_bank \
    proto/*.proto
	statik -src=./docs/swagger -dest=./docs

evans:
	evans --host localhost --port 8081 -r repl

redis:
	docker run --name redis -p 6379:6379 -d redis:7-alpine

.PHONY: postgres createdb dropdb migrateup migrateup1 migratedown migratedown1 test runserver mock protoc evans redis