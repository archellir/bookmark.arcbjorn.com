## Arc bookmark system

### Development

```shell
# run api and postgres containers
make up

# run all migrations
make migrate_up

# regenerate ORM code (just in case)
make generate_orm

# shut down & delete api and postgres containers with volumes
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
make test_orm

# run all tests
make test

# shut down & delete api and postgres containers with volumes
make down
```

Refer to `Makefile` for other commands.
