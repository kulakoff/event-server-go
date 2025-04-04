package repository

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kulakoff/event-server-go/internal/repository/models"
	"log/slog"
)

type CameraRepository interface {
	GetStreamByIP(ctx context.Context, ip string) (*models.Stream, error)
	GetCamera(ctx context.Context, id int)
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

func (r *CameraRepositoryImpl) GetCamera(ctx context.Context, id int) {
	//TODO: implement me
}
