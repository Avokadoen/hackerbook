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

type Category struct {
	Id     bson.ObjectId `json:"_id, omitempty"`
	Name   string        `json:"name"`
	Topics []Topic
	//MORE?
}

//Topic within a category
type Topic struct {
	Id        bson.ObjectId `json:"_id, omitempty"`
	Title     string        `json:"name"`
	Content   string
	Comments  []Comment
	CreatedBy bson.ObjectId `json:""` //user
}

//Comment within a post
type Comment struct {
	Text      string
	CreatedBy bson.ObjectId
	ReplyTo   *Comment //if not a reply -> nil
}

func GenerateHomePage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Generating Home Page")
	var categories []Category
	Server.Database.GetCategories(&categories)
	//TODO: get stuff from DB... for now I'll use mocked data
	tmpl := template.Must(template.ParseFiles("./web/index.html"))
	data := HomePage{categories}
	tmpl.Execute(w, data)
}

func GenerateCategoryPage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Generating Category Page")
	vars := mux.Vars(r) //use vars to obtain data from db

	var category Category
	Server.Database.GetCategory(vars["category"], &category)
	//TODO:Discuss wether to flip the relation around, topics hold collection name as reference

	tmpl := template.Must(template.ParseFiles("./web/category.html"))

	tmpl.Execute(w, category)
}

func GenerateTopicPage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Generating Topic Page")
	//TODO: get stuff from DB... for now I'll use mocked data
	vars := mux.Vars(r) //use vars to obtain data from db

	tmpl, err := template.ParseFiles("./web/topic.html")

	if err != nil {
		fmt.Println(err)
	}

	data := struct {
		Category string
		Topic    Topic
	}{
		vars["category"],
		Topic{
			Id:        bson.ObjectId("5bb0ed24ed8bad61aa93bd31"),
			Title:     "How to deal with Catfish",
			Content:   "How? please discuss bellow",
			CreatedBy: bson.ObjectId("5bb0ed24ed8bad61aa93bd85"),
			Comments: []Comment{
				{
					Text:      "I also want to know this, caught one earlier this week, only just realized it was a catfish :(",
					CreatedBy: bson.ObjectId("5bb0ed4fed8bad61aa93bf4e"),
				},
				{
					//scriptkiddie user
					Text:      "<script>alert(\"get hacked boi!\")</script>",
					CreatedBy: bson.ObjectId("5bb0ed24ed8bad61aa93bd85"),
				},
			},
		},
	}

	tmpl.Execute(w, data)
}
