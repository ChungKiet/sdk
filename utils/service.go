package utils
import (
	"errors"
)
func GetServiceName(eventName string) (string,error){
	arr:=Explode(eventName,"|")
	if len(arr)==0 {
		return "",errors.New("Service name is empty")
	}
	if arr[0]==""{
		return "",errors.New("Service name is empty")
	}
	return arr[0],nil
}
func  GetServiceMethod(eventName string) (string,error){
	arr:=Explode(eventName,"|")
	if len(arr)<2 {
		return "",errors.New("Service Method is empty")
	}
	if arr[1]==""{
		return "",errors.New("Service Method is empty")
	}
	return arr[1],nil
}