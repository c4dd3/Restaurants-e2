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

type MenuRepoMongo struct {
	db   *mongo.Database
	coll *mongo.Collection
}

var _ ports.MenuRepository = (*MenuRepoMongo)(nil)

func (r *MenuRepoMongo) FindByID(ctx context.Context, id string) (*domain.Menu, error) {
	var m domain.Menu
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&m)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	cur, err := r.db.Collection("products").Find(ctx, bson.M{"menu_id": id}, options.Find().SetSort(bson.D{{Key: "name", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	products := make([]domain.Product, 0)
	if err := cur.All(ctx, &products); err != nil {
		return nil, err
	}
	m.Products = products
	return &m, nil
}

func (r *MenuRepoMongo) Create(ctx context.Context, m *domain.Menu) error {
	now := time.Now().UTC()
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = now
	}
	m.UpdatedAt = now
	_, err := r.coll.InsertOne(ctx, m)
	return err
}

func (r *MenuRepoMongo) Update(ctx context.Context, id string, req *domain.UpdateMenuRequest) (*domain.Menu, error) {
	set := bson.M{"updated_at": time.Now().UTC()}
	if req.Name != "" {
		set["name"] = req.Name
	}
	if req.Description != "" {
		set["description"] = req.Description
	}

	var updated domain.Menu
	err := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": set},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&updated)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	if req.Products != nil {
		session, err := r.db.Client().StartSession()
		if err != nil {
			return nil, err
		}
		defer session.EndSession(ctx)

		_, err = session.WithTransaction(ctx, func(sc mongo.SessionContext) (interface{}, error) {
			productsColl := r.db.Collection("products")
			if _, err := productsColl.DeleteMany(sc, bson.M{"menu_id": id}); err != nil {
				return nil, err
			}
			docs := make([]interface{}, 0, len(req.Products))
			updated.Products = make([]domain.Product, 0, len(req.Products))
			for _, rp := range req.Products {
				p := domain.Product{
					ID:          uuid.NewString(),
					MenuID:      id,
					Name:        rp.Name,
					Description: rp.Description,
					Category:    rp.Category,
					Price:       rp.Price,
					Available:   rp.Available,
				}
				docs = append(docs, p)
				updated.Products = append(updated.Products, p)
			}
			if len(docs) > 0 {
				if _, err := productsColl.InsertMany(sc, docs); err != nil {
					return nil, err
				}
			}
			return nil, nil
		})
		if err != nil {
			return nil, err
		}
		return &updated, nil
	}

	return r.FindByID(ctx, id)
}

func (r *MenuRepoMongo) Delete(ctx context.Context, id string) error {
	session, err := r.db.Client().StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sc mongo.SessionContext) (interface{}, error) {
		if _, err := r.db.Collection("products").DeleteMany(sc, bson.M{"menu_id": id}); err != nil {
			return nil, err
		}
		if _, err := r.coll.DeleteOne(sc, bson.M{"_id": id}); err != nil {
			return nil, err
		}
		return nil, nil
	})
	return err
}
