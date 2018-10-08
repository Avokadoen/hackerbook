package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	validator "github.com/asaskevich/govalidator"
	"github.com/globalsign/mgo/bson"
	"gitlab.com/avokadoen/softsecoblig2/lib/database"
)

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

	if adminID != bson.ObjectId(0) {
		return adminID
	} else {
		fmt.Printf("User not admin, err: %v", err)
		return bson.ObjectId(0)
	}
}

func AuthenticateAdminHandler(w http.ResponseWriter, r *http.Request) {

	adminID := AuthenticateAdmin(w, r)

	if adminID != bson.ObjectId(0) {
		w.Write([]byte("Admin granted"))
		fmt.Printf("User is admin\n")
	} else {
		fmt.Printf("User not admin, err:")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Not granted"))
	}
}

func CreateNewCategoryHandler(w http.ResponseWriter, r *http.Request) {

	adminID := AuthenticateAdmin(w, r)

	if adminID != bson.ObjectId(0) {
		fmt.Printf("User is admin\n")
	} else {
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

	if !Server.Database.IsExistingCategory(category.Name, sessPtr) {

		if valid, err := validator.ValidateStruct(category); !valid {
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Printf("unable to validate topic: %+v\n", err)
				fmt.Printf("request body: %+v", string(rBody))
			}
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Category name contains non-alphanumeric characters!\n")
			return
		}

		Server.Database.InsertToCollection(database.TableCategory, category, sessPtr)
		fmt.Println("Category inserted to database!")
		w.WriteHeader(http.StatusCreated) //201
		w.Write([]byte("Category created"))
		return
	}
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprintf(w, "Category with name \"%s\" already exists!", category.Name)

}
