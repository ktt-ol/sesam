package conf

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/sirupsen/logrus"
	"io"
	"os"
)

var logger = logrus.WithField("where", "keys")

func GetKeys(keyStoreFile string) *Keys {

	if _, err := os.Stat(keyStoreFile); os.IsNotExist(err) {
		logger.WithField("keyStore", keyStoreFile).Info("Key store doesn't exist. Create new keys.")
		return createAndSaveNewKeys(keyStoreFile)
	}

	return readKeysFromFile(keyStoreFile)
}

func createAndSaveNewKeys(keyStoreFile string) *Keys {
	file, err := os.OpenFile(keyStoreFile, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	defer file.Close()
	if err != nil {
		logger.WithError(err).WithField("keyStoreFile", keyStoreFile).Fatal("Can't open file (for exclusive write).")
	}

	keys := Keys{
		SessionAuthKey:       GenerateRandomBytes(32),
		SessionEncryptionKey: GenerateRandomBytes(32),
		CsrfKey:              GenerateRandomString(32),
	}
	file.Write(keys.SessionAuthKey)
	file.Write(keys.SessionEncryptionKey)
	file.WriteString(keys.CsrfKey)

	return &keys
}

func readKeysFromFile(keyStoreFile string) *Keys {
	file, err := os.Open(keyStoreFile)
	defer file.Close()
	if err != nil {
		logger.WithError(err).WithField("keyStoreFile", keyStoreFile).Fatal("Can't open file (for read).")
	}

	keys := Keys{
		SessionAuthKey:       make([]byte, 32),
		SessionEncryptionKey: make([]byte, 32),
		CsrfKey:              "",
	}
	if _, err := file.Read(keys.SessionAuthKey); err != nil {
		logger.WithError(err).WithField("keyStoreFile", keyStoreFile).Fatal("Can't read 'SessionAuthKey'.")
	}
	if _, err := file.Read(keys.SessionEncryptionKey); err != nil {
		logger.WithError(err).WithField("keyStoreFile", keyStoreFile).Fatal("Can't read 'SessionEncryptionKey'.")
	}
	// the base64 encoding of 32 bytes must be less than 64 bytes
	buffer := make([]byte, 64)
	read, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		logger.WithError(err).WithField("keyStoreFile", keyStoreFile).Fatal("Can't read 'SessionEncryptionKey'.")
	}
	keys.CsrfKey = string(buffer[:read])

	return &keys
}

// https://elithrar.github.io/article/generating-secure-random-numbers-crypto-rand/

// GenerateRandomBytes returns securely generated random bytes.
// It will fail with a fatal log if the system's secure random
// number generator fails to function correctly
func GenerateRandomBytes(n int) []byte {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		// Note that err == nil only if we read len(b) bytes.
		logrus.Fatal("Could not read random bytes")
	}

	return b
}

// GenerateRandomString returns a URL-safe, base64 encoded
// securely generated random string.
// It will fail with a fatal log if the system's secure random
// number generator fails to function correctly
func GenerateRandomString(s int) string {
	b := GenerateRandomBytes(s)
	return base64.URLEncoding.EncodeToString(b)
}

type Keys struct {
	SessionAuthKey       []byte
	SessionEncryptionKey []byte
	CsrfKey              string
}
