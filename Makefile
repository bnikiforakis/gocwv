postgres:
	docker run --name cruxdb -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=crux -d postgres:17.2-alpine3.21

createdb:
	docker exec -it cruxdb createdb --username=root --owner=root crux

migrateup:
	migrate -path db/migration -database "postgresql://root:crux@localhost:5432/crux?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://root:crux@localhost:5432/crux?sslmode=disable" -verbose down

run:
	go run crux.go

dropdb:
	docker exec -it cruxdb dropdb crux

stopdocker:
	docker stop cruxdb

killdocker:
	docker rm cruxdb




.PHONY: postgres createdb dropdb stopdocker migrateup migratedown run