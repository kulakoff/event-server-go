package repository

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"time"
)

type HouseHoldRepository interface {
	UpdateRFIDLastSeen(ctx context.Context, rfid string) error
	GetFlatByRFID(ctx context.Context, rfid string) (int, error)
}

type HouseholdRepositoryImpl struct {
	db     *pgxpool.Pool
	logger *slog.Logger
}

func NewHouseholdRepository(db *pgxpool.Pool, logger *slog.Logger) *HouseholdRepositoryImpl {
	return &HouseholdRepositoryImpl{
		db:     db,
		logger: logger,
	}
}

// UpdateRFIDLastSeen update last usage key
func (r *HouseholdRepositoryImpl) UpdateRFIDLastSeen(ctx context.Context, rfid string) error {
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

func (r *HouseholdRepositoryImpl) GetFlatByRFID(ctx context.Context, rfid string) (int, error) {
	// TODO implement me
	return 0, nil
}
