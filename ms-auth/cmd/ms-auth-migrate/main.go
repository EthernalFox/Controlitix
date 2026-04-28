package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/EthernalFox/Controlitix/ms-auth/migrations"
	"github.com/joho/godotenv"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

const (
	databaseURLEnvironmentKey = "MS_AUTH_DATABASE_URL"
)

func main() {
	_ = godotenv.Load()

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := strings.ToLower(strings.TrimSpace(os.Args[1]))
	switch command {
	case "hash-password", "--hash-password":
		hash, hashError := hashPasswordFromStdin()
		if hashError != nil {
			fmt.Fprintf(os.Stderr, "hash password: %v\n", hashError)
			os.Exit(1)
		}
		fmt.Println(hash)
	case "up", "down", "status":
		if runError := runMigrations(command); runError != nil {
			fmt.Fprintf(os.Stderr, "migrations %s failed: %v\n", command, runError)
			os.Exit(1)
		}
	default:
		printUsage()
		os.Exit(1)
	}
}

func runMigrations(command string) error {
	databaseURL := strings.TrimSpace(os.Getenv(databaseURLEnvironmentKey))
	if databaseURL == "" {
		return errors.New("MS_AUTH_DATABASE_URL is required")
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

	if command == "up" {
		if setError := setInitialAdminSessionVariables(databaseConnection); setError != nil {
			return setError
		}
	}

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
	default:
		return fmt.Errorf("unknown command: %s", command)
	}

	return nil
}

func setInitialAdminSessionVariables(databaseConnection *sql.DB) error {
	settings := map[string]string{
		"ms_auth.initial_admin_username":      os.Getenv("MS_AUTH_INITIAL_ADMIN_USERNAME"),
		"ms_auth.initial_admin_password_hash": os.Getenv("MS_AUTH_INITIAL_ADMIN_PASSWORD_HASH"),
		"ms_auth.initial_admin_email":         os.Getenv("MS_AUTH_INITIAL_ADMIN_EMAIL"),
	}

	for settingKey, settingValue := range settings {
		if _, setError := databaseConnection.Exec(
			"SELECT set_config($1, $2, false)",
			settingKey,
			settingValue,
		); setError != nil {
			// Session-level GUC values are reused by goose on the same connection (SetMaxOpenConns(1)).
			return fmt.Errorf("set %s: %w", settingKey, setError)
		}
	}

	return nil
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  ms-auth-migrate up")
	fmt.Println("  ms-auth-migrate down")
	fmt.Println("  ms-auth-migrate status")
	fmt.Println("  ms-auth-migrate hash-password")
}
