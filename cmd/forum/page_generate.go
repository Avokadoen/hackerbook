package main

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/globalsign/mgo/bson"
	"github.com/gorilla/mux"
)

type HomePage struct {
	Categories []Category
}

//Category - Shallow category, not containing other than id to reference topics
type Category struct {
	Id     bson.ObjectId   `bson:"_id,omitempty"`
	Name   string          `json:"name"`
	Topics []bson.ObjectId `json:"topics"`
	//MORE?
}

type CategoryWithTopics struct {
	Id     bson.ObjectId `bson:"_id,omitempty"`
	Name   string        `json:"name"`
	Topics []Topic       `json:"topics"`
	//MORE?
}

/*
	"_id" : ObjectId("5bb175765499851637a9379d"),
	"name" : "phishing",
	"topic" : {
		"_id" : ObjectId("5bb177bc5499851637a9379e"),
		"title" : "Test Post Pls Ignore",
		"content" : "test ok",
		"comments" : [ ],
		"createdBy" : ObjectId("5bb0ed24ed8bad61aa93bd85"),
		"creationTime" : ISODate("2018-10-01T01:26:20.214Z")
	}

*/
type TopicAndCategory struct {
	Id   bson.ObjectId `bson:"_id,omitempty"`
	Name string        `json:"name"`
	Topic
}

//Topic within a category
type Topic struct {
	Id        bson.ObjectId `bson:"_id,omitempty"`
	Title     string        `json:"title"`
	Content   string        `json:"content"`
	Comments  []Comment     `json:"comments"`
	CreatedBy bson.ObjectId `json:"createdBy"` //user
}

//Comment within a post
type Comment struct {
	Text      string        `json:"text"`
	CreatedBy bson.ObjectId `bson:"_id"`
	ReplyTo   *Comment      //if not a reply -> nil
}

func GenerateHomePage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Generating Home Page")
	var categories []Category
	Server.Database.GetCategories(&categories)
	//TODO: get stuff from DB... for now I'll use mocked data
	tmpl := template.Must(template.ParseFiles("./web/index.html"))
	data := HomePage{categories}

	fmt.Printf("%+v", categories)

	tmpl.Execute(w, data)
}

func GenerateCategoryPage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Generating Category Page")
	vars := mux.Vars(r) //use vars to obtain data from db

	var category CategoryWithTopics
	Server.Database.GetCategory(vars["category"], &category)
	tmpl := template.Must(template.ParseFiles("./web/category.html"))
	tmpl.Execute(w, category)
}

func GenerateTopicPage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Generating Topic Page")
	vars := mux.Vars(r) //use vars to obtain data from db

	var topic TopicAndCategory

	if !bson.IsObjectIdHex(vars["topicID"]) {
		NoContentHandler(w, r)
		return
	}

	Server.Database.GetTopic(vars["category"], vars["topicID"], &topic)
	fmt.Printf("Topic Generated page:\n%+v", topic)

	if topic.Name == "" {
		NoContentHandler(w, r)
		return
	}

	tmpl := template.Must(template.ParseFiles("./web/topic.html"))
	err := tmpl.Execute(w, topic)
	if err != nil {
		fmt.Println(err)
	}

}

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
	fmt.Fprintf(w, "Create Comment Handle Recieved your request, though it's not yet implemented 😮")

	// vars := mux.Vars(r)
	// categoryName := vars["category"]
	// topicID := vars["topicID"]

	//TODO: validate existance of topic in category with Id = topicID, simple pipe with lookup and match...
	//TODO: Unmarshal json body into Comment structure, modify if neccessary...

	//TODO: Update topic with new comment
}

//MISC handlers
func NoContentHandler(w http.ResponseWriter, r *http.Request) {

	tmpl := template.Must(template.ParseFiles("./web/no_content.html"))
	tmpl.Execute(w, nil) //TODO: generating actual static pages is kinda bad...
}

//TODO add handlers for other typical Status****, i.e. 404?
