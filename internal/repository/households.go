package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"time"
)

type HouseHoldRepository interface {
	UpdateRFIDLastSeen(ctx context.Context, rfid string) error
	GetFlatByRFID(ctx context.Context, rfid string) (int, error)
	GetDomophoneByIP(ctx context.Context, ip string) (int, error)
	GetEntrace(ctx context.Context, domophoneId int) (interface{}, error)
}

type Entrace struct {
}

type HouseholdRepositoryImpl struct {
	db     *pgxpool.Pool
	logger *slog.Logger
	parent *PostgresRepository
}

func NewHouseholdRepository(parent *PostgresRepository) *HouseholdRepositoryImpl {
	return &HouseholdRepositoryImpl{
		db:     parent.db,
		logger: parent.logger,
		parent: parent,
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

func (r *HouseholdRepositoryImpl) GetDomophoneByIP(ctx context.Context, ip string) (int, error) {
	r.logger.Debug("Getting domophone by IP", "ip", ip)
	query := `SELECT house_domophone_id FROM houses_domophones WHERE ip = $1`

	var domophoneID int
	err := r.db.QueryRow(ctx, query, ip).Scan(&domophoneID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Warn("Domophone not found", "ip", ip)
			return 0, fmt.Errorf("domophone with IP %s not found", ip)
		}

		r.logger.Error("Database query failed", "error", err, "ip", ip)
		return 0, fmt.Errorf("failed to query domophone: %w", err)
	}

	r.logger.Debug("Domophone found", "domophoneID", domophoneID)

	return domophoneID, nil
}

func (r *HouseholdRepositoryImpl) GetEntrace(ctx context.Context, domophoneId int) (interface{}, error) {
	r.logger.Debug("Getting entrace by domophoneId")

	query := `SELECT * FROM houses_entraces WHERE domophone_id = $1`
	var entrace interface{}
	err := r.db.QueryRow(ctx, query, domophoneId).Scan(&entrace)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {

		}
	}
	r.logger.Debug("Entrace found", "domophoneId", entrace)
	return entrace, nil
}

/**
получить camera streamId
*/
