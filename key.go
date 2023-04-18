package main
import (
	"github.com/goonma/sdk/encrypt"
	"fmt"
)
func main(){
    data:="D8H7ItKNMYlIqm2JksMT54UkBlfD/Yeb+KkFfLYwc495U/HmOMg1JDAtGPk0C/+LzYd/MDa0JVJouPh6TyTlhBW5ZAGnmSAoz1vDFG2oVbg9dOhUgBuK8r1Kp1w6hHPJdPKrhKSaKXrj/5tS9y04t3AVVW4ew4w1AlTeSx2dstwYEGk7ukq6EmxjDzUGN8yvI1+fs/NjK3SNpV7zHN3PQRFRNKsdVJ/QzVeE3MDzENzch1hw8uZfPAuP+caEuX4NXiWHKRVFGWzLXQxnmhsKDVXvVX5vufzKexbTI/pEBdFoPQSer4/71HiWjmiWXNFmMWX4i0L35DYsCQBCniy1Bw=="
	
	decrypt_text,err:=encrypt.RSA_OAEP_Decrypt(data,pri_key)
	fmt.Println(err)
	fmt.Println(decrypt_text)
}