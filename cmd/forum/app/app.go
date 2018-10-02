package app

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"fmt"

	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/gorilla/securecookie"
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

const CookieName = "HackerBook"
const CookieExpiration = time.Hour

var secureCookieInstance = &securecookie.SecureCookie{}

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

// TODO: we need to recreate securecookie if it is nil
func InitSecureCookie() {
	secureCookieInstance = securecookie.New(securecookie.GenerateRandomKey(32), securecookie.GenerateRandomKey(32))
}

func FetchCookie(r *http.Request) database.CookieData {

	cookieData := database.CookieData{}

	cookie, err := r.Cookie(CookieName)
	if err != nil {
		fmt.Printf("when requesting cookie error: %v+", err)
		return cookieData
	}
	err = secureCookieInstance.Decode(CookieName, cookie.Value, &cookieData)
	if err != nil {
		fmt.Printf("when decoding cookie error: %v+", err)
	}

	return cookieData
}

func CreateCookie(w http.ResponseWriter, m bson.ObjectId, urlString string) string {
	timeCreated := time.Now().UnixNano()
	token := CreateHash(string(timeCreated))
	userID := m

	u, err := url.Parse(urlString)
	if err != nil {
		fmt.Printf("error at url parse error: %v+", err)
		return ""
	}
	cookieData := database.CookieData{
		Id:    userID,
		Token: token,
	}
	if encoded, err := secureCookieInstance.Encode(CookieName, cookieData); err == nil {
		tokenCookie := http.Cookie{
			Name:     CookieName,
			Value:    encoded,
			HttpOnly: true,
			Domain:   u.Hostname(),
			Expires:  time.Now().Add(CookieExpiration),
		}
		fmt.Println("created cookie")

		http.SetCookie(w, &tokenCookie)
		return encoded
	}
	fmt.Println("failed to create cookie")
	return ""
}

func DecodeDBCookieData(data database.CookieData) database.CookieData {

	decodeData := database.CookieData{}
	err := secureCookieInstance.Decode(CookieName, data.Token, &decodeData)
	if err != nil {
		fmt.Printf("when decoding dbCookie error: %v+", err)
		return database.CookieData{}
	}
	return decodeData
}
