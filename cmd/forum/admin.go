package main

import (
	"net/http"
	"github.com/globalsign/mgo/bson"
	"io/ioutil"
	"encoding/json"
	"gitlab.com/avokadoen/softsecoblig2/lib/database"
)

// AuthenticateAdmin returns the objectID of the admin account if it exists
func AuthenticateAdmin(w http.ResponseWriter, r *http.Request) bson.ObjectId {
	//Get user posting
	cookie, err := SecureCookie.FetchCookie(r)
	if err != nil {
		return bson.ObjectId(0)
	}

	sessPtr, err := Server.Database.CreateSessionPtr()
	defer sessPtr.Close()
	if err != nil {
		return bson.ObjectId(0)
	}

	SecureCookie.AuthenticateCookie(w, Server, cookie, sessPtr)
	if err != nil {
		return bson.ObjectId(0)
	}

	adminID := Server.Database.AuthenticateAdmin(cookie.Id, sessPtr)

	if adminID != bson.ObjectId(0){
		return adminID
	} else{
		return bson.ObjectId(0)
	}
}

// AuthenticateAdminHandler is the handler called to verify if the user is logged in as an admin
func AuthenticateAdminHandler(w http.ResponseWriter, r *http.Request) {

	adminID := AuthenticateAdmin(w, r)

	if adminID != bson.ObjectId(0){
		w.Write([]byte("Admin granted"))
	} else{
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Not granted"))
	}
}

// CreateNewCategoryHandler is the handler called for an admin to create new categories
func CreateNewCategoryHandler(w http.ResponseWriter, r *http.Request) {

	adminID := AuthenticateAdmin(w, r)

	if adminID != bson.ObjectId(0){
		// User is not admin
	} else{
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	sessPtr, err := Server.Database.CreateSessionPtr()
	defer sessPtr.Close()
	if err != nil {
		return
	}


	var category database.Category //TODO: Remove duplicate category struct in dbinterface and structs.go
	rBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = json.Unmarshal(rBody, &category)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var topics []bson.ObjectId
	category.Topics = topics

	if !Server.Database.IsExistingCategory(category.Name, sessPtr){
		Server.Database.InsertToCollection(database.TableCategory, category, sessPtr)
		w.Write([]byte("Category inserted"))
	}
}
