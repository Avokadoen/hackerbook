package app

import (
	"fmt"
	"github.com/globalsign/mgo/bson"
	"github.com/gorilla/securecookie"
	"gitlab.com/avokadoen/softsecoblig2/lib/database"
	"net/http"
	"net/url"
	"time"
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
	CreateCookie(w http.ResponseWriter, m bson.ObjectId, urlString string) (string)
	DeleteCookie(w http.ResponseWriter, urlString string) (string)
	DecodeDBCookieData(data database.CookieData) database.CookieData

	//AuthenticateCookie()
}

// TODO: we need to recreate securecookie if it is nil
func (SCManager *SCManager) Init(){
	SCManager.secureCoIns = securecookie.New(securecookie.GenerateRandomKey(32), securecookie.GenerateRandomKey(32))
}

func (SCManager *SCManager) FetchCookie(r *http.Request) database.CookieData{

	cookieData := database.CookieData{}

	cookie, err := r.Cookie(CookieName)
	if err != nil {
		fmt.Printf("when requesting cookie error: %v+", err)
		return cookieData
	}
	err = SCManager.secureCoIns.Decode(CookieName, cookie.Value, &cookieData)
	if err != nil {
		fmt.Printf("when decoding cookie error: %v+", err)
	}

	return cookieData
}

func (SCManager *SCManager) CreateCookie(w http.ResponseWriter, m bson.ObjectId, urlString string) (string) {
	timeCreated := time.Now().UnixNano()
	token := CreateHash(string(timeCreated))
	userID := m

	u, err := url.Parse(urlString)
	if err != nil {
		fmt.Printf("error at url parse error: %v+", err)
		return ""
	}
	cookieData := database.CookieData {
		Id: userID,
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

func (SCManager *SCManager) DeleteCookie(w http.ResponseWriter, urlString string) (string) {
	u, err := url.Parse(urlString)
	if err != nil {
		fmt.Printf("error at url parse error: %v+", err)
		return ""
	}
	cookieData := database.CookieData {
		Id:bson.ObjectId(0),
		Token:"",
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
		return encoded
	}
	fmt.Println("failed to delete cookie")
	return ""
}

func (SCManager *SCManager) DecodeDBCookieData(data database.CookieData) database.CookieData{

	decodeData := database.CookieData{}
	err := SCManager.secureCoIns.Decode(CookieName, data.Token, &decodeData)
	if err != nil {
		fmt.Printf("when decoding dbCookie error: %v+", err)
		return database.CookieData{}
	}
	return decodeData
}