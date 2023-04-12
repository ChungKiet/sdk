package mongo

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/goonma/sdk/db/mongo/enum"
	"github.com/goonma/sdk/db/mongo/status"
	"github.com/goonma/sdk/log"
	"github.com/goonma/sdk/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	//"fmt"
)

type Collection struct {
	ColName        string
	TemplateObject interface{}
	TimeSeries     bool

	db  *mongo.Database
	col *mongo.Collection
}

func (m *Collection) newObject() interface{} {
	t := reflect.TypeOf(m.TemplateObject)
	v := reflect.New(t)
	return v.Interface()
}

func (m *Collection) newList(limit int) interface{} {
	t := reflect.TypeOf(m.TemplateObject)
	return reflect.MakeSlice(reflect.SliceOf(t), 0, limit).Interface()
}

func (m *Collection) interfaceSlice(slice interface{}) ([]interface{}, error) {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		return nil, errors.New("given a non-slice type")
	}

	ret := make([]interface{}, s.Len())

	for i := 0; i < s.Len(); i++ {
		ret[i] = s.Index(i).Interface()
	}

	return ret, nil
}

func (m *Collection) convertToObject(b bson.M) (interface{}, error) {
	obj := m.newObject()

	if b == nil {
		return obj, nil
	}

	bytes, err := bson.Marshal(b)
	if err != nil {
		return nil, err
	}

	bson.Unmarshal(bytes, obj)
	return obj, nil
}

func (m *Collection) convertToBson(ent interface{}) (bson.M, error) {
	if ent == nil {
		return bson.M{}, nil
	}

	sel, err := bson.Marshal(ent)
	if err != nil {
		return nil, err
	}

	obj := bson.M{}
	bson.Unmarshal(sel, &obj)

	return obj, nil
}

func (m *Collection) checkCollection() *status.DBResponse {
	if m.col == nil {
		return &status.DBResponse{
			Status:  status.DBStatus.Error,
			Message: "DB error: Collection " + m.ColName + " has not been initialized",
		}
	}

	return nil
}

func (m *Collection) parseSingleResult(result *mongo.SingleResult, action string) *status.DBResponse {
	// parse result
	obj := m.newObject()
	err := result.Decode(obj)
	if err != nil {
		return &status.DBResponse{
			Status:    status.DBStatus.Error,
			Message:   "DB Error: " + err.Error(),
			ErrorCode: enum.MAP_OBJECT_FAILED,
		}
	}

	// put to slice
	list := m.newList(1)
	listValue := reflect.Append(reflect.ValueOf(list), reflect.Indirect(reflect.ValueOf(obj)))

	return &status.DBResponse{
		Status:  status.DBStatus.Ok,
		Message: action + " " + m.ColName + " successfully",
		Data:    listValue.Interface(),
	}
}

func (m *Collection) parseConversionError(err error, action string) *status.DBResponse {
	return &status.DBResponse{
		Status:    status.DBStatus.Error,
		Message:   "DB error: " + action + " - Cannot convert object - " + err.Error(),
		ErrorCode: enum.MAP_OBJECT_FAILED,
	}
}

func (m *Collection) parseError(err error, action, code string) *status.DBResponse {
	if err == mongo.ErrNoDocuments {
		return &status.DBResponse{
			Status:    status.DBStatus.NotFound,
			Message:   "Not found any matched " + m.ColName,
			ErrorCode: enum.NO_DOCUMENT_FOUND,
		}
	} else {
		return &status.DBResponse{
			Status:    status.DBStatus.Error,
			Message:   "DB Error: " + err.Error(),
			ErrorCode: code,
		}
	}
}

func (m *Collection) GetClient() *mongo.Client {
	return m.db.Client()
}

func (m *Collection) GetDatabase() *mongo.Database {
	return m.db
}

func (m *Collection) GetCollection() *mongo.Collection {
	return m.col
}

func (m *Collection) ApplyDatabase(database *mongo.Database) *Collection {
	m.db = database
	m.col = database.Collection(m.ColName)
	return m
}

func (m *Collection) GetIndexes() ([]primitive.M, error) {
	cursor, err := m.col.Indexes().List(context.TODO())
	if err != nil {
		return nil, err
	}
	if e := cursor.Err(); e != nil {
		return nil, e
	}

	var result []primitive.M
	err = cursor.All(context.TODO(), &result)
	return result, err
}

func (m *Collection) CreateIndex(keys bson.D, options *options.IndexOptions) error {
	_, err := m.col.Indexes().CreateOne(context.TODO(), mongo.IndexModel{
		Keys:    keys,
		Options: options,
	})
	return err
}

//	@handler: the transaction will be committed when give a non-error
//	@isolation: will be default value when given nil attributes
//
// args[0] // number of retry time
// args[1] //sleep time for each  retry => milisecond
func (m *Collection) ApplyTransaction(handler TransactionHandler, isolation *Isolation, args ...interface{}) *status.DBResponse {
	// setup Isolation & txn option
	if isolation == nil {
		isolation = &defaultIsolation
	} else {
		if isolation.Read == nil {
			isolation.Read = defaultIsolation.Read
		}
		if isolation.Write == nil {
			isolation.Write = defaultIsolation.Write
		}
	}
	txnOpts := options.Transaction().SetWriteConcern(isolation.Write).SetReadConcern(isolation.Read)
	sessionOpts := options.Session().SetDefaultReadPreference(readpref.Primary())
	// start session
	session, err := m.db.Client().StartSession(sessionOpts)
	if err != nil {
		return &status.DBResponse{
			Status:    status.DBStatus.Error,
			Message:   "Failed to start session on " + m.db.Name() + " with error: " + err.Error(),
			ErrorCode: enum.START_SESSION_FAILED,
		}
	}
	defer session.EndSession(context.TODO())
	num_retry := 1
	sleep_time := 0 //milisecond
	if len(args) > 0 {
		temp := utils.ItoInt(args[0])
		if temp > 1 {
			num_retry = temp
		}
		if len(args) > 1 {
			temp := utils.ItoInt(args[1])
			if temp > 0 {
				sleep_time = temp
			}
		}
	}
	// apply transaction
	results, txnErr := session.WithTransaction(context.TODO(), handler, txnOpts)
	if txnErr != nil && num_retry > 1 {
		for i := 1; i < num_retry; i++ {
			results, txnErr = session.WithTransaction(context.TODO(), handler, txnOpts)
			if txnErr == nil {
				break
			}
			log.Warn(txnErr.Error()+" | retry", "DBTransactionError")
			time.Sleep(time.Duration(sleep_time) * time.Millisecond)
		}
	}
	//

	if txnErr != nil {
		return &status.DBResponse{
			Status:    status.DBStatus.Error,
			Message:   "Failed to commit transaction with error: " + txnErr.Error(),
			ErrorCode: enum.TRANSACTION_ABORTED,
		}
	}
	return &status.DBResponse{
		Status:  status.DBStatus.Ok,
		Data:    results,
		Message: "Transaction has been committed successfully",
	}
}
