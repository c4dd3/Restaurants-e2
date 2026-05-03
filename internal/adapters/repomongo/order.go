package repomongo

// order.go — sub-DAO de órdenes para MongoDB.

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"restaurants-e2/internal/domain"
	"restaurants-e2/internal/ports"
)

type OrderRepoMongo struct{ coll *mongo.Collection }

var _ ports.OrderRepository = (*OrderRepoMongo)(nil)

func (r *OrderRepoMongo) Create(ctx context.Context, o *domain.Order) error {
	if o.ID == "" {
		o.ID = uuid.NewString()
	}
	if o.Status == "" {
		o.Status = domain.StatusPending
	}
	if o.CreatedAt.IsZero() {
		o.CreatedAt = time.Now().UTC()
	}
	for i := range o.Items {
		if o.Items[i].ID == "" {
			o.Items[i].ID = uuid.NewString()
		}
		o.Items[i].OrderID = o.ID
	}
	_, err := r.coll.InsertOne(ctx, o)
	return err
}

func (r *OrderRepoMongo) FindByID(ctx context.Context, id string) (*domain.Order, error) {
	var out domain.Order
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&out)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &out, nil
}
