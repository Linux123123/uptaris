package main

import (
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/cockroachdb"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/uptaris/uptaris/backend/internal/config"
)

func main() {
	if len(os.Args) > 1 {
		fmt.Fprintln(os.Stderr, "usage: migrate (forward migrations only; no arguments)")
		os.Exit(2)
	}

	cfg := config.Load()
	m, err := migrate.New("file://migrations", cfg.DatabaseURL)
	if err != nil {
		panic(err)
	}

	defer m.Close()
	err = m.Up()

	if err != nil && err != migrate.ErrNoChange {
		panic(err)
	}

	fmt.Println("migrations complete")
}
