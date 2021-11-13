package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/status-owl/user-service/pkg/model"
)

// mongoUserStore implements UserStore using mongodb as backing db
type mongoUserStore struct {
	client *mongo.Client
}

const (
	databaseName   = "user-service"
	collectionName = "users"
)

type mongoUser struct {
	ID      primitive.ObjectID `bson:"_id"`
	Name    string             `bson:"name"`
	EMail   string             `bson:"email"`
	PwdHash string             `bson:"pwd_hash"`
	Role    string             `bson:"role"`
}

var (
	ErrIndexCreation = errors.New("failed to create index")
)

// newMongoUser creates a new mongoUser from given user instance
// note that the id is going to be overwritten with generated one based on current timestamp
func newMongoUser(user *model.User) *mongoUser {
	return &mongoUser{
		ID:      primitive.NewObjectID(),
		Name:    user.Name,
		EMail:   user.EMail,
		PwdHash: user.PwdHash,
		Role:    string(user.Role),
	}
}

func (u *mongoUser) toUser() *model.User {
	return &model.User{
		ID:      u.ID.Hex(),
		Name:    u.Name,
		EMail:   u.EMail,
		PwdHash: u.PwdHash,
		Role:    model.RoleFromString(u.Role),
	}
}

// returns users collections
func (s *mongoUserStore) col() *mongo.Collection {
	return s.client.
		Database(databaseName).
		Collection(collectionName)
}

func (s *mongoUserStore) createIndexes() error {
	c := s.col()
	emailUnique := true
	emailIndex := mongo.IndexModel{
		Keys: bson.M{
			"email": 1,
		},
		Options: &options.IndexOptions{Unique: &emailUnique},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	_, err := c.Indexes().CreateOne(ctx, emailIndex)
	if err != nil {
		return ErrIndexCreation
	}

	return nil
}

// clear removes all users from the collection
func (s *mongoUserStore) clear(ctx context.Context) (int64, error) {
	result, err := s.col().DeleteMany(ctx, bson.M{})
	if err != nil {
		return 0, fmt.Errorf("failed to clear collection %s: %w", collectionName, err)
	}

	return result.DeletedCount, nil
}

func (s *mongoUserStore) Create(ctx context.Context, user *model.User) (string, error) {
	c := s.col()
	result, err := c.InsertOne(ctx, newMongoUser(user))
	if err != nil {
		return "", err
	}

	return result.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (s *mongoUserStore) FindByID(ctx context.Context, id string) (*model.User, error) {
	var u mongoUser
	// first check if the id can be converted to a mongo's object id
	// if not - pretend that the user doesn't exist
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrNotFound
	}

	err = s.col().FindOne(ctx, bson.M{"_id": objectId}).Decode(&u)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to find user by id: %w", err)
	}

	return u.toUser(), nil
}

func (s *mongoUserStore) FindByEMail(ctx context.Context, email string) (*model.User, error) {
	var u mongoUser
	if err := s.col().FindOne(ctx, bson.M{"email": email}).Decode(&u); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}

	return u.toUser(), nil
}

func (s *mongoUserStore) HasUsersWithRole(ctx context.Context, role model.Role) (bool, error) {
	matchStage := bson.D{{"$match", bson.D{{"role", string(role)}}}}
	countStage := bson.D{{"$count", "count"}}

	cursor, err := s.col().Aggregate(ctx, mongo.Pipeline{matchStage, countStage})
	if err != nil {
		return false, fmt.Errorf("failed to execute pipeline: %w", err)
	}

	var results []bson.M
	err = cursor.All(ctx, &results)
	if err != nil {
		return false, fmt.Errorf("failed to extract pipeline results: %w", err)
	}

	if len(results) == 0 {
		return false, nil
	}

	count := results[0]["count"].(int32)
	fmt.Println(count)

	return count > 0, nil
}
