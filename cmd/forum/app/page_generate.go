package app

import (
	"html/template"
	"net/http"

	"github.com/globalsign/mgo/bson"
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
	Title     string `json:"name"`
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
	//TODO: get stuff from DB... for now I'll use mocked data
	tmpl := template.Must(template.ParseFiles("./web/index.html"))

	data := HomePage{
		Categories: []Category{
			{
				Name: "fishing",
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
				Name: "hentai",
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
		},
	}
	tmpl.Execute(w, data)
}

func GenerateCategoryPage(w http.ResponseWriter, r *http.Request) {

}
