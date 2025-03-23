package repository

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
)

type PostgresRepository struct {
	db     *pgxpool.Pool
	logger *slog.Logger
	// sub repository
	Cameras    CameraRepository
	Households HouseHoldRepository
}

func NewPostgresRepository(db *pgxpool.Pool, logger *slog.Logger) (*PostgresRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("db pool is nil")
	}

	if err := db.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("could not connect to database: %w", err)
	}

	logger.Info("postgres repository initialized")

	cameraRepo := NewCameraRepository(db, logger)
	householdRepo := NewHouseholdRepository(db, logger)

	return &PostgresRepository{
		db:         db,
		logger:     logger,
		Cameras:    cameraRepo,
		Households: householdRepo,
	}, nil
}
