package repository

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"time"
)

type Stream struct {
	ID     int
	UrlDVR string
	UrlFRS string
}

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

// UpdateRFIDLastSeen update last usage key
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

func (r *PostgresRepository) GetStreamByIP(ctx context.Context, ip string) (*Stream, error) {
	// TODO implement me
	// FIXME: stub stream
	if ip == "192.168.13.152" {
		return &Stream{
			ID:     8,
			UrlDVR: "https://dvr-example.com/stream-name/index.m3u8",
			UrlFRS: "http://localhost:9051",
		}, nil
	}
	if ip == "192.168.88.25" {
		return &Stream{
			ID:     8,
			UrlDVR: "https://dvr-example.com/stream-name/index.m3u8",
			UrlFRS: "https://webhook.site/5c40e512-64c6-49b8-96d0-d6d028f8181f",
		}, nil
	}

	return nil, nil
}

func (r *PostgresRepository) GetFlatByRFID(ctx context.Context, rfid string) (int, error) {
	// TODO implement me
	return 0, nil
}
