package main

import (
	"fmt"
	"html/template"
	"net/http"
	"regexp"

	"github.com/globalsign/mgo/bson"
	"github.com/gorilla/mux"
	"github.com/microcosm-cc/bluemonday"
	blackfriday "gopkg.in/russross/blackfriday.v2"
)

type HomePage struct {
	Categories []Category
}

func GenerateHomePage(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
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

	var category CategoryWithTopics
	Server.Database.GetCategory(vars["category"], &category)

	modulo := func(a, b int) int {
		return a % b
	}

	tmpl := template.Must(template.New("category.html").Funcs(template.FuncMap{"mod": modulo}).ParseFiles("./web/category.html"))
	err := tmpl.ExecuteTemplate(w, "category.html", category)
	if err != nil {
		fmt.Println(err)
	}
}

func GenerateTopicPage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Generating Topic Page")
	vars := mux.Vars(r) //use vars to obtain data from db

	var topic TopicAndCategory

	if !bson.IsObjectIdHex(vars["topicID"]) {
		NotFoundHandler(w, r)
		return
	}

	Server.Database.GetTopic(vars["category"], vars["topicID"], &topic)

	if topic.Name == "" {
		NotFoundHandler(w, r)
		return
	}

	markDowner := func(args ...interface{}) template.HTML {
		unsafeMD := blackfriday.Run([]byte(fmt.Sprintf("%s", args...)))
		ugcPolicy := bluemonday.UGCPolicy()
		ugcPolicy.AllowAttrs("class").Matching(regexp.MustCompile("^language-[a-zA-Z0-9]+$")).OnElements("code")
		safeMD := ugcPolicy.SanitizeBytes(unsafeMD)
		return template.HTML(safeMD)
	}

	tmpl := template.Must(template.New("topic.html").Funcs(template.FuncMap{"markdown": markDowner}).ParseFiles("./web/topic.html"))
	err := tmpl.ExecuteTemplate(w, "topic.html", topic)
	if err != nil {
		fmt.Println(err)
	}

}

//MISC handlers
func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	tmpl := template.Must(template.ParseFiles("./web/no_content.html"))
	tmpl.Execute(w, nil) //TODO: generating actual static pages is kinda bad...
}
