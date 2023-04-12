package mongo_test

import (
	"testing"
	"time"

	"github.com/goonma/sdk/db/mongo"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ValidTimeseriesModel struct {
	ID         *primitive.ObjectID    `json:"_id" bson:"_id"`
	ClosePrice float64                `json:"close_price" bson:"close_price"`
	Metadata   map[string]interface{} `json:"metadata" bson:"metadata" timeseries:"meta_field"`
	Timestamp  *time.Time             `json:"timestamp" bson:"timestamp" timeseries:"time_field,hours"`
}

type InValidTimeseriesModel struct {
	ID *primitive.ObjectID `json:"_id" bson:"_id"`
}

type InValidTimeseriesMissingTimestampModel struct {
	ID       *primitive.ObjectID    `json:"_id" bson:"_id"`
	Metadata map[string]interface{} `json:"metadata" bson:"metadata" timeseries:"meta_field"`
}

type InValidTimeseriesWithInvalidGranularityModel struct {
	ID        *primitive.ObjectID `json:"_id" bson:"_id"`
	Timestamp *time.Time          `json:"timestamp" bson:"timestamp" timeseries:"time_field,invalid_granularity"`
}

type InValidTimeseriesWithMultipleTimeFieldModel struct {
	ID               *primitive.ObjectID `json:"_id" bson:"_id"`
	Timestamp        *time.Time          `json:"timestamp" bson:"timestamp" timeseries:"time_field"`
	AnotherTimestamp *time.Time          `json:"another_timestamp" bson:"another_timestamp" timeseries:"time_field"`
}

func TestGetTimeSeriesOptions(t *testing.T) {
	tcs := []struct {
		name     string
		model    interface{}
		hasError bool
	}{
		{"given a valid timeseries struct should return no error", ValidTimeseriesModel{ClosePrice: 99999}, false},
		{"given a valid timeseries pointer of struct should return no error", &ValidTimeseriesModel{ClosePrice: 99999}, false},
		{"given a nil should return an error", nil, true},
		{"given an invalid timeseries model should return error", InValidTimeseriesModel{}, true},
		{"given an invalid timeseries model with missing time field should return error", InValidTimeseriesMissingTimestampModel{}, true},
		{"given an invalid timeseries model with invalid granularity value should return error", InValidTimeseriesWithInvalidGranularityModel{}, true},
		{"given an invalid timeseries model with multiple time field should return error", InValidTimeseriesWithMultipleTimeFieldModel{}, true},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			_, err := mongo.GetTimeSeriesOptions(tc.model)
			if tc.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
