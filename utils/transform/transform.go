package transform
import (
	"github.com/goonma/sdk/db/mongo/model"
	micropb "github.com/goonma/sdk/base/pb/micro"
	structpb "google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/encoding/protojson"
	ev "github.com/goonma/sdk/base/event"
	e "github.com/goonma/sdk/base/error"
	"encoding/json"
	"fmt"
	"errors"
)
func PBToQueryBuilder(req *micropb.MicroRequest) (*model.QueryBuilder, error) {
	data := req.GetData()
	byteData, _ := data.MarshalJSON()

	queryBuilder := &model.QueryBuilder{}
	err := json.Unmarshal(byteData, queryBuilder)
	if err != nil {
		return nil, err
	}

	if queryBuilder.Limit == 0 {
		queryBuilder.Limit = 1000
	}
	return queryBuilder, nil
}
func MapToPBStruct(m map[string]interface{}) (*structpb.Struct,error){
	b, err := json.Marshal(m)
	if err!=nil{
		return nil,err
	}
	s:=&structpb.Struct{}
	err = protojson.Unmarshal(b, s)
	return s,err
}
func IToPBStruct(m interface{}) (*structpb.Struct, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	s := &structpb.Struct{}
	err = protojson.Unmarshal(b, s)
	return s, err
}
func EventToByteArray(event ev.Event) ([]byte,*e.Error){
	data, err := json.Marshal(event)
    if err != nil {
        //log.Error(err.Error(),"EVENT_DRIVEN_SERIALIZE")
		return nil,e.New(err.Error(),"EVENT_TRANSFORM")
    }
	return data,nil
}
//
func elabValue(value *structpb.Value) (interface{}, error) {
	var err error
	if value == nil {
		return nil, nil
	}
	if structValue, ok := value.GetKind().(*structpb.Value_StructValue); ok {
		result := make(map[string]interface{})
		for k, v := range structValue.StructValue.Fields {
			result[k], err = elabValue(v)
			if err != nil {
				return nil, err
			}
		}
		return result, err
	}
	if listValue, ok := value.GetKind().(*structpb.Value_ListValue); ok {
		result := make([]interface{}, len(listValue.ListValue.Values))
		for i, el := range listValue.ListValue.Values {
			result[i], err = elabValue(el)
			if err != nil {
				return nil, err
			}
		}
		return result, err
	}
	if _, ok := value.GetKind().(*structpb.Value_NullValue); ok {
		return nil, nil
	}
	if numValue, ok := value.GetKind().(*structpb.Value_NumberValue); ok {
		return numValue.NumberValue, nil
	}
	if strValue, ok := value.GetKind().(*structpb.Value_StringValue); ok {
		return strValue.StringValue, nil
	}
	if boolValue, ok := value.GetKind().(*structpb.Value_BoolValue); ok {
		return boolValue.BoolValue, nil
	}
	return errors.New(fmt.Sprintf("Cannot convert the value %+v", value)), nil
}

func PBStruct2Map(str *structpb.Struct) (map[string]interface{}, error) {
	var err error
	result := make(map[string]interface{})
	for k, v := range str.Fields {
		result[k], err = elabValue(v)
		if err != nil {
			return nil, err
		}
	}
	return result, err
}
func StructToJson(v interface{}) string{
	if v==nil{
		return ""
	}
	out, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return (string(out))
}