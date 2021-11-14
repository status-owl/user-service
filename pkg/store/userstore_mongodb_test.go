package store

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

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

	defer func() {
		if err = mongoContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate mongodb container: %s", err.Error())
		}
	}()

	mongoClient, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoContainer.URI))
	if err != nil {
		log.Fatalf("failed to establish a mongodb connection: %s", err.Error())
	}

	store, err = NewUserStore(mongoClient, gklog.NewJSONLogger(os.Stdout))
	if err != nil {
		log.Fatalf("failed to create the userStore: %s", err.Error())
	}

	os.Exit(m.Run())
}

func TestCreateUser(t *testing.T) {
	expectedUser := fixtures.users.withoutRole
	id, err := store.Create(context.Background(), expectedUser)
	if err != nil {
		t.Error(err)
	}

	// now we expect that the actualUser exists
	actualUser, err := store.FindByID(context.Background(), id)
	assert.Nil(t, err)
	assert.Equal(t, id, actualUser.ID)
	assert.NotEqual(t, id, primitive.NilObjectID.Hex(), "make sure the id was generated")
	assert.Equal(t, expectedUser.Name, actualUser.Name)
	assert.Equal(t, expectedUser.EMail, actualUser.EMail)
}

func TestFindByEmail(t *testing.T) {
	clearDB()
	expectedUser := fixtures.users.withoutRole

	id, err := store.Create(context.Background(), expectedUser)
	assert.Nil(t, err)

	actualUser, err := store.FindByEMail(context.Background(), expectedUser.EMail)
	assert.Nil(t, err)
	assert.Equal(t, id, actualUser.ID)
	assert.Equal(t, expectedUser.EMail, actualUser.EMail)
	assert.Equal(t, expectedUser.Name, actualUser.Name)

	_, err = store.FindByEMail(context.Background(), "not-existing-user@example.com")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestFindByID(t *testing.T) {
	_, err := store.FindByID(context.Background(), "123")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestHasUserWithRole(t *testing.T) {
	// clear db
	// make sure a admin user does not exist
	// writes a couple of users with one of them being admin
	// makes sure at least one admin user does exist
	clearDB()

	a := assert.New(t)

	// don't expect a admin user
	exist, err := store.HasUsersWithRole(context.Background(), model.Admin)
	a.Nil(err)
	a.False(exist)

	// write users
	for _, user := range fixturesAllUsers {
		_, err := store.Create(context.Background(), user)
		a.Nil(err)
	}

	// expect an admin
	exist, err = store.HasUsersWithRole(context.Background(), model.Admin)
	a.Nil(err)
	a.True(exist)
}

func TestClear(t *testing.T) {
	a := assert.New(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// persist users
	for _, u := range fixturesAllUsers {
		_, err := store.FindByEMail(ctx, u.EMail)
		if err == ErrNotFound {
			_, err = store.Create(ctx, u)
			a.Nil(err)
			continue
		}
		a.Nil(err)
	}

	count, err := store.clear(ctx)
	a.Equal(count, int64(len(fixturesAllUsers)))
	a.Nil(err)

	// make sure no user does exist
	for _, u := range fixturesAllUsers {
		_, err := store.FindByID(ctx, u.ID)
		a.ErrorIs(err, ErrNotFound)
	}
}

type mongoContainer struct {
	tc.Container
	URI string
}

func clearDB() {
	_, err := store.clear(context.Background())
	if err != nil {
		panic(err)
	}
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

	container, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err != nil {
		return nil, err
	}

	mappedPort, err := container.MappedPort(ctx, "27017")
	if err != nil {
		if err = container.Terminate(ctx); err != nil {
			panic(err)
		}
		return nil, errors.Wrap(err, "failed to determine mapped port")
	}

	host, err := container.Host(ctx)
	if err != nil {
		if err = container.Terminate(ctx); err != nil {
			panic(err)
		}
		return nil, errors.Wrap(err, "failed to determine host")
	}

	uri := fmt.Sprintf("mongodb://%s:%s@%s:%s", user, pwd, host, mappedPort.Port())
	return &mongoContainer{Container: container, URI: uri}, nil
}

var fixtures = struct {
	users struct {
		undefined, admin, reporter, withoutRole *model.User
	}
}{
	users: struct{ undefined, admin, reporter, withoutRole *model.User }{
		undefined: &model.User{
			Name:    "John Doe",
			EMail:   "john.doe@example.com",
			Role:    model.Undefined,
		},
		admin: &model.User{
			Name:    "Mary Doe",
			EMail:   "mary.doe@example.com",
			Role:    model.Admin,
		},
		reporter: &model.User{
			Name:    "Fritz Nebel",
			EMail:   "fritz.nebel@example.com",
			Role:    model.Reporter,
		},
		withoutRole: &model.User{
			Name:    "Mark Defoe",
			EMail:   "make.d@example.com",
		},
	},
}

// fixturesAllUsers contains all users from the fixture
var fixturesAllUsers = []*model.User{
	fixtures.users.undefined,
	fixtures.users.admin,
	fixtures.users.reporter,
	fixtures.users.withoutRole,
}
