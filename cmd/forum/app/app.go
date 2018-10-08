package app

import (
	"crypto/sha512"
	"encoding/hex"
	"gitlab.com/avokadoen/softsecoblig2/lib/database"
)

/* sources:
url parse: https://stackoverflow.com/a/49258338
securing cookies: https://www.calhoun.io/securing-cookies-in-go/
*/

// Server struct containing server information
type Server struct {
	Port        string
	Database    database.Db
	StaticPages map[int][]byte
}

// ConvertPlainPassword hashes a raw password and returns the hashed password
func ConvertPlainPassword(rawUsername, rawPassword string) string {
	hashedName := CreateHash(rawUsername)
	return CreateHash(hashedName + rawPassword)
}

// CreateHash creates a new hash string
func CreateHash(key string) string {
	hasher := sha512.New() // TODO: maybe move so that new is only called once
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}
