package mongo

import (
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

var (
	after            = options.After
	t                = true
	defaultIsolation = Isolation{
		Read:  readconcern.Snapshot(),
		Write: writeconcern.New(writeconcern.WMajority()),
	}
)

type Configuration struct {
	AuthMechanism      string
	Host            string
	Username           string
	Password           string
	AuthDB             string
	ReplicaSetName     string
	DatabaseName       string
	SSL                bool
	SecondaryPreferred bool
	DoHealthCheck      bool
}

func MapToDBConfig(m map[string]string) Configuration{
	cfg:=Configuration{
		Host: m["HOST"],
		DatabaseName: m["DB"],
		Username: m["USERNAME"],
		Password: m["PASSWORD"],
		AuthDB: m["AUTHDB"],
	}
	return cfg	
}
//  Find mongodb isolation levels at:
//	https://docs.mongodb.com/manual/reference/read-concern/
//	https://docs.mongodb.com/manual/reference/write-concern/
type Isolation struct {
	Read  *readconcern.ReadConcern
	Write *writeconcern.WriteConcern
}

type Client struct {
	Name        string
	Config      *Configuration
	OnConnected OnConnectedHandler
	c           *mongo.Client
}

type OnConnectedHandler = func(database *mongo.Database) error

type TransactionHandler = func(ctx mongo.SessionContext) (interface{}, error)

func MapCusorI_contains(m map[string]*Client, item string) bool {
	if len(m)==0{
		return false
	}
	if _, ok := m[item]; ok {
		return true
	}
	return false
}