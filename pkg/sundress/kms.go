package sundress

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/joho/godotenv"
	"log"
	"os"
)

type secret struct {
	keyId  *string
	client *kms.KMS
}

var (
	sec *secret
)

func init() {
	envErr := godotenv.Load()
	if envErr != nil {
		log.Fatal("Error loading .env file")
	}
	sec = getSecret()
}

func getSecret() *secret {
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
	return &secret{
		keyId:  aws.String(os.Getenv("AWS_KMS_KEYID")),
		client: svc,
	}
}

func Encrypt(s string) string {
	// if true {
	// 	return s + "_encrypted"
	// }
	// Encrypt the data
	result, err := sec.client.Encrypt(&kms.EncryptInput{
		KeyId:     sec.keyId,
		Plaintext: []byte(s),
	})

	if err != nil {
		log.Fatal(err)
	}
	return string(result.CiphertextBlob)
}

func Decrypt(s string) string {
	// if true {
	// 	return s + "_decrypted"
	// }
	result, err := sec.client.Decrypt(&kms.DecryptInput{
		CiphertextBlob: []byte(s),
	})

	if err != nil {
		log.Fatal(err)
	}
	return string(result.Plaintext)
}
