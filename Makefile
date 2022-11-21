# Docker

up ::
	docker-compose -f docker-compose.dev.yml up -d

down ::
	docker-compose -f docker-compose.dev.yml down --volumes --remove-orphans

# Database

create_db ::
	docker exec -it bookmarkarcbjorncom_postgres_1 createdb --username=root --owner=root arc_bookmark

create_db_test ::
	docker exec -it bookmarkarcbjorncom_postgres_1 createdb --username=root --owner=root arc_bookmark_test

drop_db ::
	docker exec -it bookmarkarcbjorncom_postgres_1 dropdb arc_bookmark

migration ::
	migrate create -ext sql -dir internal/db/migrations -seq $(name)

migrate_up ::
	migrate -path internal/db/migrations --database "postgresql://root:root@localhost:5435/arc_bookmark?sslmode=disable" -verbose up

migrate_up_test ::
	migrate -path internal/db/migrations --database "postgresql://root:root@localhost:5435/arc_bookmark_test?sslmode=disable" -verbose up

migrate_down ::
	migrate -path internal/db/migrations --database "postgresql://root:root@localhost:5435/arc_bookmark?sslmode=disable" -verbose down

# Code gen

generate_orm ::
	sqlc generate

# Testing

test_backend_orm ::
	go test -v -cover -coverpkg "github.com/arcbjorn/bookmark.arcbjorn.com/internal/db/orm" "github.com/arcbjorn/bookmark.arcbjorn.com/internal/db/orm/tests"

test_backend:
	go test -v ./...

test_frontend_unit:
	pnpm --prefix ./web test:unit

test_frontend_e2e:
	pnpm --prefix ./web test:e2e

# Development

dev_backend:
	go run cmd/main.go

dev_frontend:
	pnpm --prefix ./web dev

dev_full:
	pnpm --prefix ./web build && go run cmd/main.go

# Production

prod:
	pnpm --prefix ./web build && go build cmd/main.go