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

type ReservationRepoMongo struct {
	db   *mongo.Database
	coll *mongo.Collection
}

var _ ports.ReservationRepository = (*ReservationRepoMongo)(nil)

func (r *ReservationRepoMongo) Create(ctx context.Context, res *domain.Reservation) error {
	if res.RestaurantID == "" {
		return mongo.ErrNilValue
	}
	if res.ID == "" {
		res.ID = uuid.NewString()
	}
	if res.Status == "" {
		res.Status = domain.StatusPending
	}
	if res.CreatedAt.IsZero() {
		res.CreatedAt = time.Now().UTC()
	}
	_, err := r.coll.InsertOne(ctx, res)
	return err
}

func (r *ReservationRepoMongo) FindByID(ctx context.Context, id string) (*domain.Reservation, error) {
	var res domain.Reservation
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&res)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *ReservationRepoMongo) Cancel(ctx context.Context, id string) error {
	res, err := r.coll.UpdateOne(
		ctx,
		bson.M{"_id": id, "status": bson.M{"$ne": domain.StatusCancelled}},
		bson.M{"$set": bson.M{"status": domain.StatusCancelled}},
	)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		// El id no existe O ya estaba cancelada — igual que el comportamiento de Postgres.
		return domain.ErrNotFound
	}
	return nil
}

func (r *ReservationRepoMongo) CheckAvailability(ctx context.Context, restaurantID string, partySize int) (int, error) {
	var rest domain.Restaurant
	err := r.db.Collection("restaurants").FindOne(ctx, bson.M{"_id": restaurantID}).Decode(&rest)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return 0, domain.ErrNotFound
	}
	if err != nil {
		return 0, err
	}

	now := time.Now().UTC()
	windowEnd := now.Add(2 * time.Hour)
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.M{
			"restaurant_id": restaurantID,
			"status":        domain.StatusConfirmed,
			"date":          bson.M{"$gte": now, "$lt": windowEnd},
		}}},
		bson.D{{Key: "$group", Value: bson.M{
			"_id":   nil,
			"total": bson.M{"$sum": "$party_size"},
		}}},
	}

	cur, err := r.coll.Aggregate(ctx, pipeline, options.Aggregate())
	if err != nil {
		return 0, err
	}
	defer cur.Close(ctx)

	total := 0
	if cur.Next(ctx) {
		var agg struct {
			Total int `bson:"total"`
		}
		if err := cur.Decode(&agg); err != nil {
			return 0, err
		}
		total = agg.Total
	}
	if err := cur.Err(); err != nil {
		return 0, err
	}

	return rest.Capacity - total, nil
}
