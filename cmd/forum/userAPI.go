package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"gitlab.com/avokadoen/softsecoblig2/lib/database"
)

//Content creation handlers
func CreateNewTopic(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Create Topic Handle Recieved your request, though it's not yet implemented 😮")
	categoryName := mux.Vars(r)["category"]
	//TODO: validate existance of category...

	var category Category
	Server.Database.GetCategory(categoryName, &category)

	//TODO: DECODE JSON FROM POST
	//TODO: Create Topic WITH ObjectId, i.e. add ID manually after decode?
	//TODO: push ObjectId to TableCategory, push Topic to TableTopic... use db.Upsert?

}

func CreateNewComment(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Create Comment Handle Recieved your request")

	//Get user posting
	cookie, err := SecureCookie.FetchCookie(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "You appear to not be logged in 😮\nrelog to post your comment\n")
		//TODO, be nice with the user and store the comment while he's logging in?
		return
	}

	SecureCookie.AuthenticateCookie(w, Server, cookie)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "Your session is no longer valid 😮\nrelog to post your comment\n")
		//TODO, be nice with the user and store the comment while he's logging in?
		return
	}

	username := Server.Database.GetUsername(cookie.Id) //TODO handle errors?

	vars := mux.Vars(r)
	categoryName := vars["category"]
	topicID := vars["topicID"]

	err = Server.Database.ValidateSession()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//validate existance of topic in category with Id = topicID, simple pipe with lookup and match...
	var topicInCategory TopicAndCategory
	err = Server.Database.GetTopic(categoryName, topicID, &topicInCategory)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Something bad happened 😮\n")
		return
	}

	//TODO: Unmarshal json body into Comment structure, modify if neccessary...
	var comment database.Comment
	rBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Printf("unable to read, err: %v", err)
		return
	}
	err = json.Unmarshal(rBody, &comment)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Printf("unable to unmarshal, err: %v", err)
		return
	}
	comment.Username = username

	// Update topic with new comment
	Server.Database.PushTopicComment(topicID, comment)
}
