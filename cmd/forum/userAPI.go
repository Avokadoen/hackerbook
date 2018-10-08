package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	validator "github.com/asaskevich/govalidator"
	"github.com/gorilla/mux"
	"gitlab.com/avokadoen/softsecoblig2/lib/database"
)

func authenticateUserHandler(w http.ResponseWriter, r *http.Request) string {
	//Get user posting
	cookie, err := SecureCookie.FetchCookie(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "You appear to not be logged in 😮\nrelog to post your comment\n")
		//TODO, be nice with the user and store the comment while he's logging in?
		return ""
	}

	sessPtr, err := Server.Database.CreateSessionPtr()
	defer sessPtr.Close()
	if err != nil {
		fmt.Println(err)
		return ""
	}

	SecureCookie.AuthenticateCookie(w, Server, cookie, sessPtr)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "Your session is no longer valid 😮\nrelog to post your comment\n")
		//TODO, be nice with the user and store the comment while he's logging in?
		return ""
	}
	return Server.Database.GetUsername(cookie.ID, sessPtr) //TODO handle errors?
}

//Content creation handlers

// CreateNewTopic creates a new topic, sets the user who posted it and sends its ID into the category in which it exists
func CreateNewTopic(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Create Topic Handle Recieved your request!")

	username := authenticateUserHandler(w, r)

	categoryName := mux.Vars(r)["category"]
	//TODO: validate existance of category...

	sessPtr, err := Server.Database.CreateSessionPtr()
	defer sessPtr.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	var category Category
	err = Server.Database.GetCategory(categoryName, &category, sessPtr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Something bad happened 😮\n")
		return
	}

	// DECODE JSON FROM POST
	var topic database.Topic
	rBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Printf("unable to read, err: %v", err)
		return
	}
	err = json.Unmarshal(rBody, &topic)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Printf("unable to unmarshal, err: %v", err)
		return
	}
	topic.Username = username

	topic.Category = category.Name

	if valid, err := validator.ValidateStruct(topic); !valid {
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Printf("unable to validate topic: %+v\n", err)
			fmt.Printf("request body: %+v", string(rBody))
		}
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Topic contains disallowed letters")
		return
	}

	//push ObjectId to TableCategory, push Topic to TableTopic... use db.Upsert?
	if err = Server.Database.CreateTopic(categoryName, topic, sessPtr); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Something bad happened 😮\n")
		fmt.Println(err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// CreateNewComment creates a new comment, sets the user to comment and validates the topic and category which it exists in
func CreateNewComment(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Create Comment Handle Recieved your request!")

	//Get user posting
	username := authenticateUserHandler(w, r)

	vars := mux.Vars(r)
	categoryName := vars["category"]
	topicID := vars["topicID"]

	sessPtr, err := Server.Database.CreateSessionPtr()
	defer sessPtr.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	//validate existance of topic in category with ID = topicID, simple pipe with lookup and match...
	var topicInCategory TopicAndCategory
	err = Server.Database.GetTopic(categoryName, topicID, &topicInCategory, sessPtr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Something bad happened 😮\n")
		return
	}

	//TODO: Unmarshal json body into Comment structure, modify if neccessary...
	var comment database.Comment
	rBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Printf("unable to read, err: %v\n", err)
		return
	}
	err = json.Unmarshal(rBody, &comment)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Printf("unable to unmarshal, err: %v\n", err)
		return
	}
	comment.Username = username

	if valid, err := validator.ValidateStruct(comment); !valid {
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Printf("unable to validate comment: %+v\n", err)
			fmt.Printf("request body: %+v", string(rBody))
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Comment contains disallowed letters")
		return
	}

	// Update topic with new comment
	if err = Server.Database.PushTopicComment(topicID, comment, sessPtr); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Something bad happened 😮\n")
		fmt.Println(err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
