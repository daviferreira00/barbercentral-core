package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"barbercentral-core/internal/config"
	"barbercentral-core/internal/database"
	"barbercentral-core/internal/server"
)

func main() {
	zerolog.TimeFieldFormat = time.RFC3339

	// Carrega .env se presente
	if err := godotenv.Load(); err != nil {
		log.Info().Msg(".env não encontrado ou não pôde ser carregado; usando variáveis do ambiente")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("falha ao carregar a configuração")
	}

	if !cfg.IsProduction() {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	}

	ctx := context.Background()

	// Aplica as migrações na inicialização
	migrationURL := cfg.Database.MigrationURL()
	log.Info().Str("migration_url", migrationURL).Msg("aplicando migrações no banco...")
	if err := database.RunMigrations(migrationURL, "migrations"); err != nil {
		log.Fatal().Err(err).Msg("falha ao aplicar migrations no banco")
	}

	db, err := database.New(ctx, cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("falha ao conectar no banco de dados")
	}
	defer func() { _ = db.Close() }()
	log.Info().Msg("conexão com o banco de dados estabelecida com sucesso")

	// Instancia servidor HTTP
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      server.New(db, cfg),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		log.Info().
			Str("porta", cfg.Server.Port).
			Str("ambiente", cfg.Server.Environment).
			Msg("servidor BarberCentral iniciado")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("falha ao iniciar o servidor HTTP")
		}
	}()

	// Encerramento gracioso
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("sinal de encerramento recebido, finalizando o servidor...")

	shutdownCtx, cancel := context.WithTimeout(ctx, cfg.Server.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("falha no encerramento gracioso do servidor")
	}
	log.Info().Msg("servidor BarberCentral encerrado")
}
