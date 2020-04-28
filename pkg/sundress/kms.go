package sundress

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"log"
)

type secret struct {
	keyId  *string
	client *kms.KMS
}

var (
	sec *secret
)

func init() {
	sec = newSecret()
}

func newSecret() *secret {
	//need to change how credentials are stored, rn using my local
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	)

	if err != nil {
		log.Fatal(err)
	}

	svc := kms.New(sess)

	//need to replace the key with bdc stuff
	return &secret{
		keyId:  aws.String("arn:aws:kms:us-east-1:460622542287:key/8b4b0689-b1f2-4fb0-90cd-11f8a0c080ca"),
		client: svc,
	}
}

func Encrypt(s string) string {
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
	result, err := sec.client.Decrypt(&kms.DecryptInput{
		CiphertextBlob: []byte(s),
	})

	if err != nil {
		log.Fatal(err)
	}
	return string(result.Plaintext)
}
