package repomongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"restaurants-e2/internal/config"
)

func NewClient(ctx context.Context, cfg config.MongoConfig) (*mongo.Client, error) {
	clientCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	opts := options.Client().
		ApplyURI(cfg.URI).
		SetMaxPoolSize(50).
		SetMinPoolSize(5).
		SetServerSelectionTimeout(5 * time.Second).
		SetReadPreference(readpref.PrimaryPreferred())

	client, err := mongo.Connect(clientCtx, opts)
	if err != nil {
		return nil, err
	}

	pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)
	defer pingCancel()

	if err := client.Ping(pingCtx, readpref.Primary()); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, err
	}

	return client, nil
}
