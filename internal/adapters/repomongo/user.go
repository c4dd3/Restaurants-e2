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

type UserRepoMongo struct{ coll *mongo.Collection }

var _ ports.UserRepository = (*UserRepoMongo)(nil)

func (r *UserRepoMongo) FindByID(ctx context.Context, id string) (*domain.User, error) {
	var u domain.User
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&u)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepoMongo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	err := r.coll.FindOne(ctx, bson.M{"email": email}).Decode(&u)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepoMongo) Create(ctx context.Context, u *domain.User) error {
	now := time.Now().UTC()
	if u.ID == "" {
		u.ID = uuid.NewString()
	}
	if u.CreatedAt.IsZero() {
		u.CreatedAt = now
	}
	u.UpdatedAt = now
	_, err := r.coll.InsertOne(ctx, u)
	return err
}

func (r *UserRepoMongo) Update(ctx context.Context, id string, req *domain.UpdateUserRequest) (*domain.User, error) {
	set := bson.M{"updated_at": time.Now().UTC()}
	if req.Name != "" {
		set["name"] = req.Name
	}
	if req.Email != "" {
		set["email"] = req.Email
	}

	var out domain.User
	err := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": set},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&out)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *UserRepoMongo) Delete(ctx context.Context, id string) error {
	_, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
