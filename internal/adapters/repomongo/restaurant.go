package repomongo

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"restaurants-e2/internal/domain"
	"restaurants-e2/internal/ports"
)

type RestaurantRepoMongo struct{ coll *mongo.Collection }

var _ ports.RestaurantRepository = (*RestaurantRepoMongo)(nil)

func (r *RestaurantRepoMongo) Create(ctx context.Context, rest *domain.Restaurant) error {
	now := time.Now().UTC()
	if rest.ID == "" {
		rest.ID = uuid.NewString()
	}
	if rest.CreatedAt.IsZero() {
		rest.CreatedAt = now
	}
	rest.UpdatedAt = now
	_, err := r.coll.InsertOne(ctx, rest)
	return err
}

func (r *RestaurantRepoMongo) FindByID(ctx context.Context, id string) (*domain.Restaurant, error) {
	var out domain.Restaurant
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&out)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *RestaurantRepoMongo) FindAll(ctx context.Context) ([]domain.Restaurant, error) {
	cur, err := r.coll.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	out := make([]domain.Restaurant, 0)
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}
