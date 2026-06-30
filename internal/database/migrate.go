package database

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/rs/zerolog/log"
)

func RunMigrations(databaseURL, migrationsPath string) error {
	m, err := migrate.New("file://"+migrationsPath, databaseURL)
	if err != nil {
		return fmt.Errorf("falha ao iniciar o migrate: %w", err)
	}
	defer func() {
		sourceErr, dbErr := m.Close()
		if sourceErr != nil {
			log.Warn().Err(sourceErr).Msg("erro ao fechar source de migrations")
		}
		if dbErr != nil {
			log.Warn().Err(dbErr).Msg("erro ao fechar conexão de migrations")
		}
	}()

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Info().Msg("nenhuma migration pendente")
			return nil
		}
		return fmt.Errorf("falha ao aplicar migrations: %w", err)
	}

	log.Info().Msg("migrations aplicadas com sucesso")
	return nil
}
