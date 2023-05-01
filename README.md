# Lite Migrate

### What it does

This tool allows you to run DB schema migrations against a Postgres database. The tool keeps track of the migrations
that were already run, and ensures that new migrations are run in order. If a migration fails to run, the process stops
with an error and the tool will prevent future runs until the failed migration was cleaned up manually.

### How it works

Lite Migrate will create a new table (`_migrations`) in your database. In there, it keeps track of all migrations it
runs and whether they have been completed successfully. The migration files (located in the folder `./migrations`)

### Limitations

- Lite Migrate currently only supports Postgres Databases
- Migrations are NOT wrapped into transactions that can be rolled back. Not everything can be wrapped into a
  transaction, so it is in the user's responsibility touse transactions in their migration files.

### Config options (Set as ENV variables)

| Option | Description                                                                                                                                                       | Default        |
|--------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------|----------------|
| ENV    | Determines whether to log in JSON or human readable format. Possible values: `local`, `production`                                                                | `production`   |
| TABLE  | Name of the table that will be created to keep track of migrations                                                                                                | `_migrations`  |
| DIR    | Path to the folder that contains the migration files. If you run it inside a container, you need to mount your migrations files into the same path (see example). | `./migrations` |
| DRIVER | Database driver to use. Currently only `postgres` is supported.                                                                                                   | -none-         |
| HOST   | Hostname of the database server                                                                                                                                   | -none-         |
| PORT   | Port of the database server                                                                                                                                       | -none-         |
| USER   | Username to use for connecting to the database                                                                                                                    | -none-         |
| PASS   | Password to use for connecting to the database                                                                                                                    | -none-         |
| DB     | Name of the database to connect to                                                                                                                                | -none-         |

### How to use

#### Option 1: Docker

Full Example, assuming you have a postgres server running locally on port 5432) and your migrations are located in
`./my_migrations_folder` (relative to the current working directory):

```sh
 

docker run \
  -e ENV=local \
  -e TABLE="_migrations" \
  -e DIR="/migrations" \
  -e DRIVER="postgres" \
  -e HOST="localhost" \
  -e PORT=5432 \
  -e USER="my_user" \
  -e PASS="my_password" \
  -e DB="my_db" \
  --net=host \
  -v "`pwd`/my_migrations_folder:/migrations" \
  y11a/litemigrate:1
```

```Set DIR to the same path as to where you mount the migrations folder:
if      DIR="/folder_inside_container"
then    -v "`pwd`/my_local_migrations_folder:/folder_inside_container"
```

#### Option 2: As a Go library

```go
package main

import (
	"github.com/ymakhloufi/litemigrate/pkg/migrator"
	"github.com/ymakhloufi/litemigrate/pkg/migrator/store"
	"go.uber.org/zap"

	"log"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}

	pathToMigrationFiles := "/tmp/migrations"
	migrationsTableName := "_migrations"
	connectionString := "postgres://user:password@localhost:5432/my_db?sslmode=disable"

	pgstore, err := store.NewPostgresStore(logger, migrationsTableName, connectionString)
	migrator, err := migrator.New(logger, pgstore, pathToMigrationFiles)
	defer func() {
		_, err := migrator.Close()
		if err != nil {
			logger.Error("error closing migrator", zap.Error(err))
		}
	}

	err = migrator.Up()
	if err != nil {
		logger.Error("error running migrations", zap.Error(err))
	}
}
```

### The Migration table will look like this:

------------------

| id | filename                          | started_at                 | completed_at               |
|----|-----------------------------------|----------------------------|----------------------------|
| 1  | 0001_create_some_table.sql        | 2021-01-01 01:23:34.123456 | 2021-01-01 01:23:35.123456 |
| 2  | 0002_alter_some_table.sql         | 2021-01-02 01:23:34.123456 | 2021-01-02 01:23:35.123456 |
| 3  | 0003_failing_schema_migration.sql | 2021-01-03 01:23:34.123456 | NULL                       |

The NULL in the completed_at column will result in future runs not executing and exiting with a non-zero exit status.
You will have to login to your DB and fix it by hand once you made sure that the botched migration didn't cause any
issues. Either set the `completed_at` timestamp to womething non-null, or delete the row altogether (depending on
whether you want to re-run the migration or not).
