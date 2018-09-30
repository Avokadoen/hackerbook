package app

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

	//TODO: get stuff from DB... for now I'll use mocked data
	tmpl := template.Must(template.ParseFiles("./web/index.html"))

	data := HomePage{
		Categories: []Category{
			{
				Name: "phishing",
				Topics: []Topic{
					{
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
				},
			},
			{
				Name: "programming",
				Topics: []Topic{
					{
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
				},
			},
			{
				Name:   "programming",
				Topics: []Topic{{}, {}, {}},
			},
			{
				Name:   "cooking",
				Topics: []Topic{{}, {}, {}, {}, {}, {}, {}, {}, {}},
			},
			{
				Name:   "movies",
				Topics: []Topic{{}, {}, {}, {}, {}},
			},
			{
				Name:   "lifehacks",
				Topics: []Topic{{}, {}, {}, {}, {}, {}, {}},
			},
			{
				Name:   "MORE",
				Topics: []Topic{{}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}},
			},
		},
	}
	tmpl.Execute(w, data)
}

func GenerateCategoryPage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Generating Category Page")

	//TODO: get stuff from DB... for now I'll use mocked data
	vars := mux.Vars(r) //use vars to obtain data from db

	tmpl := template.Must(template.ParseFiles("./web/category.html"))

	data := Category{
		Name: vars["category"],
		Topics: []Topic{
			{
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
		},
	}
	tmpl.Execute(w, data)
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
