package repomongo

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"restaurants-e2/internal/domain"
	"restaurants-e2/internal/ports"
)

type ProductRepoMongo struct {
	coll *mongo.Collection
}

var _ ports.ProductRepository = (*ProductRepoMongo)(nil)

func NewProductRepository(coll *mongo.Collection) *ProductRepoMongo {
	return &ProductRepoMongo{coll: coll}
}

func (r *ProductRepoMongo) FindByID(ctx context.Context, id string) (*domain.Product, error) {
	var product domain.Product
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&product)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *ProductRepoMongo) FindByCategory(ctx context.Context, category string) ([]domain.Product, error) {
	cur, err := r.coll.Find(ctx, bson.M{"category": category})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var products []domain.Product
	if err := cur.All(ctx, &products); err != nil {
		return nil, err
	}
	return products, nil
}

func (r *ProductRepoMongo) FindAll(ctx context.Context) ([]domain.Product, error) {
	cur, err := r.coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var products []domain.Product
	if err := cur.All(ctx, &products); err != nil {
		return nil, err
	}
	return products, nil
}

func (r *ProductRepoMongo) Create(ctx context.Context, p *domain.Product) error {
	if p.ID == "" {
		p.ID = uuid.NewString()
	}
	if p.Category == "" {
		return errors.New("category es obligatoria para productos en MongoDB")
	}
	_, err := r.coll.InsertOne(ctx, p)
	return err
}

func (r *ProductRepoMongo) Update(ctx context.Context, p *domain.Product) error {
	if p.ID == "" {
		return errors.New("id es obligatorio para actualizar producto")
	}
	if p.Category == "" {
		return errors.New("category es obligatoria para productos en MongoDB")
	}

	update := bson.M{
		"$set": bson.M{
			"menu_id":       p.MenuID,
			"restaurant_id": p.RestaurantID,
			"name":          p.Name,
			"description":   p.Description,
			"category":      p.Category,
			"price":         p.Price,
			"available":     p.Available,
		},
	}
	_, err := r.coll.UpdateOne(ctx, bson.M{"_id": p.ID}, update)
	return err
}

func (r *ProductRepoMongo) Delete(ctx context.Context, id string) error {
	_, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
