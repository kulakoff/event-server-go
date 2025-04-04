package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kulakoff/event-server-go/internal/repository/models"
	"log/slog"
	"strconv"
	"time"
)

type HouseHoldRepository interface {
	UpdateRFIDLastSeen(ctx context.Context, rfid string) error
	GetFlatByRFID(ctx context.Context, rfid string) (int, error)
	GetDomophoneIDByIP(ctx context.Context, ip string) (int, error)
	GetEntrace(ctx context.Context, domophoneId int) (*models.HouseEntrance, error)
	GetDomophone(ctx context.Context, by string, ip string) (*models.Domophone, error)
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

func (r *HouseholdRepositoryImpl) GetDomophoneIDByIP(ctx context.Context, ip string) (int, error) {
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

func (r *HouseholdRepositoryImpl) GetEntrace(ctx context.Context, domophoneId int) (*models.HouseEntrance, error) {
	r.logger.Debug("Getting entrace by domophoneId")
	query := `
			SELECT
				house_entrance_id, entrance_type, entrance, lat, lon,
				shared, plog, caller_id, camera_id, house_domophone_id,
				domophone_output, cms, cms_type, cms_levels, path,
				distance, alt_camera_id_1, alt_camera_id_2, alt_camera_id_3,
				alt_camera_id_4, alt_camera_id_5, alt_camera_id_6, alt_camera_id_7
			FROM houses_entrances
			WHERE house_domophone_id = $1
`
	var entrace models.HouseEntrance
	err := r.db.QueryRow(ctx, query, domophoneId).Scan(
		&entrace.HouseEntranceID,
		&entrace.EntranceType,
		&entrace.Entrance,
		&entrace.Lat,
		&entrace.Lon,
		&entrace.Shared,
		&entrace.Plog,
		&entrace.CallerID,
		&entrace.CameraID,
		&entrace.HouseDomophoneID,
		&entrace.DomophoneOutput,
		&entrace.CMS,
		&entrace.CMSType,
		&entrace.CMSLevels,
		&entrace.Path,
		&entrace.Distance,
		&entrace.AltCameraID1,
		&entrace.AltCameraID2,
		&entrace.AltCameraID3,
		&entrace.AltCameraID4,
		&entrace.AltCameraID5,
		&entrace.AltCameraID6,
		&entrace.AltCameraID7,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Warn("Entrace not found", "domophone_id", domophoneId)
			return nil, fmt.Errorf("domophone with ID %s not found", domophoneId)
		}
		r.logger.Error("Database query failed", "error", err, "domophone_id", domophoneId)
		return nil, fmt.Errorf("failed to get entrance: %w", err)
	}

	// process errors v2
	//switch {
	//case errors.Is(err, pgx.ErrNoRows):
	//	r.logger.Warn("Entrace not found", "domophone_id", domophoneId)
	//	return nil, fmt.Errorf("domophone with ID %s not found", domophoneId)
	//case err != nil:
	//	r.logger.Error("Database query failed",
	//		"error", err,
	//		"domophone_id", domophoneId)
	//	return nil, fmt.Errorf("failed to get entrance: %w", err)
	//}

	r.logger.Debug("Entrace found", "domophoneID", domophoneId, "entraceId", entrace.HouseEntranceID)
	return &entrace, nil
}

func (r *HouseholdRepositoryImpl) GetDomophone(ctx context.Context, by string, param string) (*models.Domophone, error) {
	r.logger.Debug("Getting domophone", "by", by, "param", param)

	var query string
	var queryParam interface{}

	switch by {
	case "id":
		id, err := strconv.Atoi(param)
		if err != nil {
			r.logger.Error("Invalid id format", "param", param, "error", err)
			return nil, fmt.Errorf("invalid id format: %s must be an integer", param)
		}
		query = `
            SELECT house_domophone_id, enabled, model, server, url, credentials, dtmf, first_time, nat, 
                   locks_are_open, ip, sub_id, name, comments, display, video
            FROM houses_domophones 
            WHERE house_domophone_id = $1
        `
		queryParam = id
	case "ip":
		query = `
            SELECT house_domophone_id, enabled, model, server, url, credentials, dtmf, first_time, nat, 
                   locks_are_open, ip, sub_id, name, comments, display, video
            FROM houses_domophones 
            WHERE ip = $1
        `
		queryParam = param
	default:
		r.logger.Error("Invalid search type", "by", by)
		return nil, fmt.Errorf("invalid search type: %s; must be 'id' or 'ip'", by)
	}

	var domophone models.Domophone
	err := r.db.QueryRow(ctx, query, queryParam).Scan(
		&domophone.HouseDomophoneID,
		&domophone.Enabled,
		&domophone.Model,
		&domophone.Server,
		&domophone.URL,
		&domophone.Credentials,
		&domophone.DTMF,
		&domophone.FirstTime,
		&domophone.NAT,
		&domophone.LocksAreOpen,
		&domophone.IP,
		&domophone.SubID,
		&domophone.Name,
		&domophone.Comments,
		&domophone.Display,
		&domophone.Video,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Warn("Domophone not found", "by", by, "param", param)
			return nil, fmt.Errorf("domophone with %s %s not found", by, param)
		}
		r.logger.Error("Database query failed", "error", err, "by", by, "param", param)
		return nil, fmt.Errorf("failed to query domophone: %w", err)
	}

	r.logger.Debug("Domophone found", "house_domophone_id", domophone.HouseDomophoneID)
	return &domophone, nil
}

/**
получить camera streamId
*/
