package storage

import (
	"context"
	"fmt"
	"github.com/kulakoff/event-server-go/internal/config"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log/slog"
	"time"
)

type MongoHandler struct {
	logger *slog.Logger
	client *mongo.Client
	db     *mongo.Database
}

func NewMongoDb(logger *slog.Logger, mongoDbConfig *config.MongoDbConfig) (*MongoHandler, error) {
	clientOptions := options.Client().ApplyURI(mongoDbConfig.URI)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// check connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	logger.Info("Connected to MongoDB")
	return &MongoHandler{
		logger: logger,
		client: client,
		db:     client.Database(mongoDbConfig.Database),
	}, nil
}

func (m *MongoHandler) SaveFile(filename string, metadata map[string]interface{}, filedata []byte) (string, error) {
	// TODO: files.md5 is deprecated. add  md5 hash to metadata
	bucket, err := gridfs.NewBucket(m.db)
	if err != nil {
		return "", fmt.Errorf("failed to create GRIDFS bucket: %w", err)
	}

	uploadOptions := options.GridFSUpload().SetMetadata(metadata)
	uploadStream, err := bucket.OpenUploadStream(filename, uploadOptions)
	if err != nil {
		return "", fmt.Errorf("failed to open upload stream: %w", err)
	}
	defer uploadStream.Close()

	_, err = uploadStream.Write(filedata)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}
	fileId := uploadStream.FileID.(primitive.ObjectID)
	fileIdHex := fileId.Hex()

	return fileIdHex, nil
}
