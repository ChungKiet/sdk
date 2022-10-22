package utils
import (
	"fmt"
	"reflect"
	"strings"
	"crypto/rand"
	"encoding/json"
	"strconv"
	"errors"
	"math"
)
//check string item exist in array string
func Contains(arr []string, item string) bool {
	for _, e := range arr {
		if strings.ToLower(item) == strings.ToLower(e) {
			return true
		}
	}
	return false
}
//check int item exist in array int
func ContainsInt(arr []int, item int) bool {
	for _, e := range arr {
		if item == e {
			return true
		}
	}
	return false
}
//check map string contains key
func Map_contains(m map[string]string, item string) bool {
	if len(m)==0{
		return false
	}
	if _, ok := m[item]; ok {
		return true
	}
	return false
}
//check map interface contains key
func MapI_contains(m map[string]interface{}, item string) bool {
	if len(m)==0{
		return false
	}
	if _, ok := m[item]; ok {
		return true
	}
	return false
}
//Generate token
func TokenGenerator(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
//convert interface to bool
func ItoBool(value interface{}) (bool,error) {
	if value==nil{
		return false,errors.New("Empty value")
	}
	v, err := value.(bool)
	if !err{
		return  false,errors.New("Error convert Interface to bool")
	}
	return v,nil
}
//Convert Interface to String
func ItoString(value interface{}) string {
	if value==nil{
		return ""
	}
	str := fmt.Sprintf("%v", value)
	return str
}
//Convert Interface to Integer
func ItoInt(value interface{}) int{
	if value==nil{
		return math.MinInt32
	}
	e := reflect.ValueOf(value)
	//e reflect.Value
	switch e.Type().Kind() {
		case reflect.String:
			r, err :=strconv.Atoi(value.(string))
			if err!=nil{
			 	return math.MinInt32
			}
			return r
		case reflect.Float64:
			f,err:=GetFloat(value)
			if err!=nil{
				return math.MinInt32
			}
			r:=int(f)
			return r
		case reflect.Int,reflect.Int32,reflect.Int64:
			return StringToInt(ItoString(value))
		default:
			return math.MinInt32
	}

}
//Convert Interface to Integer
func ItoInt64(value interface{}) int64{
	if value==nil{
		return math.MinInt64
	}
	e := reflect.ValueOf(value)
	//e reflect.Value
	switch e.Type().Kind() {
		case reflect.String:
			r, err :=strconv.ParseInt(value.(string), 10, 64)
			if err!=nil{
			 	return math.MinInt64
			}
			return r
		case reflect.Float64:
			f,err:=GetFloat(value)
			if err!=nil{
				return math.MinInt64
			}
			r:=int64(f)
			return r
		case reflect.Int,reflect.Int32,reflect.Int64:
			return int64(StringToInt(ItoString(value)))
		default:
			return math.MinInt64
	}

}
//Convert Interface to slice Interface
func ItoSlice(value interface{}) ([]interface{},error){
	s := reflect.ValueOf(value)
	if s.Kind() != reflect.Slice {
		return nil,errors.New("Object not is slice")
	}

	ret := make([]interface{}, s.Len())

	for i:=0; i<s.Len(); i++ {
		ret[i] = s.Index(i).Interface()
	}
	return ret,nil
}
//Convert interface to Slice int
func ItoSliceInt(value interface{}) ([]int,error){
	arr,err:= ItoSlice(value)
	if err!=nil{
		return nil,err 
	}
	res:=[]int{}
	for _,v:=range(arr){
		res=append(res,ItoInt(v))
	}
	return res,nil
}
//convert Iinterface to slice string
func ItoSliceString(value interface{}) ([]string,error){
	arr,err:= ItoSlice(value)
	if err!=nil{
		return nil,err 
	}
	res:=[]string{}
	for _,v:=range(arr){
		res=append(res,ItoString(v))
	}
	return res,nil
}
//convert interface to Dictionary
func ItoDictionary(value interface{}) (map[string]interface{},error){
	whereType := reflect.TypeOf(value).Kind()
	switch whereType {
	case reflect.Map:
		row_map := value.(map[string]interface{})
		return row_map,nil
	default:
		return nil, errors.New("ItoDictionary=>_DATA_ROW_TYPE_NOT_SUPPORT_")
	}
}
//convert interface to Dictionary
func ItoDictionaryS(value interface{}) (map[string]string,error){
	whereType := reflect.TypeOf(value).Kind()
	switch whereType {
	case reflect.Map:
		row_map := value.(map[string]string)
		return row_map,nil
	default:
		return nil, errors.New("ItoDictionary=>_DATA_ROW_TYPE_NOT_SUPPORT_")
	}
}
//Convert array to string with separate char is ,
func ArrToS(arr []interface{})string{
	res:=""
	for _,v:=range arr{
		Type := reflect.TypeOf(v).Kind()
		switch Type {
			case reflect.Int,reflect.Int32, reflect.Int64:
				res+=ItoString(v)+","
			case reflect.String:
				res+="'"+ItoString(v)+"',"
			default:
				res+=""
		}
		
	}
	res=strings.TrimSuffix(res,",")
	return res
}
//convert array interface to arrage string
func ArrItoS(arr []interface{}) ([]string){
	var res []string
	for _,v:=range arr{
		Type := reflect.TypeOf(v).Kind()
		switch Type {
			case reflect.Int,reflect.Int32, reflect.Int64:
				res = append(res,ItoString(v))
			case reflect.String:
				res = append(res,ItoString(v))
		}
		
	}
	return res
}

//init dictionary interface
func Dictionary() map[string]interface{}{
	res:=make(map[string]interface{})
	return res
}
//init dictionary string
func DictionaryString() map[string]string{
	res:=make(map[string]string)
	return res
}
//Convert Int to String
func IntToS(v int) string{
	return strconv.Itoa(v)
}
func Int64ToS(v int64) string{
	return strconv.FormatInt(v, 10)
}
//return Type of interface as: 'int', 'string'...
func Type(t interface{}) string{
	objtype := reflect.TypeOf(t).Kind()
	switch objtype {
		case reflect.String:
			return "string"
		case reflect.Struct:
			return "struct"
		case reflect.Map:
			return "map"
		case reflect.Array,reflect.Slice:
			return "array"
		case reflect.Bool:
			return "bool"
		case reflect.Float32, reflect.Float64:
			return "float"
		case reflect.Int,reflect.Int32, reflect.Int64:
			return "int"
		default:
			return "undefine"
	}
}
//convert map string string to JSON string
func MapToJSONString(m map[string]string) (string,error) {
	jsonStr, err := json.Marshal(m)
	if err!=nil{
		return "",err 
	}
	return string(jsonStr),nil
}
//Convert String to Json
func StringToJSON(s string) interface{}{
	var result interface{}
	//var result []string
	json.Unmarshal([]byte(s), &result)
	return result
}


func IsEmptyValue(value interface{}) bool {
	e := reflect.ValueOf(value)
	//e reflect.Value
	is_show_checking := false
	var checking_type string
	is_empty := true
	switch e.Type().Kind() {
	case reflect.String:
		checking_type = "string"
		if e.String() != "" {
			// fmt.Println("Empty string")
			is_empty = false
		}
	case reflect.Array:
		checking_type = "array"
		for j := e.Len() - 1; j >= 0; j-- {
			is_empty = IsEmptyValue(e.Index(j))
			if is_empty == false {
				break
			}
			/*if e.Index(j).Float() != 0 {
				// fmt.Println("Empty float")
				is_empty = false
				break
			}*/
		}
	case reflect.Float32, reflect.Float64:
		checking_type = "float"
		if e.Float() != 0 {
			is_empty = false
		}
	case reflect.Int32, reflect.Int64:
		checking_type = "int"
		if e.Int() != 0 {
			is_empty = false

		}
	case reflect.Ptr:
		checking_type = "Ptr"
		if e.Pointer() != 0 {
			is_empty = false
		}
	case reflect.Struct:
		checking_type = "struct"
		for i := e.NumField() - 1; i >= 0; i-- {
			is_empty = IsEmptyValue(e.Field(i))
			// fmt.Println(e.Field(i).Type().Kind())
			if !is_empty {
				break
			}
		}
	default:
		checking_type = "default"
		// is_empty = IsEmptyStruct(e)
	}
	if is_show_checking {
		fmt.Println("Checking type :", checking_type, e.Type().Kind())
	}
	return is_empty
}
func GetFloat(unk interface{}) (float64, error) {
	var floatType = reflect.TypeOf(float64(0))
	v := reflect.ValueOf(unk)
	v = reflect.Indirect(v)
	if !v.Type().ConvertibleTo(floatType) {
		return 0, fmt.Errorf("cannot convert %v to float64", v.Type())
	}
	fv := v.Convert(floatType)
	return fv.Float(), nil
}
func StructExistField(t interface{},fieldName string) bool {
	metaValue := reflect.ValueOf(t).Elem()
	field := metaValue.FieldByName(fieldName)
	if field == (reflect.Value{}) {
		return false
	}
	return true
}