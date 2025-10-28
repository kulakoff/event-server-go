package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kulakoff/event-server-go/internal/app/event-server-go/repository/models"
	"log/slog"
)

type CameraRepository interface {
	GetStreamByIP(ctx context.Context, ip string) (*models.Stream, error)
	GetCamera(ctx context.Context, id int) (*models.Camera, error)
	GetCameraByIP(ctx context.Context, ip string) (*models.Camera, error)
}

type CameraRepositoryImpl struct {
	db     *pgxpool.Pool
	logger *slog.Logger
	parent *PostgresRepository
}

func NewCameraRepository(parent *PostgresRepository) *CameraRepositoryImpl {
	return &CameraRepositoryImpl{
		db:     parent.db,
		logger: parent.logger,
		parent: parent,
	}
}

func (r *CameraRepositoryImpl) GetStreamByIP(ctx context.Context, ip string) (*models.Stream, error) {
	// TODO implement me
	// FIXME: stub stream
	if ip == "192.168.13.152" {
		return &models.Stream{
			ID:     8,
			UrlDVR: "https://dvr-example.com/stream-name/index.m3u8",
			UrlFRS: "http://localhost:9051",
		}, nil
	}
	if ip == "37.235.188.212" {
		return &models.Stream{
			ID:     8,
			UrlDVR: "https://dvr-example.com/stream-name/index.m3u8",
			UrlFRS: "https://webhook.site/5c40e512-64c6-49b8-96d0-d6d028f8181f",
		}, nil
	}

	return nil, nil
}

func (r *CameraRepositoryImpl) GetCamera(ctx context.Context, id int) (*models.Camera, error) {
	//TODO: implement me
	r.logger.Debug("GetCamera", "camera_id", id)

	query := `
        SELECT camera_id, enabled, model, url, stream, credentials, name, dvr_stream, timezone, 
               lat, lon, direction, angle, distance, frs, common, ip, sub_id, sound, comments, 
               md_area, rc_area, frs_mode, ext
        FROM cameras 
        WHERE camera_id = $1
    `
	var camera models.Camera
	err := r.db.QueryRow(ctx, query, id).Scan(
		&camera.CameraID,
		&camera.Enabled,
		&camera.Model,
		&camera.URL,
		&camera.Stream,
		&camera.Credentials,
		&camera.Name,
		&camera.DVRStream,
		&camera.Timezone,
		&camera.Lat,
		&camera.Lon,
		&camera.Direction,
		&camera.Angle,
		&camera.Distance,
		&camera.FRS,
		&camera.Common,
		&camera.IP,
		&camera.SubID,
		&camera.Sound,
		&camera.Comments,
		&camera.MdArea,
		&camera.RcArea,
		&camera.FrsMode,
		&camera.Ext,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Warn("Camera not ")
		}
		r.logger.Error("Database query failed", "error", err, "camera_id", id)
		return nil, fmt.Errorf("failed to query camera: %s", err)
	}
	return &camera, nil
}

func (r *CameraRepositoryImpl) GetCameraByIP(ctx context.Context, ip string) (*models.Camera, error) {
	//TODO: implement me
	r.logger.Debug("GetCameraByIP", "camera_ip", ip)

	query := `
        SELECT camera_id, enabled, model, url, stream, credentials, name, dvr_stream, timezone, 
               lat, lon, direction, angle, distance, frs, common, ip, sub_id, sound, comments, 
               md_area, rc_area, frs_mode, ext
        FROM cameras 
        WHERE ip = $1
    `
	var camera models.Camera
	err := r.db.QueryRow(ctx, query, ip).Scan(
		&camera.CameraID,
		&camera.Enabled,
		&camera.Model,
		&camera.URL,
		&camera.Stream,
		&camera.Credentials,
		&camera.Name,
		&camera.DVRStream,
		&camera.Timezone,
		&camera.Lat,
		&camera.Lon,
		&camera.Direction,
		&camera.Angle,
		&camera.Distance,
		&camera.FRS,
		&camera.Common,
		&camera.IP,
		&camera.SubID,
		&camera.Sound,
		&camera.Comments,
		&camera.MdArea,
		&camera.RcArea,
		&camera.FrsMode,
		&camera.Ext,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Warn("Camera not ")
		}
		r.logger.Error("Database query failed", "error", err, "camera_id", ip)
		return nil, fmt.Errorf("failed to query camera: %s", err)
	}
	return &camera, nil
}
