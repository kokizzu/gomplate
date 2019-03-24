package aws

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/service/kms"
	b64 "github.com/hairyhenderson/gomplate/base64"
	"github.com/stretchr/testify/assert"
)

// MockKMS is a mock KMSAPI implementation
type MockKMS struct{}

// Mocks Encrypt operation returns an upper case version of plaintext
func (m *MockKMS) Encrypt(input *kms.EncryptInput) (*kms.EncryptOutput, error) {
	return &kms.EncryptOutput{
		CiphertextBlob: []byte(strings.ToUpper(string(input.Plaintext))),
	}, nil
}

// Mocks Decrypt operation
func (m *MockKMS) Decrypt(input *kms.DecryptInput) (*kms.DecryptOutput, error) {
	s := []byte(strings.ToLower(string(input.CiphertextBlob)))
	return &kms.DecryptOutput{
		Plaintext: s,
	}, nil
}

func TestEncrypt(t *testing.T) {
	// create a mock KMS client
	c := &MockKMS{}
	kmsClient := &KMS{Client: c}

	// Success
	resp, err := kmsClient.Encrypt("dummykey", "plaintextvalue")
	assert.NoError(t, err)
	expectedResp, _ := b64.Encode([]byte("PLAINTEXTVALUE"))
	assert.EqualValues(t, expectedResp, resp)
}

func TestDecrypt(t *testing.T) {
	// create a mock KMS client
	c := &MockKMS{}
	kmsClient := &KMS{Client: c}
	encodedCiphertextBlob, _ := b64.Encode([]byte("CIPHERVALUE"))

	// Success
	resp, err := kmsClient.Decrypt(encodedCiphertextBlob)
	assert.NoError(t, err)
	assert.EqualValues(t, "ciphervalue", resp)
}
