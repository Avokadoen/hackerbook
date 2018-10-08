package main

import (
	"net/http"
	"encoding/json"
	"io/ioutil"

	validator "github.com/asaskevich/govalidator"
	"github.com/globalsign/mgo/bson"
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

	adminID := Server.Database.AuthenticateAdmin(cookie.ID, sessPtr)

	if adminID != bson.ObjectId(0) {
		return adminID
	}
	
	return bson.ObjectId(0)

}

// AuthenticateAdminHandler is the handler called to verify if the user is logged in as an admin
func AuthenticateAdminHandler(w http.ResponseWriter, r *http.Request) {

	adminID := AuthenticateAdmin(w, r)

	if adminID != bson.ObjectId(0) {
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


	if !Server.Database.IsExistingCategory(category.Name, sessPtr) {

		if valid, err := validator.ValidateStruct(category); !valid {
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		Server.Database.InsertToCollection(database.TableCategory, category, sessPtr)
		w.WriteHeader(http.StatusCreated) //201
		w.Write([]byte("Category created"))
		return
	}
	w.WriteHeader(http.StatusBadRequest)

}
