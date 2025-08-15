package migrate

import (
	"fmt"
	"multi-tenant-service/package/config"
	"path/filepath"

	"github.com/urfave/cli/v2"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Migrate struct {
	conf *config.Config
}

func (h *Migrate) Migrate(c *cli.Context) error {
	migrationsPath, err := filepath.Abs("migrations")
	if err != nil {
		fmt.Println("error", err)
	}
	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		h.conf.Database.URL,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

func NewMigrate(conf *config.Config) []*cli.Command {
	h := Migrate{
		conf: conf,
	}
	return []*cli.Command{
		{
			Name:   "migrate",
			Usage:  "Migrate database",
			Action: h.Migrate,
		},
	}
}