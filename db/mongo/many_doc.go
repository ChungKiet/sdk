package mongo

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/goonma/sdk/db/mongo/enum"
	"github.com/goonma/sdk/db/mongo/status"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	//"fmt"
)

func (m *Collection) Aggregate(pipeline interface{}, result interface{}) *status.DBResponse {
	colErr := m.checkCollection()
	if colErr != nil {
		return colErr
	}

	cursor, err := m.col.Aggregate(context.TODO(), pipeline)
	if err != nil {
		return &status.DBResponse{
			Status:    status.DBStatus.Error,
			Message:   "DB error: Aggregate - " + err.Error(),
			ErrorCode: enum.AGGREGATE_FAILED,
		}
	}

	err = cursor.All(context.TODO(), result)
	if err != nil {
		return &status.DBResponse{
			Status:    status.DBStatus.Error,
			Message:   "DB error: Aggregate - " + err.Error(),
			ErrorCode: enum.AGGREGATE_FAILED,
		}
	}

	return &status.DBResponse{
		Status:  status.DBStatus.Ok,
		Message: "Aggregate documents successfully",
	}
}

func (m *Collection) Distinct(filter interface{}, field string, opt ...*options.DistinctOptions) *status.DBResponse {
	colErr := m.checkCollection()
	if colErr != nil {
		return colErr
	}

	// convert query to bson
	bFilter, err := m.convertToBson(filter)
	if err != nil {
		return m.parseConversionError(err, "Distinct")
	}

	result, err := m.col.Distinct(context.TODO(), field, bFilter, opt...)
	if err != nil {
		return &status.DBResponse{
			Status:    status.DBStatus.Error,
			Message:   "DB error: Distinct " + err.Error(),
			ErrorCode: enum.DISTINCT_FAILED,
		}
	}
	return &status.DBResponse{
		Status:  status.DBStatus.Ok,
		Data:    result,
		Message: "Distinct documents successfully",
		Total:   int64(len(result)),
	}
}

func (m *Collection) Count(ctx context.Context, query interface{}, opts ...*options.CountOptions) *status.DBResponse {
	colErr := m.checkCollection()
	if colErr != nil {
		return colErr
	}

	// convert query to bson
	bFilter, err := m.convertToBson(query)
	if err != nil {
		return m.parseConversionError(err, "Count")
	}

	if ctx == nil {
		context.TODO()
	}
	total, err := m.col.CountDocuments(ctx, bFilter, opts...)
	if err != nil {
		return &status.DBResponse{
			Status:    status.DBStatus.Error,
			Message:   "Count error: " + err.Error(),
			ErrorCode: enum.COUNT_FAILED,
		}
	}
	return &status.DBResponse{
		Status:  status.DBStatus.Ok,
		Message: "Count documents successfully",
		Total:   total,
	}
}

func (m *Collection) QueryMany(ctx context.Context, query interface{}, offset int64, limit int64, sortFields *bson.M) *status.DBResponse {
	result, err := m.getQueryManyCursor(ctx, query, offset, limit, sortFields)
	if err != nil {
		return err
	}

	// !!! don't use the given context in this step
	// convert output to object
	list := m.newList(int(limit))
	//fmt.Printf("len List: %+v\r\n",list)
	//fmt.Printf("Before List: %+v\r\n",list)
	mgoErr := result.All(context.TODO(), &list)
	result.Close(context.TODO())
	//fmt.Printf("After List: %+v\r\n",list)
	if mgoErr != nil {
		return &status.DBResponse{
			Status:    status.DBStatus.Error,
			Message:   "DB error: " + mgoErr.Error(),
			ErrorCode: enum.QUERY_FAILED,
		}
	}
	count := reflect.ValueOf(list).Len()
	//fmt.Printf("Count: %+v\r\n",count)
	if count == 0 {
		return &status.DBResponse{
			Status:    status.DBStatus.NotFound,
			Message:   "Not found any matched " + m.ColName,
			ErrorCode: enum.NO_DOCUMENT_FOUND,
		}
	}
	return &status.DBResponse{
		Status:  status.DBStatus.Ok,
		Message: "Query " + m.ColName + " successfully",
		Data:    list,
		Total:   int64(count),
	}
}

// For stream result
func (m *Collection) getQueryManyCursor(ctx context.Context, query interface{}, offset int64, limit int64, sortFields *bson.M) (*mongo.Cursor, *status.DBResponse) {
	colErr := m.checkCollection()
	if colErr != nil {
		return nil, colErr
	}

	opt := &options.FindOptions{}
	k := int64(1000000000)
	if limit <= 0 {
		opt.Limit = &k
	} else {
		opt.Limit = &limit
	}
	if offset > 0 {
		opt.Skip = &offset
	}
	if sortFields != nil {
		opt.Sort = sortFields
	}

	// convert to bson
	bFilter, err := m.convertToBson(query)
	//fmt.Printf("convertToBson Filter: %+v\r\n",bFilter)
	if err != nil {
		return nil, m.parseConversionError(err, "QueryMany")
	}

	// find
	if ctx == nil {
		ctx = context.TODO()
	}

	resultCursor, e := m.col.Find(ctx, bFilter, opt)
	if e != nil {
		return nil, m.parseError(e, "QueryMany", enum.QUERY_FAILED)
	}
	if e := resultCursor.Err(); e != nil {
		return nil, m.parseError(e, "QueryMany", enum.QUERY_FAILED)
	}
	return resultCursor, nil
}

func (m *Collection) QueryManyWithCursorCallback(ctx context.Context, query interface{}, offset int64, limit int64, sortFields *bson.M, cursorCallback func(val interface{})) error {
	result, err := m.getQueryManyCursor(ctx, query, offset, limit, sortFields)
	if err != nil {
		return fmt.Errorf(err.Message)
	}

	defer result.Close(ctx)
	for result.Next(ctx) {
		data := bson.M{}
		result.Decode(&data)
		entity, err := m.convertToObject(data)
		if err != nil {
			return err
		}
		cursorCallback(reflect.Indirect(reflect.ValueOf(entity)).Interface())
	}

	return nil
}

func (m *Collection) UpdateMany(ctx context.Context, query interface{}, updater interface{}) *status.DBResponse {
	colErr := m.checkCollection()
	if colErr != nil {
		return colErr
	}

	// convert updater to bson
	bUpdater, err1 := m.convertToBson(updater)
	if err1 != nil {
		return m.parseConversionError(err1, "UpdateMany")
	}
	bUpdater[enum.UPDATED_TIME] = time.Now()

	// convert query to bson
	bFilter, err2 := m.convertToBson(query)
	if err2 != nil {
		return m.parseConversionError(err2, "UpdateMany")
	}

	// update
	if ctx == nil {
		ctx = context.TODO()
	}
	result, e := m.col.UpdateMany(ctx, bFilter, bson.M{
		"$set": bUpdater,
	})
	if e != nil {
		return &status.DBResponse{
			Status:    status.DBStatus.Error,
			Message:   "DB error: Update - " + e.Error(),
			ErrorCode: enum.UPDATE_FAILED,
		}
	}

	if result.MatchedCount == 0 {
		return &status.DBResponse{
			Status:    status.DBStatus.NotFound,
			Message:   "Not found any matched " + m.ColName,
			ErrorCode: enum.NO_DOCUMENT_FOUND,
		}
	}

	return &status.DBResponse{
		Status:  status.DBStatus.Ok,
		Message: "Update " + m.ColName + " successfully",
		Total:   result.ModifiedCount,
	}
}

func (m *Collection) CreateMany(ctx context.Context, entityList interface{}) *status.DBResponse {
	colErr := m.checkCollection()
	if colErr != nil {
		return colErr
	}

	list, err := m.interfaceSlice(entityList)
	if err != nil {
		return &status.DBResponse{
			Status:    status.DBStatus.Error,
			Message:   "DB error: Create many - Invalid input data",
			ErrorCode: enum.INVALID_SLICE,
		}
	}

	// convert to bson
	var bsonList []interface{}
	now := time.Now()
	for _, item := range list {
		obj, e := m.convertToBson(item)
		if e != nil {
			return m.parseConversionError(e, "CreateMany")

		}
		if obj[enum.CREATED_TIME] == nil {
			obj[enum.CREATED_TIME] = now
		}
		if obj[enum.UPDATED_TIME] == nil {
			obj[enum.UPDATED_TIME] = obj[enum.CREATED_TIME]
		}
		bsonList = append(bsonList, obj)
	}

	// insert
	if ctx == nil {
		ctx = context.TODO()
	}
	result, err := m.col.InsertMany(ctx, bsonList)
	if err != nil {
		return &status.DBResponse{
			Status:    status.DBStatus.Error,
			Message:   "DB Error: " + err.Error(),
			ErrorCode: enum.CREATE_FAILED,
		}
	}

	return &status.DBResponse{
		Status:  status.DBStatus.Ok,
		Message: "Create " + m.ColName + " successfully",
		Data:    result.InsertedIDs,
	}
}

func (m *Collection) DeleteMany(ctx context.Context, query interface{}) *status.DBResponse {
	colErr := m.checkCollection()
	if colErr != nil {
		return colErr
	}

	// convert query to bson
	bFilter, err := m.convertToBson(query)
	if err != nil {
		return m.parseConversionError(err, "DeleteMany")
	}

	if ctx == nil {
		ctx = context.TODO()
	}
	result, err := m.col.DeleteMany(ctx, bFilter)
	if err != nil {
		return &status.DBResponse{
			Status:    status.DBStatus.Error,
			Message:   "Delete error: " + err.Error(),
			ErrorCode: enum.DELETE_FAILED,
		}
	}
	return &status.DBResponse{
		Status:  status.DBStatus.Ok,
		Message: "Delete " + m.ColName + " successfully",
		Total:   result.DeletedCount,
	}
}
