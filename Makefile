up ::
	docker-compose -f docker-compose.dev.yml up -d

down ::
	docker-compose -f docker-compose.dev.yml down --volumes --remove-orphans

create_db ::
	docker exec -it bookmark.arcbjorn.com_postgres_1 createdb --username=root --owner=root arc_bookmark

create_db_test ::
	docker exec -it bookmark.arcbjorn.com_postgres_1 createdb --username=root --owner=root arc_bookmark_test

drop_db ::
	docker exec -it bookmark.arcbjorn.com_postgres_1 dropdb arc_bookmark

migration ::
	migrate create -ext sql -dir internal/db/migrations -seq $(name)

migrate_up ::
	migrate -path internal/db/migrations --database "postgresql://root:root@localhost:5435/arc_bookmark?sslmode=disable" -verbose up

migrate_up_test ::
	migrate -path internal/db/migrations --database "postgresql://root:root@localhost:5435/arc_bookmark_test?sslmode=disable" -verbose up

migrate_down ::
	migrate -path internal/db/migrations --database "postgresql://root:root@localhost:5435/arc_bookmark?sslmode=disable" -verbose down

generate_orm ::
	sqlc generate

test_orm ::
	go test -v -cover -coverpkg "github.com/arcbjorn/bookmark.arcbjorn.com/internal/db/orm" "github.com/arcbjorn/bookmark.arcbjorn.com/internal/db/orm/tests"

test:
	go test -v ./...