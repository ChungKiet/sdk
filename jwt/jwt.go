package jwt
import(
	"github.com/golang-jwt/jwt/v4"
	"time"
	"fmt"
)

type CustomClaims struct {
	UserID string `json:"userid"`
	RoleID string `json:"roleid"`
	Username string `json:"username"`
	Email string  `json:"email"`
	Type string  `json:"type"`
	jwt.RegisteredClaims
}
func GenerateJWTToken(key_sign,user_id,role_id,username,email,ttype string,expired int) (string,error){
	signingKey := []byte(key_sign)
	// Create the claims
	claims := CustomClaims{
		user_id,
		role_id,
		username,
		email,
		ttype,
		jwt.RegisteredClaims{
			// A usual scenario is to set the expiration time relative to the current time
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expired) *  time.Second)),			
			Issuer:    "Goonma.com",
		},
	}
	//fmt.Printf("%+v\r\n",claims)
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	res, err := token.SignedString(signingKey)
	if err!=nil{
		return "",err
	}
	return res,nil
}
func VerifyJWTToken(key,token_string string) (*CustomClaims,error){
	token, err := jwt.ParseWithClaims(token_string, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(key), nil
	})
	claim, ok := token.Claims.(*CustomClaims); 
	//fmt.Printf("%+v\r\n",token.Claims)
	if ok && token.Valid {
		return claim,nil
	} 
	return nil,err
}