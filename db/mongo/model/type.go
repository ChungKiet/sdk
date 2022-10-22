package model

import "go.mongodb.org/mongo-driver/bson"

type QueryBuilder struct {
	Query   interface{} `json:"query,omitempty" bson:"query,omitempty"`
	Updater interface{} `json:"updater,omitempty" bson:"updater,omitempty"`

	Limit         int64   `json:"limit,omitempty"`
	Offset        int64   `json:"offset,omitempty"`
	SortFields    *bson.M `json:"sort_fields,omitempty" bson:"sort_fields,omitempty"`
	DistinctField string  `json:"distinct_field,omitempty"`
}
