package store

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/pkg/errors"
	"github.com/status-owl/user-service/pkg/model"
	"github.com/stretchr/testify/assert"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	gklog "github.com/go-kit/log"
)

var mongoClient *mongo.Client
var store UserStore

func TestMain(m *testing.M) {
	ctx := context.Background()

	mongoContainer, err := setupMongo(ctx)
	if err != nil {
		log.Fatalf("failed to initialize mongodb via docker: %s", err.Error())
	}

	defer mongoContainer.Terminate(ctx)

	mongoClient, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoContainer.URI))
	if err != nil {
		log.Fatalf("failed to establish a mongodb connection: %s", err.Error())
	}

	baseStore, err := NewUserStore(mongoClient)
	if err != nil {
		log.Fatalf("failed to create the userStore: %s", err.Error())
	}
	store = NewLoggingUserStore(baseStore, gklog.NewJSONLogger(os.Stdout))

	os.Exit(m.Run())
}

func TestCreateUser(t *testing.T) {
	expectedUser := &model.User{
		Name:    "John Doe",
		EMail:   "john.doe@example.com",
		PwdHash: "abcde",
	}
	id, err := store.Create(context.Background(), expectedUser)
	if err != nil {
		t.Error(err)
	}

	// now we expect that the actualUser exists
	actualUser, err := store.FindByID(context.Background(), id)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, id, actualUser.ID)
	assert.NotEqual(t, id, primitive.NilObjectID.Hex(), "make sure the id was generated")
	assert.Equal(t, expectedUser.Name, actualUser.Name)
	assert.Equal(t, expectedUser.EMail, actualUser.EMail)
	assert.Equal(t, expectedUser.PwdHash, actualUser.PwdHash)
}

type mongoContainer struct {
	tc.Container
	URI string
}

func setupMongo(ctx context.Context) (*mongoContainer, error) {
	user, pwd := "root", "secret"

	req := tc.ContainerRequest{
		Image:        "mongo:4.4.2-bionic",
		ExposedPorts: []string{"27017/tcp"},
		Env: map[string]string{
			"MONGO_INITDB_ROOT_USERNAME": user,
			"MONGO_INITDB_ROOT_PASSWORD": pwd,
		},
		WaitingFor: wait.ForLog("Waiting for connections"),
	}

	mongo, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err != nil {
		return nil, err
	}

	mappedPort, err := mongo.MappedPort(ctx, "27017")
	if err != nil {
		mongo.Terminate(ctx)
		return nil, errors.Wrap(err, "failed to determine mapped port")
	}

	host, err := mongo.Host(ctx)
	if err != nil {
		mongo.Terminate(ctx)
		return nil, errors.Wrap(err, "failed to determine host")
	}

	uri := fmt.Sprintf("mongodb://%s:%s@%s:%s", user, pwd, host, mappedPort.Port())
	return &mongoContainer{Container: mongo, URI: uri}, nil
}
