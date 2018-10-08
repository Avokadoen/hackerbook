package main

import (
	"net/http"
	"fmt"
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
		fmt.Printf("Failed to fetch cookie, err: %v", err)
		return bson.ObjectId(0)
	}

	sessPtr, err := Server.Database.CreateSessionPtr()
	defer sessPtr.Close()
	if err != nil {
		fmt.Println(err)
		return bson.ObjectId(0)
	}

	SecureCookie.AuthenticateCookie(w, Server, cookie, sessPtr)
	if err != nil {
		fmt.Printf("Failed to authenticate cookie, err: %v", err)
		return bson.ObjectId(0)
	}

	adminID := Server.Database.AuthenticateAdmin(cookie.Id, sessPtr)

	if adminID != bson.ObjectId(0){
		return adminID
	} else{
		fmt.Printf("User not admin, err: %v", err)
		return bson.ObjectId(0)
	}
}

// AuthenticateAdminHandler is the handler called to verify if the user is logged in as an admin
func AuthenticateAdminHandler(w http.ResponseWriter, r *http.Request) {

	adminID := AuthenticateAdmin(w, r)

	if adminID != bson.ObjectId(0){
		w.Write([]byte("Admin granted"))
		fmt.Printf("User is admin\n")
	} else{
		fmt.Printf("User not admin, err:")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Not granted"))
	}
}

// CreateNewCategoryHandler is the handler called for an admin to create new categories
func CreateNewCategoryHandler(w http.ResponseWriter, r *http.Request) {

	adminID := AuthenticateAdmin(w, r)

	if adminID != bson.ObjectId(0){
		fmt.Printf("User is admin\n")
	} else{
		fmt.Printf("User not admin, err:")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	sessPtr, err := Server.Database.CreateSessionPtr()
	defer sessPtr.Close()
	if err != nil {
		fmt.Println(err)
		return
	}


	var category database.Category //TODO: Got duplicate category struct in dbinterface and structs.go
	rBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Printf("unable to read, err: %v", err)
		return
	}
	err = json.Unmarshal(rBody, &category)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Printf("unable to unmarshal, err: %v", err)
		return
	}

	var topics []bson.ObjectId
	category.Topics = topics

	if !Server.Database.IsExistingCategory(category.Name, sessPtr){ //TODO: This somehow says it already exists
		Server.Database.InsertToCollection(database.TableCategory, category, sessPtr) //TODO: This doesn't put it properly into the db
		fmt.Println("Category inserted to database!")
		w.Write([]byte("Category inserted"))
	}
}
