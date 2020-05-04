package sundress

import (
	"encoding/hex"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"log"
	"os"
)

type secret struct {
	keyId  *string
	client *kms.KMS
}

var (
	sec     *secret
	dataKey string
)

func InitSundress() {
	initSecret()
	initDataKey()
}

func initDataKey() {
	log.Println("Initializing Bdc datakey")
	cipherStr := os.Getenv("BDC_CIPHER_KEY")
	cipher, err := hex.DecodeString(cipherStr)
	if err != nil {
		log.Fatalf("Unable to decode cipher string: %v", err)
	}
	dataKey = Decrypt(cipher)
}

func initSecret() {
	log.Println("Initializing Secret...")
	//for some reason wasnt pulling region from ~/.aws/config
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-east-1"),
		Credentials: credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), ""),
	})
	if err != nil {
		log.Fatal(err)
	}
	svc := kms.New(sess)
	//need to replace the key with bdc stuff
	sec = &secret{
		keyId:  aws.String(os.Getenv("AWS_KMS_KEYID")),
		client: svc,
	}
}

func Encrypt(s string) []byte {
	log.Print("Calling kms to encrpyt string!")
	// Encrypt the data
	result, err := sec.client.Encrypt(&kms.EncryptInput{
		KeyId:     sec.keyId,
		Plaintext: []byte(s),
	})

	if err != nil {
		log.Fatal(err)
	}
	return result.CiphertextBlob
}

func Decrypt(s []byte) string {
	log.Print("Calling kms to decrypt string!")
	result, err := sec.client.Decrypt(&kms.DecryptInput{
		CiphertextBlob: s,
	})

	if err != nil {
		log.Fatal(err)
	}
	return string(result.Plaintext)
}
