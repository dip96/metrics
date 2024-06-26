package migrator

import (
	"github.com/dip96/metrics/internal/config"
	"github.com/dip96/metrics/internal/database/migrator/driver"
	"github.com/golang-migrate/migrate/v4"
	"github.com/pkg/errors"
	"sync"
)

var (
	instance *Migrator
)

type Migrator struct {
	migrate *migrate.Migrate
}

// NewMigrator создает новый объект Migrator для управления миграциями базы данных.
func NewMigrator() (*Migrator, error) {
	var err error

	//TODO https://habr.com/ru/articles/553298/ ???
	funcOnce := sync.OnceFunc(func() {
		instance, err = newMigrator()
	})

	funcOnce()

	return instance, err
}
func newMigrator() (*Migrator, error) {
	cnf, err := config.LoadServer()

	if err != nil {
		return nil, err
	}

	driver.InitFile()
	driver.InitPostgres()
	m, err := migrate.New(cnf.MigrationPath, cnf.DatabaseDsn)
	if err != nil {
		err = errors.Wrap(err, "error creating the instance \"Migration\"")
		return nil, err
	}

	return &Migrator{migrate: m}, nil
}

// Up применяет миграции базы данных.
func (m *Migrator) Up() error {
	if err := m.migrate.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return nil
		}

		if errors.Is(err, migrate.ErrNilVersion) {
			return nil
		}

		return errors.Wrap(err, "error migrating up")
	}

	return nil
}

func (m *Migrator) Down() error {
	if err := m.migrate.Down(); err != nil {
		return errors.Wrap(err, "error migrating down")
	}

	return nil
}
