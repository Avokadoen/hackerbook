package app

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"io"

	"gitlab.com/avokadoen/softsecoblig2/lib/database"
)

/* sources:
url parse: https://stackoverflow.com/a/49258338
securing cookies: https://www.calhoun.io/securing-cookies-in-go/
*/
type Server struct {
	Port        string
	Database    database.Db
	StaticPages map[int][]byte
}

func ConvertPlainPassword(rawUsername, rawPassword string) string {
	hashedName := CreateHash(rawUsername)
	return CreateHash(hashedName + rawPassword)
}

func CreateHash(key string) string {
	hasher := sha512.New() // TODO: maybe move so that new is only called once
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}

func Encrypt(data []byte, passphrase string) []byte {
	block, _ := aes.NewCipher([]byte(CreateHash(passphrase)))
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext
}

func Decrypt(data []byte, passphrase string) []byte {
	key := []byte(CreateHash(passphrase))
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err.Error())
	}
	return plaintext
}
