package main
import(
	"github.com/goonma/sdk/jwt"
	"fmt"
)
func main(){
	
	key:="isiuriwruweirmuriewcr94829420"
	/*user_id:="1"
	role_id:="2"
	username:="sonhoang"
	email:="suhoangson@gmail.com"
	expired:=15
	tkn,err:=jwt.GenerateJWTToken(key,user_id,role_id,username,email,expired)
	fmt.Println(err)
	fmt.Println(tkn)*/
	tkn_string:="eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJ1c2VyaWQiOiIxIiwicm9sZWlkIjoiMiIsInVzZXJuYW1lIjoic29uaG9hbmciLCJlbWFpbCI6InN1aG9hbmdzb25AZ21haWwuY29tIiwiaXNzIjoiR29vbm1hLmNvbSIsImV4cCI6MTY2NjUwMjA2MH0.JtuQAnTBAqs4IsBBsGfrY8peuo8YYMLQreHe4v4HOK8Uq-H2omNrzdNV0T_ddFRfq7-V8-l6UKawJD2hdSSLKA"
	claim,err:=jwt.VerifyJWTToken(key,tkn_string)
	fmt.Println(err)
	fmt.Printf("%+v",claim)
}