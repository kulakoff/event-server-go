package repository

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"time"
)

type PostgresRepository struct {
	db     *pgxpool.Pool
	logger *slog.Logger
}

func NewPostgresRepository(db *pgxpool.Pool, logger *slog.Logger) (*PostgresRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("db pool is nil")
	}

	if err := db.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("could not connect to database: %w", err)
	}

	logger.Info("postgres repository initialized")

	return &PostgresRepository{
		db:     db,
		logger: logger,
	}, nil
}

func (r *PostgresRepository) UpdateRFIDLastSeen(ctx context.Context, rfid string) error {
	lastSeen := time.Now().Unix()

	query := `UPDATE houses_rfids SET last_seen = $1 WHERE rfid = $2`
	result, err := r.db.Exec(ctx, query, lastSeen, rfid)
	if err != nil {
		return fmt.Errorf("failed to update last_seen for RFID %s: %w", rfid, err)
	}

	rowsAffected := result.RowsAffected()
	r.logger.Debug("Updated RFID last_seen", "rfid", rfid, "last_seen", lastSeen, "rowsAffected", rowsAffected)

	return nil
}
