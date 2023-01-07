package encrypt
import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
    "crypto/x509"
    "encoding/pem"
	"errors"
	//"fmt"
)
func RSA_Generate_KEY(size int) (string,string,error){
	privateKey, err := rsa.GenerateKey(rand.Reader, size)
	if err!=nil{
		return "","",err
	}
	pri_str:=ExportRsaPrivateKeyAsStr(privateKey)
	publicKey := privateKey.PublicKey
	pub_str,err:=ExportRsaPublicKeyAsStr(&publicKey)
	if err!=nil{
		return "","",err 
	}
	return pri_str,pub_str,nil
}
func RSA_OAEP_Encrypt(secretMessage string, pub_key string) (string,error) {
	key,err:=ParseRsaPublicKeyFromStr(pub_key)
	if err!=nil{
		return "",err 
	}
	label := []byte("OAEP Encrypted")
	rng := rand.Reader
	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rng, key, []byte(secretMessage), label)
	if err!=nil{
		return "",err 
	}
	return base64.StdEncoding.EncodeToString(ciphertext),err
}

func RSA_OAEP_Decrypt(cipherText string,pri_key string) (string,error) {
	privKey,err:=ParseRsaPrivateKeyFromStr(pri_key)
	if err!=nil{
		return "",err 
	}
	ct, _ := base64.StdEncoding.DecodeString(cipherText)
	label := []byte("OAEP Encrypted")
	rng := rand.Reader
	plaintext, err := rsa.DecryptOAEP(sha256.New(), rng, privKey, ct, label)
	if err!=nil{
		return "",err 
	}
	return string(plaintext),nil
}



func ExportRsaPrivateKeyAsStr(privkey *rsa.PrivateKey) string {
    privkey_bytes := x509.MarshalPKCS1PrivateKey(privkey)
    privkey_pem := pem.EncodeToMemory(
            &pem.Block{
                    Type:  "RSA PRIVATE KEY",
                    Bytes: privkey_bytes,
            },
    )
    return string(privkey_pem)
}

func ParseRsaPrivateKeyFromStr(privPEM string) (*rsa.PrivateKey, error) {
    block, _ := pem.Decode([]byte(privPEM))
    if block == nil {
            return nil, errors.New("failed to parse PEM block containing the key")
    }

    priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
    if err != nil {
            return nil, err
    }

    return priv, nil
}

func ExportRsaPublicKeyAsStr(pubkey *rsa.PublicKey) (string, error) {
    pubkey_bytes, err := x509.MarshalPKIXPublicKey(pubkey)
    if err != nil {
            return "", err
    }
    pubkey_pem := pem.EncodeToMemory(
            &pem.Block{
                    Type:  "RSA PUBLIC KEY",
                    Bytes: pubkey_bytes,
            },
    )

    return string(pubkey_pem), nil
}

func ParseRsaPublicKeyFromStr(pubPEM string) (*rsa.PublicKey, error) {
    block, _ := pem.Decode([]byte(pubPEM))
    if block == nil {
            return nil, errors.New("failed to parse PEM block containing the key")
    }

    pub, err := x509.ParsePKIXPublicKey(block.Bytes)
    if err != nil {
            return nil, err
    }

    switch pub := pub.(type) {
    case *rsa.PublicKey:
            return pub, nil
    default:
            break // fall through
    }
    return nil, errors.New("Key type is not RSA")
}