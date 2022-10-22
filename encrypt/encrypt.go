package  encrypt
import(
	"golang.org/x/crypto/bcrypt"
	"crypto/md5"
    "encoding/hex"
	"encoding/base64"
)
//hash function
func HashBcrypt(password string) (string, error) {
	//bcrypt.GenerateFromPassword is auto generate Salt
	hash,err:=bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return  string(hash),err
}
//check plain text string same with hash string
func VerifyHashBcrypt(hashedText, plainText string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedText), []byte(plainText))
}
func HashMD5(text string) string {
    hasher := md5.New()
    hasher.Write([]byte(text))
    return hex.EncodeToString(hasher.Sum(nil))
}
func Base64(data string) string{
	return base64.StdEncoding.EncodeToString([]byte(data))
}