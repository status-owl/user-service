package service

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// UserStore is responsible for storing and fetching of users
type UserStore interface {
	Create(ctx context.Context, user *User) (string, error)
	FindByID(ctx context.Context, id string) (*User, error)
}

// mongoUserStore implements UserStore using mongodb as backing db
type mongoUserStore struct {
	client *mongo.Client
}

const (
	databaseName   = "owl-users"
	collectionName = "users"
)

func NewUserStore(client *mongo.Client) UserStore {
	return &mongoUserStore{client}
}

type mongoUser struct {
	ID      primitive.ObjectID `bson:"_id"`
	Name    string             `bson:"name"`
	EMail   string             `bson:"email"`
	PwdHash string             `bson:"pwd_hash"`
	Role    string             `bson:"role"`
}

// newMongoUser creates a new mongoUser from given user instance
// note that the id is going to be overwritten with generated one based on current timestamp
func newMongoUser(user *User) *mongoUser {
	return &mongoUser{
		ID:      primitive.NewObjectID(),
		Name:    user.Name,
		EMail:   user.EMail,
		PwdHash: user.PwdHash,
		Role:    string(user.Role),
	}
}

func (u *mongoUser) toUser() *User {
	return &User{
		ID:      u.ID.Hex(),
		Name:    u.Name,
		EMail:   u.EMail,
		PwdHash: u.PwdHash,
		Role:    RoleFromString(u.Role),
	}
}

// returns users collections
func (s *mongoUserStore) col() *mongo.Collection {
	return s.client.
		Database(databaseName).
		Collection(collectionName)
}

func (s *mongoUserStore) Create(ctx context.Context, user *User) (string, error) {
	c := s.col()
	result, err := c.InsertOne(ctx, newMongoUser(user))
	if err != nil {
		return "", err
	}

	return result.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (s *mongoUserStore) FindByID(ctx context.Context, id string) (*User, error) {
	c := s.col()
	var u mongoUser
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	err = c.FindOne(ctx, bson.M{"_id": objectId}).Decode(&u)
	if err != nil {
		return nil, err
	}

	return u.toUser(), nil
}
