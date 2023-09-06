package mongo

import (
	"strings"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
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
	Host               string
	Username           string
	Password           string
	AuthDB             string
	ReplicaSetName     string
	DatabaseName       string
	SSL                bool
	SecondaryPreferred bool
	DoHealthCheck      bool
	WriteConcern       bool
	ReadConcern        bool
}

func MapToDBConfig(m map[string]string) Configuration {
	cfg := Configuration{
		Host:           m["HOST"],
		DatabaseName:   m["DB"],
		Username:       m["USERNAME"],
		Password:       m["PASSWORD"],
		AuthDB:         m["AUTHDB"],
		ReplicaSetName: m["RS_NAME"],
	}

	if strings.ToLower(m["WRITE_CONCERN"]) == "true" {
		cfg.WriteConcern = true
	}

	if strings.ToLower(m["READ_CONCERN"]) == "true" {
		cfg.ReadConcern = true
	}

	return cfg
}

//	 Find mongodb isolation levels at:
//		https://docs.mongodb.com/manual/reference/read-concern/
//		https://docs.mongodb.com/manual/reference/write-concern/
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
	if len(m) == 0 {
		return false
	}
	if _, ok := m[item]; ok {
		return true
	}
	return false
}
