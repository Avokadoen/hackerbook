package main

import (
	"net/http"
	"fmt"
	"github.com/globalsign/mgo/bson"
)

func AuthenticateAdmin(w http.ResponseWriter, r *http.Request) bson.ObjectId {
	//Get user posting
	cookie, err := SecureCookie.FetchCookie(r)
	if err != nil {
		fmt.Printf("Failed to fetch cookie, err: %v", err)
		return bson.ObjectId(0)
	}

	SecureCookie.AuthenticateCookie(w, Server, cookie)
	if err != nil {
		fmt.Printf("Failed to authenticate cookie, err: %v", err)
		return bson.ObjectId(0)
	}

	adminID := Server.Database.AuthenticateAdmin(cookie.Id)

	if adminID != bson.ObjectId(0){
		return adminID
	} else{
		fmt.Printf("User not admin, err: %v", err)
		return bson.ObjectId(0)
	}
}

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
