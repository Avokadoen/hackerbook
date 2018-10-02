package app

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/gorilla/securecookie"
	"gitlab.com/avokadoen/softsecoblig2/lib/database"
)

const CookieName = "HackerBook"
const CookieExpiration = time.Hour

//var secureCookieInstance = &securecookie.SecureCookie{}

type SCManager struct {
	secureCoIns *securecookie.SecureCookie
}

type CookieManager interface {
	Init()
	FetchCookie(r *http.Request) database.CookieData
	CreateCookie(w http.ResponseWriter, m bson.ObjectId, urlString string) string
	DeleteClientCookie(w http.ResponseWriter, urlString string) string
	DeleteDBCookie(clientCookie database.CookieData) error
	DecodeDBCookieData(data database.CookieData) database.CookieData
	AuthenticateCookie(w http.ResponseWriter, Server *Server, clientCookie database.CookieData) error
}

// TODO: we need to recreate securecookie if it is nil
func (SCManager *SCManager) Init() {
	SCManager.secureCoIns = securecookie.New(securecookie.GenerateRandomKey(32), securecookie.GenerateRandomKey(32))
}

func (SCManager *SCManager) FetchCookie(r *http.Request) (database.CookieData, error) {

	cookieData := database.CookieData{}

	cookie, err := r.Cookie(CookieName)
	if err != nil {
		return cookieData, err
	}
	err = SCManager.secureCoIns.Decode(CookieName, cookie.Value, &cookieData)
	if err != nil {
		return cookieData, err
	}

	return cookieData, nil
}

func (SCManager *SCManager) CreateCookie(w http.ResponseWriter, m bson.ObjectId, urlString string) string {
	timeCreated := time.Now().UnixNano()
	token := CreateHash(string(timeCreated))
	userID := m

	u, err := url.Parse(urlString)
	if err != nil {
		fmt.Printf("error at url parse error: %+v", err)
		return ""
	}
	cookieData := database.CookieData{
		Id:    userID,
		Token: token,
	}
	if encoded, err := SCManager.secureCoIns.Encode(CookieName, cookieData); err == nil {
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

func (SCManager *SCManager) DeleteClientCookie(w http.ResponseWriter, urlString string) error {
	u, err := url.Parse(urlString)
	if err != nil {
		return fmt.Errorf("error at url parse error: %+v", err)
	}
	cookieData := database.CookieData{
		Id:    bson.ObjectId(0),
		Token: "",
	}
	if encoded, err := SCManager.secureCoIns.Encode(CookieName, cookieData); err == nil {
		tokenCookie := http.Cookie{
			Name:     CookieName,
			Value:    encoded,
			HttpOnly: true,
			Domain:   u.Hostname(),
			Expires:  time.Now(),
		}
		fmt.Println("delete cookie")

		http.SetCookie(w, &tokenCookie)
		return nil
	}
	return fmt.Errorf("failed to delete client cookie")
}

func (SCManager *SCManager) DeleteDBCookie(clientCookie database.CookieData, Server *Server) error {
	if len(clientCookie.Token) <= 0 {
		return fmt.Errorf("invalid token in cookie")
	}
	encodedDbCookie := new(database.CookieData)

	Server.Database.GetCookie(clientCookie, encodedDbCookie)
	dbData := SCManager.DecodeDBCookieData(*encodedDbCookie)

	if dbData != clientCookie {
		return fmt.Errorf("clientCookie did not match db")
	}
	Server.Database.DeleteCookie(dbData.Id)
	return nil
}

func (SCManager *SCManager) DecodeDBCookieData(data database.CookieData) database.CookieData {

	decodeData := database.CookieData{}
	err := SCManager.secureCoIns.Decode(CookieName, data.Token, &decodeData)
	if err != nil {
		fmt.Printf("when decoding dbCookie error: %+v", err)
		return database.CookieData{}
	}
	return decodeData
}

func (SCManager *SCManager) AuthenticateCookie(w http.ResponseWriter, Server *Server, clientCookie database.CookieData) error {
	if len(clientCookie.Token) <= 0 {
		return fmt.Errorf("invalid token in cookie")
	}
	encodedDbCookie := new(database.CookieData)

	Server.Database.GetCookie(clientCookie, encodedDbCookie)
	dbData := SCManager.DecodeDBCookieData(*encodedDbCookie)

	if dbData != clientCookie {
		return fmt.Errorf("clientCookie did not match db")
	}
	return nil
}
