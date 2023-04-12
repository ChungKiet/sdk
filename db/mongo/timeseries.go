package mongo

import (
	"fmt"
	"reflect"

	"github.com/goonma/sdk/utils"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	StructTagTimeseries = "timeseries"
	DefaultGranularity  = "hours" // available values: seconds, minutes, hours
)

var (
	validGranularityValues = []string{"seconds", "minutes", "hours"}
)

// Please check unittest for user guide
//
// MongoDB Refs:
// https://www.mongodb.com/docs/manual/core/timeseries/timeseries-procedures/
//
// Example:
//
//	type ValidTimeseriesModel struct {
//		ID *primitive.ObjectID `json:"_id" bson:"_id"`
//		ClosePrice float64 `json:"close_price" bson:"close_price"` // Measurement
//		Metadata map[string]interface{} `json:"metadata" bson:"metadata" timeseries:"meta_field"` // Metadata
//		Timestamp *time.Time `json:"timestamp" bson:"timestamp" timeseries:"time_field,hours"` // Time field
//	}
func GetTimeSeriesOptions(model interface{}) (*options.TimeSeriesOptions, error) {
	var err error

	if model == nil {
		return nil, fmt.Errorf("input model should not be nil")
	}

	// validate template object type
	modelKind := reflect.TypeOf(model).Kind()
	switch modelKind {
	case reflect.Struct:
	case reflect.Pointer:
		modelVal := reflect.ValueOf(model)
		val := modelVal.Elem().Interface()
		return GetTimeSeriesOptions(val)
	default:
		return nil, fmt.Errorf("input model should be a struct or pointer of struct, current %v", modelKind)
	}

	fields := reflect.VisibleFields(reflect.TypeOf(model))
	tseriesFields := map[string][]string{
		"time_field":  []string{},
		"meta_field":  []string{},
		"granularity": []string{},
	}

	for _, f := range fields {
		bsonKey := getKeysFromTag(f.Tag.Get("bson"))
		// ignore field tag
		if len(bsonKey) == 0 || bsonKey[0] == "-" {
			continue
		}

		tsKey := getKeysFromTag(f.Tag.Get(StructTagTimeseries))
		// ignore field tag
		if len(tsKey) == 0 || tsKey[0] == "-" {
			continue
		}

		// just check valid tag values
		if _, ok := tseriesFields[tsKey[0]]; ok {
			tseriesFields[tsKey[0]] = append(tseriesFields[tsKey[0]], bsonKey[0])
			if tsKey[0] == "time_field" && len(tsKey) == 2 {
				tseriesFields["granularity"] = append(tseriesFields["granularity"], tsKey[1])
			}

			if tsKey[0] == "meta_field" && f.Type.Kind() == reflect.Slice {
				return nil, fmt.Errorf("meta field cannot be a slice or array")
			}
		}
	}

	opts := options.TimeSeries()
	granularity := DefaultGranularity

	if len(tseriesFields["time_field"]) != 1 {
		return nil, fmt.Errorf("time field is required and shouldn't have more than one, count: %v", len(tseriesFields["time_field"]))
	}

	if len(tseriesFields["meta_field"]) != 1 {
		return nil, fmt.Errorf("meta field have more than one, count: %v", len(tseriesFields["meta_field"]))
	}

	if len(tseriesFields["granularity"]) == 1 {
		granularity = tseriesFields["granularity"][0]
		if !utils.Contains(validGranularityValues, granularity) {
			return nil, fmt.Errorf("invalid granularity option for time field")
		}
	}

	opts.SetTimeField(tseriesFields["time_field"][0]).
		SetGranularity(granularity)

	if len(tseriesFields["meta_field"]) > 0 {
		opts.SetMetaField(tseriesFields["meta_field"][0])
	}

	return opts, err
}

func getKeysFromTag(tag string) []string {
	pieces := utils.StringSlice(tag, ",")
	if len(pieces) == 0 {
		return []string{}
	}
	return pieces
}
