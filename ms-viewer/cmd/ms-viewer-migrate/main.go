package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/EthernalFox/Controlitix/ms-viewer/migrations"
	"github.com/joho/godotenv"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

const databaseURLEnvironmentKey = "MS_VIEWER_DATABASE_URL"

func main() {
	_ = godotenv.Load()

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := strings.ToLower(strings.TrimSpace(os.Args[1]))
	if runError := runMigrations(command); runError != nil {
		fmt.Fprintf(os.Stderr, "migrations %s failed: %v\n", command, runError)
		os.Exit(1)
	}
}

func runMigrations(command string) error {
	databaseURL := strings.TrimSpace(os.Getenv(databaseURLEnvironmentKey))
	if databaseURL == "" {
		return errors.New("MS_VIEWER_DATABASE_URL is required")
	}

	databaseConnection, openError := sql.Open("pgx", databaseURL)
	if openError != nil {
		return fmt.Errorf("open postgres connection: %w", openError)
	}
	defer databaseConnection.Close()

	databaseConnection.SetMaxOpenConns(1)
	databaseConnection.SetMaxIdleConns(1)

	if pingError := databaseConnection.Ping(); pingError != nil {
		return fmt.Errorf("ping postgres connection: %w", pingError)
	}

	if dialectError := goose.SetDialect("postgres"); dialectError != nil {
		return fmt.Errorf("set goose dialect: %w", dialectError)
	}
	goose.SetBaseFS(migrations.Files)

	switch command {
	case "up":
		if upError := goose.Up(databaseConnection, migrations.Directory); upError != nil {
			return fmt.Errorf("goose up: %w", upError)
		}
	case "down":
		if downError := goose.DownTo(databaseConnection, migrations.Directory, 0); downError != nil {
			return fmt.Errorf("goose down to base: %w", downError)
		}
	case "status":
		if statusError := goose.Status(databaseConnection, migrations.Directory); statusError != nil {
			return fmt.Errorf("goose status: %w", statusError)
		}
	case "version":
		version, versionError := goose.GetDBVersion(databaseConnection)
		if versionError != nil {
			return fmt.Errorf("goose get version: %w", versionError)
		}
		fmt.Println(version)
	default:
		printUsage()
		return fmt.Errorf("unknown command: %s", command)
	}

	return nil
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  ms-viewer-migrate up")
	fmt.Println("  ms-viewer-migrate down")
	fmt.Println("  ms-viewer-migrate status")
	fmt.Println("  ms-viewer-migrate version")
}
