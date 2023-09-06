package mongo

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/goonma/sdk/db/mongo/status"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

/*
Write concern:
	-[1]: primary only => default
	-[2]: 1 primary + 1 secondary , same for 2,3...
	-[0]: majority: (total node/2 + 1)
*/
/*
Read concern
	- [1] primary node : read alway go to primary
	- [2] primary nodePreferred: read go to primary, if primary die will to go secondary
	- [3] secondary node: read go to secondary only
	- [4] secondary nodePreferred: read go to secondary if die go to primary
*/
func NewClient(conf Configuration) *Client {
	return &Client{
		Config: &conf,
	}
}

func (c *Client) Connect() error {
	// setup default config
	hostname, _ := os.Hostname()
	heartBeat := 60 * time.Second
	maxIdle := 180 * time.Second
	socketTimeout := 60 * time.Second
	connectTimeout := 60 * time.Second
	serverSelectTimeout := 60 * time.Second
	min := uint64(2)

	// setup options
	opt := &options.ClientOptions{
		AppName: &hostname,
		Auth: &options.Credential{
			AuthMechanism: c.Config.AuthMechanism,
			AuthSource:    c.Config.AuthDB,
			Username:      c.Config.Username,
			Password:      c.Config.Password,
		},
		HeartbeatInterval:      &heartBeat,
		MaxConnIdleTime:        &maxIdle,
		MinPoolSize:            &min,
		SocketTimeout:          &socketTimeout,
		ConnectTimeout:         &connectTimeout,
		ServerSelectionTimeout: &serverSelectTimeout,
	}

	if c.Config.WriteConcern {
		opt.WriteConcern = writeconcern.New(writeconcern.WMajority())
	}

	if c.Config.ReadConcern {
		opt.ReadConcern = readconcern.Majority()
	}

	opt.ApplyURI(c.Config.Host)
	if c.Config.SecondaryPreferred {
		opt.ReadPreference = readpref.SecondaryPreferred()
	}

	Client, err := mongo.NewClient(opt)
	if err != nil {
		return err
	}
	err = Client.Connect(context.TODO())
	if err != nil {
		return err
	}

	c.c = Client
	database := Client.Database(c.Config.DatabaseName)

	// try to test write & log connection
	if c.Config.DoHealthCheck {
		inst := Collection{
			ColName:        "_db_connection",
			TemplateObject: bson.M{},
		}
		inst.ApplyDatabase(database)
		go inst.CreateIndex(
			bson.D{{"created_time", 1}},
			&options.IndexOptions{
				Background: &t,
				ExpireAfterSeconds: func() *int32 {
					var s int32 = 86400
					return &s
				}(),
				Name: func() *string {
					n := "expire_after_1_day"
					return &n
				}(),
			})
		testResult := inst.Create(nil, bson.M{
			"host": hostname,
			"time": time.Now(),
		})

		if testResult.Status != status.DBStatus.Ok {
			return errors.New(testResult.Status + " / " + testResult.ErrorCode + " => " + testResult.Message)
		}
	}

	// on connected
	if c.OnConnected != nil {
		return c.OnConnected(database)
	}

	return nil
}

func (c *Client) Disconnect() error {
	if c.c == nil {
		return errors.New("connection has not been established")
	}
	return c.c.Disconnect(context.TODO())
}
