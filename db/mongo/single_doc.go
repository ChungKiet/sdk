package mongo

import (
	"context"
	"github.com/goonma/sdk/db/mongo/enum"
	"github.com/goonma/sdk/db/mongo/status"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"reflect"
	"time"
)

func (m *Collection) Create(ctx context.Context, entity interface{}) *status.DBResponse {
	colErr := m.checkCollection()
	if colErr != nil {
		return colErr
	}

	// convert to bson
	obj, err := m.convertToBson(entity)
	if err != nil {
		return &status.DBResponse{
			Status:    status.DBStatus.Error,
			Message:   "DB error: Create - Cannot convert object - " + err.Error(),
			ErrorCode: enum.MAP_OBJECT_FAILED,
		}
	}

	if obj[enum.CREATED_TIME] == nil {
		obj[enum.CREATED_TIME] = time.Now()
	}
	if obj[enum.UPDATED_TIME] == nil {
		obj[enum.UPDATED_TIME] = obj[enum.CREATED_TIME]
	}

	// insert
	if ctx == nil {
		ctx = context.TODO()
	}
	result, err := m.col.InsertOne(ctx, obj)
	if err != nil {
		return &status.DBResponse{
			Status:  status.DBStatus.Error,
			Message: "DB Error: " + err.Error(),
		}
	}

	obj["_id"] = result.InsertedID
	entity, _ = m.convertToObject(obj)
	list := m.newList(1)
	listValue := reflect.Append(reflect.ValueOf(list), reflect.Indirect(reflect.ValueOf(entity)))

	return &status.DBResponse{
		Status:  status.DBStatus.Ok,
		Message: "Create " + m.ColName + " successfully",
		Data:    listValue.Interface(),
	}

}

func (m *Collection) Query(ctx context.Context, query interface{}, opts ...*options.FindOneOptions) *status.DBResponse {
	colErr := m.checkCollection()
	if colErr != nil {
		return colErr
	}

	// convert to bson
	bFilter, err := m.convertToBson(query)
	if err != nil {
		return m.parseConversionError(err, "Query")
	}

	// find
	if ctx == nil {
		ctx = context.TODO()
	}
	result := m.col.FindOne(ctx, bFilter, opts...)
	if e := result.Err(); e != nil {
		return m.parseError(e, "Query", enum.QUERY_FAILED)
	}

	return m.parseSingleResult(result, "Query")
}

func (m *Collection) Update(ctx context.Context, query interface{}, updater interface{}, opts ...*options.FindOneAndUpdateOptions) *status.DBResponse {
	colErr := m.checkCollection()
	if colErr != nil {
		return colErr
	}

	// convert updater to bson
	bUpdater, err1 := m.convertToBson(updater)
	if err1 != nil {
		return m.parseConversionError(err1, "Update")
	}
	bUpdater[enum.UPDATED_TIME] = time.Now()

	// convert query to bson
	bFilter, err2 := m.convertToBson(query)
	if err2 != nil {
		return m.parseConversionError(err2, "Update")
	}

	if opts == nil {
		opts = []*options.FindOneAndUpdateOptions{
			{
				ReturnDocument: &after,
			},
		}
	}

	// update
	if ctx == nil {
		ctx = context.TODO()
	}
	result := m.col.FindOneAndUpdate(ctx, bFilter, bson.M{"$set": bUpdater}, opts...)
	if e := result.Err(); e != nil {
		return m.parseError(e, "Update", enum.UPDATE_FAILED)
	}

	return m.parseSingleResult(result, "Update")
}

func (m *Collection) Upsert(ctx context.Context, query interface{}, updater interface{}) *status.DBResponse {
	colErr := m.checkCollection()
	if colErr != nil {
		return colErr
	}

	// convert updater to bson
	bUpdater, err1 := m.convertToBson(updater)
	if err1 != nil {
		return m.parseConversionError(err1, "Upsert")
	}
	bUpdater[enum.UPDATED_TIME] = time.Now()
	createdTime, ok := bUpdater[enum.CREATED_TIME]
	if !ok || createdTime == nil {
		createdTime = bUpdater[enum.UPDATED_TIME]
	}
	delete(bUpdater, enum.CREATED_TIME)
	delete(bUpdater, "_id")

	// convert query to bson
	bFilter, err2 := m.convertToBson(query)
	if err2 != nil {
		return m.parseConversionError(err2, "Upsert")
	}

	// update
	if ctx == nil {
		ctx = context.TODO()
	}
	opts := []*options.FindOneAndUpdateOptions{
		{
			ReturnDocument: &after,
			Upsert:         &t,
		},
	}
	result := m.col.FindOneAndUpdate(ctx, bFilter, bson.M{
		"$set": bUpdater,
		"$setOnInsert": bson.M{
			"created_time": createdTime,
		},
	}, opts...)
	if e := result.Err(); e != nil {
		return m.parseError(e, "Upsert", enum.UPDATE_FAILED)
	}

	return m.parseSingleResult(result, "Update")
}

func (m *Collection) Increase(ctx context.Context, query interface{}, fieldName string, value int) *status.DBResponse {
	colErr := m.checkCollection()
	if colErr != nil {
		return colErr
	}

	// init updater
	updater := bson.M{
		"$inc": bson.D{
			{fieldName, value},
		},
	}

	// convert query to bson
	bFilter, err := m.convertToBson(query)
	if err != nil {
		return m.parseConversionError(err, "Incre")
	}

	// update
	if ctx == nil {
		ctx = context.TODO()
	}
	opt := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
		Upsert:         &t,
	}
	result := m.col.FindOneAndUpdate(ctx, bFilter, updater, &opt)
	if e := result.Err(); e != nil {
		return m.parseError(e, "Increase", enum.UPDATE_FAILED)
	}

	return m.parseSingleResult(result, "Increase "+fieldName+" of")
}

func (m *Collection) Delete(ctx context.Context, query interface{}, opts ...*options.DeleteOptions) *status.DBResponse {
	colErr := m.checkCollection()
	if colErr != nil {
		return colErr
	}

	// convert query to bson
	bFilter, err := m.convertToBson(query)
	if err != nil {
		return m.parseConversionError(err, "Delete")
	}

	// delete
	if ctx == nil {
		ctx = context.TODO()
	}
	result, err := m.col.DeleteOne(ctx, bFilter, opts...)
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

func (m *Collection) Replace(ctx context.Context, query interface{}, replacement interface{}, opts ...*options.FindOneAndReplaceOptions) *status.DBResponse {
	colErr := m.checkCollection()
	if colErr != nil {
		return colErr
	}

	// convert replacement to bson
	bReplacement, err1 := m.convertToBson(replacement)
	if err1 != nil {
		return m.parseConversionError(err1, "Replace")
	}

	if bReplacement[enum.CREATED_TIME] == nil {
		bReplacement[enum.CREATED_TIME] = time.Now()
	}
	bReplacement[enum.UPDATED_TIME] = bReplacement[enum.CREATED_TIME]

	// convert query to bson
	bFilter, err2 := m.convertToBson(query)
	if err2 != nil {
		return m.parseConversionError(err2, "Replace")
	}

	// replace
	if ctx == nil {
		ctx = context.TODO()
	}
	result := m.col.FindOneAndReplace(ctx, bFilter, bReplacement, opts...)
	if e := result.Err(); e != nil {
		return m.parseError(e, "Replace", enum.UPDATE_FAILED)
	}

	return m.parseSingleResult(result, "Update")
}
