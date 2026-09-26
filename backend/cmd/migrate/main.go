package main

import (
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/uptaris/uptaris/backend/internal/config"
)

func main() {
	cfg := config.Load()
	m, err := migrate.New("file://migrations", cfg.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer m.Close()
	if len(os.Args) > 1 && os.Args[1] == "down" {
		err = m.Steps(-1)
	} else {
		err = m.Up()
	}
	if err != nil && err != migrate.ErrNoChange {
		panic(err)
	}
	fmt.Println("migrations complete")
}
