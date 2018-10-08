package app

import (
	"fmt"
	"github.com/globalsign/mgo"
	"net/http"
	"net/url"
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/gorilla/securecookie"
	"gitlab.com/avokadoen/softsecoblig2/lib/database"
)

// SCManager implements CookieManager interface.
// Also holds a SecureCookie
type SCManager struct {
	secureCoIns *securecookie.SecureCookie
}

// CookieManager is for managing one secure cookie
type CookieManager interface {
	Init()
	FetchCookie(r *http.Request) database.CookieData
	CreateCookie(w http.ResponseWriter, m bson.ObjectId, urlString string) string
	DeleteClientCookie(w http.ResponseWriter, urlString string) string
	DeleteDBCookie(clientCookie database.CookieData, session *mgo.Session) error
	DecodeDBCookieData(data database.CookieData, session *mgo.Session) database.CookieData
	AuthenticateCookie(w http.ResponseWriter, Server *Server, clientCookie database.CookieData, session *mgo.Session) error
}


// Init initializes the secure cookie instance by generating random keys of 32 byte length
func (SCManager *SCManager) Init() {
	// TODO: we need to recreate securecookie if it is nil
	SCManager.secureCoIns = securecookie.New(securecookie.GenerateRandomKey(32), securecookie.GenerateRandomKey(32))
}

// FetchCookie retrieves the client cookie and decodes its content.
// Returns the decoded data and an error
func (SCManager *SCManager) FetchCookie(r *http.Request) (database.CookieData, error) {

	cookieData := database.CookieData{}

	cookie, err := r.Cookie(database.CookieName)
	if err != nil {
		return cookieData, err
	}
	err = SCManager.secureCoIns.Decode(database.CookieName, cookie.Value, &cookieData)
	if err != nil {
		return cookieData, err
	}

	return cookieData, nil
}

// CreateCookie creates a client cookie with login credentials and encodes it
// encoded data will be returned as string as well as set in http
func (SCManager *SCManager) CreateCookie(w http.ResponseWriter, m bson.ObjectId, urlString string) string {
	timeCreated := time.Now().UnixNano()
	token := CreateHash(string(timeCreated))
	userID := m

	u, err := url.Parse(urlString)
	if err != nil {
		return ""
	}
	cookieData := database.CookieData{
		Id:    userID,
		Token: token,
	}
	if encoded, err := SCManager.secureCoIns.Encode(database.CookieName, cookieData); err == nil {
		tokenCookie := http.Cookie{
			Name:     database.CookieName,
			Value:    encoded,
			HttpOnly: true,
			Domain:   u.Hostname(),
			Expires:  time.Now().Add(database.CookieExpiration),
			Secure:   false,
		}

		http.SetCookie(w, &tokenCookie)
		return encoded
	}
	return ""
}

// DeleteClientCookie deletes(overwrites) client cookie. Returns error if it failed
func (SCManager *SCManager) DeleteClientCookie(w http.ResponseWriter, urlString string) error {
	u, err := url.Parse(urlString)
	if err != nil {
		return fmt.Errorf("error at url parse error: %+v", err)
	}
	cookieData := database.CookieData{
		Id:    bson.ObjectId(0),
		Token: "",
	}
	if encoded, err := SCManager.secureCoIns.Encode(database.CookieName, cookieData); err == nil {
		tokenCookie := http.Cookie{
			Name:     database.CookieName,
			Value:    encoded,
			HttpOnly: true,
			Domain:   u.Hostname(),
			Expires:  time.Now(),
			Secure:   true,
		}

		http.SetCookie(w, &tokenCookie)
		return nil
	}
	return fmt.Errorf("failed to delete client cookie")
}

// DeleteDBCookie deletes all cookies with same user id as sent cookie
// Returns error if failed
func (SCManager *SCManager) DeleteDBCookie(clientCookie database.CookieData, Server *Server, session *mgo.Session) error {
	if len(clientCookie.Token) <= 0 {
		return fmt.Errorf("invalid token in cookie")
	}
	encodedDbCookie := new(database.CookieData)

	Server.Database.GetCookie(clientCookie, encodedDbCookie, session)
	dbData := SCManager.DecodeDBCookieData(*encodedDbCookie, session)

	if dbData != clientCookie {
		return fmt.Errorf("clientCookie did not match db")
	}
	Server.Database.DeleteCookie(dbData.Id, session)
	return nil
}

// DecodeDBCookieData decodes cookie data
// Returns decoded cookie data if success
func (SCManager *SCManager) DecodeDBCookieData(data database.CookieData, session *mgo.Session) database.CookieData {

	decodeData := database.CookieData{}
	err := SCManager.secureCoIns.Decode(database.CookieName, data.Token, &decodeData)
	if err != nil {
		return database.CookieData{}
	}
	return decodeData
}

// AuthenticateCookie checks client cookie towards db cookie to validate the client cookie
// Returns an error if authentication failed
func (SCManager *SCManager) AuthenticateCookie(w http.ResponseWriter, Server *Server, clientCookie database.CookieData, session *mgo.Session) error {
	if len(clientCookie.Token) <= 0 {
		return fmt.Errorf("invalid token in cookie")
	}
	encodedDbCookie := new(database.CookieData)

	Server.Database.GetCookie(clientCookie, encodedDbCookie, session)
	dbData := SCManager.DecodeDBCookieData(*encodedDbCookie, session)

	if dbData != clientCookie {
		return fmt.Errorf("clientCookie did not match db")
	}
	return nil
}
