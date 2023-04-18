package main
import(
	"github.com/goonma/sdk/jwt"
	"fmt"
)
func main(){
	
	key:="UUzzvDnsvNXaWTsHVgh3T6dQmu2yXQ2c9FFkccNhu6QpdE85t5yZbWUJLsqpt5cN"
	user_id:="641727e8b47cd32bc93e0d6a"
	role_id:="1"
	username:="quoc Danh"
	email:="quoc"
	expired:=999999999
	tkn,err:=jwt.GenerateJWTToken(key,user_id,role_id,username,email,0,expired)
	fmt.Println(err)
	fmt.Println(tkn)
	/*tkn_string:="eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJ1c2VyaWQiOiI2NDE2ZDQ3ZmI0N2NkMzJiYzkzZTBkNWIiLCJyb2xlaWQiOjAsInVzZXJuYW1lIjoiMSIsImVtYWlsIjoic29uaG9hbmciLCJ0eXBlIjoic3Vob2FuZ3NvbkBnbWFpbC5jb20iLCJpc3MiOiJHb29ubWEuY29tIiwiZXhwIjoxNjg5NDExNjYyfQ.x_UYCzEOFDNpLBvCnE-yMhvbHWJd-RDOsAKQ1wKo9SdTJzUXxewEjVM1lPXt162NCNmIIUpj7i5AMO3EL7ENkg"
	claim,err:=jwt.VerifyJWTToken(key,tkn_string)
	fmt.Println(err)
	fmt.Printf("%+v",claim)*/
}