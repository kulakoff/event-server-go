package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/kulakoff/event-server-go/internal/app/event-server-go/repository/models"
	"log/slog"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HouseHoldRepository interface {
	UpdateRFIDLastSeen(ctx context.Context, rfid string) error
	GetFlatByRFID(ctx context.Context, rfid string) (int, error)
	GetDomophoneIDByIP(ctx context.Context, ip string) (int, error)
	GetEntrance(ctx context.Context, domophoneId, output int) (*models.HouseEntrance, error)
	GetDomophone(ctx context.Context, by, p string) (*models.Domophone, error)
	GetFlatIDsByRFID(ctx context.Context, rfid string) ([]int, error)
	GetFlatIDsByCode(ctx context.Context, code string) ([]int, error)
	GetFlatsByFaceIdFrs(ctx context.Context, faceId string, entranceId string) ([]int, error)
	GetFlatIDByApartment(ctx context.Context, apartment int, domophoneId int) (int, error)
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

func (r *HouseholdRepositoryImpl) GetEntrance(ctx context.Context, domophoneId, output int) (*models.HouseEntrance, error) {
	r.logger.Debug("Getting entrance by domophoneId")
	query := `
			SELECT
			    address_house_id,
			    prefix,
				house_entrance_id, 
				entrance_type, 
				entrance, 
				lat, 
				lon,
				shared, 
				plog, 
				caller_id, 
				camera_id, 
				house_domophone_id,
				domophone_output, 
				cms, 
				cms_type, 
				coalesce(cms_levels, '') as	cms_levels,
				path,
				distance, 
				alt_camera_id_1, 
				alt_camera_id_2, 
				alt_camera_id_3,
				alt_camera_id_4, 
				alt_camera_id_5, 
				alt_camera_id_6, 
				alt_camera_id_7
			FROM houses_entrances
			LEFT JOIN houses_houses_entrances USING (house_entrance_id)
			WHERE house_domophone_id = $1 AND domophone_output = $2
			ORDER BY entrance_type, entrance
`
	var entrance models.HouseEntrance
	err := r.db.QueryRow(ctx, query, domophoneId, output).Scan(
		&entrance.AddressHouseID,
		&entrance.Path,
		&entrance.HouseEntranceID,
		&entrance.EntranceType,
		&entrance.Entrance,
		&entrance.Lat,
		&entrance.Lon,
		&entrance.Shared,
		&entrance.Plog,
		&entrance.CallerID,
		&entrance.CameraID,
		&entrance.HouseDomophoneID,
		&entrance.DomophoneOutput,
		&entrance.CMS,
		&entrance.CMSType,
		&entrance.CMSLevels,
		&entrance.Path,
		&entrance.Distance,
		&entrance.AltCameraID1,
		&entrance.AltCameraID2,
		&entrance.AltCameraID3,
		&entrance.AltCameraID4,
		&entrance.AltCameraID5,
		&entrance.AltCameraID6,
		&entrance.AltCameraID7,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Warn("Entrace not found", "domophone_id", domophoneId)
			return nil, fmt.Errorf("domophone with ID %s not found", domophoneId)
		}
		r.logger.Error("Database query failed", "error", err, "domophone_id", domophoneId)
		return nil, fmt.Errorf("failed to get entrance: %w", err)
	}

	r.logger.Debug("Entrace found", "domophoneID", domophoneId, "entraceId", entrance.HouseEntranceID)
	return &entrance, nil
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

func (r *HouseholdRepositoryImpl) GetFlatIDsByRFID(ctx context.Context, rfid string) ([]int, error) {
	r.logger.Debug("GetFlatsByRFID RUN >")
	query := `
		SELECT house_flat_id 
		FROM houses_flats 
		WHERE house_flat_id IN (
		    SELECT access_to
		    FROM houses_rfids
		    WHERE access_type = 2 AND rfid = $1
		)
		GROUP BY house_flat_id
	`

	r.logger.Debug("GetFlatsByRFID query", "query", query, "rfid", rfid)

	rows, err := r.db.Query(ctx, query, rfid)
	if err != nil {
		r.logger.Error("Query executing failed", "error", err)
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var flatIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			r.logger.Error("Failed to scan house_flat_id", "error", err)
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		flatIDs = append(flatIDs, id)
	}

	if len(flatIDs) == 0 {
		r.logger.Debug("No flats found for RFID", "rfid", rfid)
		return nil, nil
	}

	return flatIDs, nil
}

func (r *HouseholdRepositoryImpl) GetFlatIDsByCode(ctx context.Context, code string) ([]int, error) {
	r.logger.Debug("GetFlatIDsByCode RUN >")
	query := `
		SELECT house_flat_id 
		FROM houses_flats
		WHERE open_code = $1
		GROUP BY house_flat_id
	`
	r.logger.Debug("GetFlatIDsByCode query", "query", query, "rfid", code)

	rows, err := r.db.Query(ctx, query, code)
	if err != nil {
		r.logger.Error("Query executing failed", "error", err)
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var flatIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			r.logger.Error("Failed to scan house_flat_id", "error", err)
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		flatIDs = append(flatIDs, id)
	}

	if len(flatIDs) == 0 {
		r.logger.Debug("No flats found for CODE", "code", code)
		return nil, nil
	}

	return flatIDs, nil
}

func (r *HouseholdRepositoryImpl) GetFlatsByFaceIdFrs(ctx context.Context, faceId string, entranceId string) ([]int, error) {
	r.logger.Debug("GetFlatsByFaceIdFrs RUN >")
	// TODO: implement me

	query := `
		SELECT
			flf.flat_id
		FROM
			houses_entrances_flats hef
			INNER JOIN frs_links_faces flf
			ON hef.house_flat_id = flf.flat_id
		WHERE 
			hef.house_entrance_id = $1
			AND flf.face_id = $2`

	r.logger.Debug("GetFlatsByFaceIdFrs query", "query", query, "faceId", faceId, "entranceId", entranceId)
	rows, err := r.db.Query(ctx, query, entranceId, faceId)
	if err != nil {
		r.logger.Error("Query executing failed", "error", err)
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var flatIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			r.logger.Error("Failed to scan house_flat_id", "error", err)
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		flatIDs = append(flatIDs, id)
	}
	if len(flatIDs) == 0 {
		r.logger.Debug("No flats found for faceId", "faceId", faceId)
		return nil, nil
	}

	return flatIDs, nil
}

// TODO: implement me
func (r *HouseholdRepositoryImpl) GetFlatIDByApartment(ctx context.Context, apartment, domophoneId int) (int, error) {
	query := `
		SELECT
			house_flat_id
		FROM
			houses_entrances_flats
		WHERE
			apartment = :apartment
			AND
			house_entrance_id in (
				SELECT 
					house_entrance_id
				FROM 
					houses_entrances
				WHERE 
					house_domophone_id = :house_domophone_id
			)
		GROUP BY 
			house_flat_id
			`

	return 0, nil
}

/**
получить camera streamId
*/
