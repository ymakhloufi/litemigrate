package main

import (
	"errors"
	"fmt"
	"os"
	"syscall"

	"github.com/ymakhloufi/litemigrate/pkg/migrator"
	"github.com/ymakhloufi/litemigrate/pkg/migrator/store"
	"go.uber.org/zap"
)

const defaultMigrationsDir = "./migrations"
const defaultMigrationsTable = "_migrations"

func main() {
	logger, flusher := instantiateLogger()
	defer flusher()

	migrationStore := instantiateStore(logger)

	migrationSvc := migrator.New(logger, migrationStore, getEnv("DIR", defaultMigrationsDir))
	defer migrationSvc.Close()

	err := migrationSvc.Up()
	if err != nil {
		logger.Fatal("failed to run migrations", zap.Error(err))
	}
	logger.Info("migrations completed successfully")
}

func instantiateLogger() (*zap.Logger, func()) {
	logger, err := zap.NewProduction()
	if os.Getenv("ENV") == "local" {
		logger, err = zap.NewDevelopment()
	}
	if err != nil {
		panic(fmt.Sprintf("failed to create logger: %v", err))
	}

	flusher := func() {
		err = logger.Sync()
		if err != nil && !errors.Is(err, syscall.ENOTTY) && !errors.Is(err, syscall.EINVAL) {
			panic(fmt.Sprintf("failed to close logger: %v", err))
		}
	}

	return logger, flusher
}

//nolint:staticcheck
func instantiateStore(logger *zap.Logger) migrator.Store {
	var err error
	migrationsTable := getEnv("TABLE", defaultMigrationsTable)

	var repo migrator.Store
	switch driver := os.Getenv("DRIVER"); driver {
	case "postgres":
		connectionString := store.UserAuthConfigFromEnv().ToConnectionString() // user/pass/... from env
		repo, err = store.NewPostgresStore(logger, migrationsTable, connectionString)
	default:
		logger.Fatal("unknown driver", zap.String("driver", driver))
	}

	if err != nil || repo == nil {
		logger.Fatal("failed to instantiate postgres store", zap.Error(err))
	}

	return repo
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}

	return defaultVal
}
