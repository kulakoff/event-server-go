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
	GetFlatByID(ctx context.Context, flatID int) (models.Flat, error)
	GetFlatIDByApartment(ctx context.Context, apartment int, domophoneId int) (int, error)
	GetWatchersByFlatID(ctx context.Context, flatID int) ([]models.Watcher, error)
	GetRFID(ctx context.Context, rfid string) ([]models.RFID, error)
	GetSubscriberIDByFlatIDandPhone(ctx context.Context, flatID int, phone string) (int, error)
	FlatIDsByDomophoneIDAndPhone(ctx context.Context, domophoneID int, phone string) ([]int, error)
	GetMobileDeviceByID(ctx context.Context, deviceID int) (models.MobileDevice, error)
	GetHouseByEntranceID(ctx context.Context, entranceID int) (models.House, error)
	GetFlatIDsByRFID_new(ctx context.Context, rfid string) ([]models.Flat, error)
	GetFlatIDsByCode_new(ctx context.Context, code string) ([]models.Flat, error)
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

func (r *HouseholdRepositoryImpl) GetFlatIDsByRFID_new(ctx context.Context, rfid string) ([]models.Flat, error) {
	r.logger.Debug("GetFlatsByRFID RUN >")
	query := `
		SELECT 
		    house_flat_id, 
		    address_house_id,
		    flat
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

	var flats []models.Flat
	for rows.Next() {
		var flat models.Flat
		if err := rows.Scan(&flat.HouseFlatID, &flat.AddressHouseID, &flat.Flat); err != nil {
			r.logger.Error("Failed to scan house_flat_id", "error", err)
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		flats = append(flats, flat)
	}

	if len(flats) == 0 {
		r.logger.Debug("No flats found for RFID", "rfid", rfid)
		return nil, nil
	}

	return flats, nil
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

func (r *HouseholdRepositoryImpl) GetFlatIDsByCode_new(ctx context.Context, code string) ([]models.Flat, error) {
	r.logger.Debug("GetFlatIDsByCode RUN >")
	query := `
		SELECT house_flat_id,
		       address_house_id,
		       flat
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

	var flats []models.Flat
	for rows.Next() {
		var flat models.Flat
		if err := rows.Scan(&flat.HouseFlatID, &flat.AddressHouseID, &flat.Flat); err != nil {
			r.logger.Error("Failed to scan house_flat_id", "error", err)
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		flats = append(flats, flat)
	}

	if len(flats) == 0 {
		r.logger.Debug("No flats found for CODE", "code", code)
		return nil, nil
	}

	return flats, nil
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

func (r *HouseholdRepositoryImpl) GetFlatByID(ctx context.Context, flatID int) (models.Flat, error) {
	r.logger.Debug("GetFlatsByFaceIdFrs RUN >")
	// TODO: implement me

	query := `
		SELECT
		    house_flat_id,
		    address_house_id,
		    floor,
		    flat,
		    code,
		    plog,
		    manual_block,
		    auto_block,
		    admin_block,
		    open_Code,
		    auto_open,
		    white_rabbit,
		    sip_enabled,
		    sip_password,
		    last_opened,
		    cms_enabled,
		    contract,
		    login,
		    password,
		    cars,
		    subscribers_limit
		FROM
			houses_flats
		WHERE 
			house_flat_id = $1`

	var flat models.Flat
	err := r.db.QueryRow(ctx, query, flatID).Scan(
		&flat.HouseFlatID,
		&flat.AddressHouseID,
		&flat.Floor,
		&flat.Flat,
		&flat.Code,
		&flat.Plog,
		&flat.ManualBlock,
		&flat.AutoBlock,
		&flat.AdminBlock,
		&flat.OpenCode,
		&flat.AutoOpen,
		&flat.WhiteRabbit,
		&flat.SipEnabled,
		&flat.SipPassword,
		&flat.LastOpened,
		&flat.CmsEnabled,
		&flat.Contract,
		&flat.Login,
		&flat.Password,
		&flat.Cars,
		&flat.SubscribersLimit,
	)
	if err != nil {
		r.logger.Error("Failed to scan house_flat_id", "error", err)
		return models.Flat{}, fmt.Errorf("query failed: %w", err)
	}

	return flat, nil
}

// TODO: implement me
func (r *HouseholdRepositoryImpl) GetFlatIDByApartment(ctx context.Context, apartment, domophoneId int) (int, error) {
	query := `
		SELECT
			house_flat_id
		FROM
			houses_entrances_flats
		WHERE
			apartment = $1
			AND
			house_entrance_id in (
				SELECT 
					house_entrance_id
				FROM 
					houses_entrances
				WHERE 
					house_domophone_id = $2
			)
		GROUP BY 
			house_flat_id
			`
	var flatID int
	err := r.db.QueryRow(ctx, query, apartment, domophoneId).Scan(&flatID)
	if err != nil {
		r.logger.Error("Query executing failed", "error", err)
		return 0, fmt.Errorf("query failed: %w", err)
	}
	return flatID, nil
}

// TODO: implement method
func (r *HouseholdRepositoryImpl) GetWatchersByFlatID(ctx context.Context, flatID int) ([]models.Watcher, error) {
	query := `
		SELECT
			house_watcher_id,
			subscriber_device_id,
			house_flat_id,
			event_type,
			event_detail,
			comments
		FROM houses_watchers
		WHERE house_flat_id = $1
		ORDER BY house_watcher_id
	`

	r.logger.Debug("GetWatchersByFlatID query", "query", query, "flatID", flatID)

	rows, err := r.db.Query(ctx, query, flatID)
	if err != nil {
		r.logger.Error("Query executing failed", "error", err)
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var watchers []models.Watcher
	for rows.Next() {
		var watcher models.Watcher
		if err := rows.Scan(
			&watcher.WatcherID,
			&watcher.DeviceID,
			&watcher.FlatID,
			&watcher.EventType,
			&watcher.EventDetail,
			&watcher.Comments,
		); err != nil {
			r.logger.Error("Failed to scan house_watchers", "error", err)
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		watchers = append(watchers, watcher)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Failed to scan house_watchers", "error", err)
		return nil, fmt.Errorf("failed to scan house_watchers: %w", err)
	}

	if len(watchers) == 0 {
		r.logger.Debug("No events found for watchers", "flatID", flatID)
		return []models.Watcher{}, nil
	}

	r.logger.Debug("Watchers found", "count", len(watchers), "flatID", flatID)
	return watchers, nil
}

func (r *HouseholdRepositoryImpl) GetDeviceByID(ctx context.Context, deviceID int) {

}

func (r *HouseholdRepositoryImpl) GetRFID(ctx context.Context, rfid string) ([]models.RFID, error) {
	r.logger.Debug("GetRFID query", "rfid", rfid)
	query := `
		SELECT house_rfid_id, rfid, access_type, access_to, last_seen, comments FROM houses_rfids WHERE rfid = $1`
	r.logger.Debug("GetRFID query", "query", query, "rfid", rfid)
	rows, err := r.db.Query(ctx, query, rfid)
	if err != nil {
		r.logger.Error("Query executing failed", "error", err)
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var rfids []models.RFID
	for rows.Next() {
		var rfid models.RFID
		if err := rows.Scan(
			&rfid.HouseRfidId,
			&rfid.RFID,
			&rfid.AccessType,
			&rfid.AccessTo,
			&rfid.LastSeen,
			&rfid.Comments); err != nil {
			r.logger.Error("Failed to scan rfid", "error", err)
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		rfids = append(rfids, rfid)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Failed to scan rfid", "error", err)
		return nil, fmt.Errorf("failed to scan rfid: %w", err)
	}

	if len(rfids) == 0 {
		r.logger.Debug("No rfids found for rfid", "rfid", rfid)
		return []models.RFID{}, nil
	}

	r.logger.Debug("RFIDs found", "count", len(rfids), "rfid", rfid)
	return rfids, nil
}

func (r *HouseholdRepositoryImpl) GetSubscriberIDByFlatIDandPhone(ctx context.Context, flatID int, phone string) (int, error) {
	r.logger.Info("GetSubscriberIDByFlatIDandNumber", "flatID", flatID, "phone", phone)

	query := `
		SELECT house_subscriber_id
		FROM houses_subscribers_mobile
		WHERE house_subscriber_id IN
			  (SELECT house_subscriber_id
			   FROM houses_flats_subscribers
			   WHERE house_flat_id = $1)
		AND id = $2
		ORDER BY id		
	`

	var subscriberID int
	err := r.db.QueryRow(ctx, query, flatID, phone).Scan(&subscriberID)
	if err != nil {
		r.logger.Error("Query executing failed", "error", err)
		return 0, fmt.Errorf("query failed: %w", err)
	}

	return subscriberID, nil
}

func (r *HouseholdRepositoryImpl) FlatIDsByDomophoneIDAndPhone(ctx context.Context, domophoneID int, phone string) ([]int, error) {
	query := `
	SELECT hf.house_flat_id
	FROM houses_flats hf
	WHERE EXISTS (
		SELECT 1 FROM houses_entrances_flats hef
	INNER JOIN houses_entrances he ON hef.house_entrance_id = he.house_entrance_id
	WHERE hef.house_flat_id = hf.house_flat_id AND he.house_domophone_id = $1
	)
	AND EXISTS (
		SELECT 1 FROM houses_flats_subscribers hfs
	INNER JOIN houses_subscribers_mobile hsm ON hfs.house_subscriber_id = hsm.house_subscriber_id
	WHERE hfs.house_flat_id = hf.house_flat_id AND hsm.id = $2
	)`

	rows, err := r.db.Query(ctx, query, domophoneID, phone)
	if err != nil {
		r.logger.Error("Query executing failed", "error", err)
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var subscriberIDs []int
	for rows.Next() {
		var subscriberID int
		if err := rows.Scan(&subscriberID); err != nil {
			r.logger.Error("Failed to scan subscriber_id", "error", err)
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		subscriberIDs = append(subscriberIDs, subscriberID)
	}
	if err := rows.Err(); err != nil {
		r.logger.Error("Failed to scan subscriber_ids", "error", err)
		return nil, fmt.Errorf("failed to scan subscriber_ids: %w", err)
	}

	return subscriberIDs, nil
}

func (r *HouseholdRepositoryImpl) GetMobileDeviceByID(ctx context.Context, deviceID int) (models.MobileDevice, error) {
	query := `SELECT 
    		subscriber_device_id,
    		house_subscriber_id,
    		device_token,
    		auth_token,
    		platform,
    		push_token,
    		push_token_type,
    		voip_token,
    		registered,
    		last_seen,
    		voip_enabled,
    		ip,
    		ua,
    		push_disable,
    		money_disable,
    		version,
    		bundle
    	FROM houses_subscribers_devices 
    	WHERE subscriber_device_id = $1`

	var device models.MobileDevice
	err := r.db.QueryRow(ctx, query, deviceID).Scan(
		&device.DeviceID,
		&device.SubscriberID,
		&device.DeviceToken,
		&device.AuthToken,
		&device.Platform,
		&device.PushToken,
		&device.PushTokenType,
		&device.VoipToken,
		&device.Registered,
		&device.LastSeen,
		&device.VoipEnabled,
		&device.IP,
		&device.UA,
		&device.PushDisable,
		&device.MoneyDisable,
		&device.Version,
		&device.Bundle,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return models.MobileDevice{}, fmt.Errorf("mobile device with ID %d not found", deviceID)
		}
		return models.MobileDevice{}, fmt.Errorf("failed to get mobile device: %w", err)
	}

	return device, nil
}

func (r *HouseholdRepositoryImpl) GetHouseByEntranceID(ctx context.Context, entranceID int) (models.House, error) {
	const query = `
        SELECT
            ah.address_house_id,
            ah.address_settlement_id,
            ah.address_street_id,
            ah.house_uuid,
            ah.house_type,
            ah.house_type_full,
            ah.house_full,
            ah.house,
            ah.lat,
            ah.lon,
            ah.company_id
        FROM addresses_houses ah
        WHERE ah.address_house_id = (
            SELECT hhe.address_house_id 
            FROM houses_houses_entrances hhe 
            WHERE hhe.house_entrance_id = $1
            LIMIT 1
        )`

	r.logger.Debug("GetHouseByEntranceID", "entranceID", entranceID)

	var house models.House
	err := r.db.QueryRow(ctx, query, entranceID).Scan(
		&house.HouseID,
		&house.SettlementID,
		&house.StreetID,
		&house.HouseUUID,
		&house.HouseType,
		&house.HouseTypeFull,
		&house.HouseFull,
		&house.House,
		&house.Lat,
		&house.Lon,
		&house.CompanyID,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Warn("House not found by entrance ID", "entranceID", entranceID)
			return models.House{}, fmt.Errorf("house with entrance ID %d not found", entranceID)
		}
		r.logger.Error("Failed to get house by entrance ID", "error", err, "entranceID", entranceID)
		return models.House{}, fmt.Errorf("failed to get house: %w", err)
	}

	return house, nil
}
