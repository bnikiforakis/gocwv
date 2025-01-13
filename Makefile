include .env
export $(shell sed 's/=.*//' .env)

postgres:
	docker run --name cruxdb -p 5432:5432 -e POSTGRES_USER=${DB_USER} -e POSTGRES_PASSWORD=${DB_PASSWORD} -d postgres:17.2-alpine3.21

createdb:
	docker exec -it cruxdb createdb --username=${DB_USER} --owner=${DB_USER} ${DB_NAME}

migrateup:
	migrate -path db/migration -database "postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:5432/${DB_NAME}?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:5432/${DB_NAME}?sslmode=disable" -verbose down

run:
	go run crux.go

dropdb:
	docker exec -it cruxdb dropdb ${DB_NAME}

stopdocker:
	docker stop cruxdb

killdocker:
	docker rm cruxdb

.PHONY: postgres createdb dropdb stopdocker migrateup migratedown run