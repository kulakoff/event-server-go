package repository

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
)

type CameraRepository interface {
	GetStreamByIP(ctx context.Context, ip string) (*Stream, error)
}

type Stream struct {
	ID     int
	UrlDVR string
	UrlFRS string
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

func (r *CameraRepositoryImpl) GetStreamByIP(ctx context.Context, ip string) (*Stream, error) {
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
