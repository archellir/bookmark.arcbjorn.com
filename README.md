## Arc bookmark system

### Development

```shell
# run app (currently disabled) and postgres containers
make up

# run all migrations
make migrate_up

# regenerate ORM code (just in case)
make generate_orm

# run backend app in dev mode - http://localhost:8080/
make dev_backend

# run frontend app in dev mode (with hot reloading) - http://127.0.0.1:5173/
make dev_frontend

# run full app in dev mode (no hot reloading) - http://localhost:8080/
make dev_full

---

# shut down & delete app (currently disabled) and postgres containers with volumes
make down
```

### Testing

```shell
# run api and postgres containers
make up

# create test database
make create_db_test

# run all migrations in test database
make migrate_up_test

# regenerate ORM code (just in case)
make generate_orm

# run ORM tests
make test_backend_orm

# run all backend tests
make test_backend

# run frontend unit tests
make test_frontend_unit

# run frontend end-to-end tests
make test_frontend_e2e

# shut down & delete api and postgres containers with volumes
make down
```

### Production

```sh
# build binary (no docker)
make prod

# run binary
./main
```

Refer to `Makefile` for other commands.
