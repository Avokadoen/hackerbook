package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"regexp"

	"github.com/globalsign/mgo/bson"
	"github.com/gorilla/mux"
	"github.com/microcosm-cc/bluemonday"
	"gopkg.in/russross/blackfriday.v2"
)

// HomePage struct containing the categories for the homepage
type HomePage struct {
	Categories []Category
}
// GenerateHomePage generates the homepage from categories in the database
func GenerateHomePage(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	sessPtr, err := Server.Database.CreateSessionPtr()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer sessPtr.Close()
	w.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
	fmt.Println("Generating Home Page")
	var categories []Category
	err = Server.Database.GetCategories(&categories, sessPtr)
	if err != nil {
		fmt.Println(err)
		//return
	}

	tmpl := template.Must(template.ParseFiles("./web/index.html"))
	data := HomePage{categories}

	tmpl.Execute(w, data)
}

// GenerateSignupPage generates the signup page with the environments captcha site key.
func GenerateSignupPage(w http.ResponseWriter, r *http.Request) {
	captcha := struct {
		Key string
	}{os.Getenv("CAPTCHA_SITE_KEY")}

	if captcha.Key == "" {
		fmt.Println("Missing captcha site-key, reCaptcha won't work now!")
	}

	tmpl := template.Must(template.ParseFiles("./web/signup.html"))
	err := tmpl.Execute(w, captcha)
	if err != nil {
		fmt.Println(err)
	}
}

// GenerateCategoryPage generates a categories page with its topics from the database
func GenerateCategoryPage(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	sessPtr, err := Server.Database.CreateSessionPtr()
	defer sessPtr.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Generating Category Page")
	vars := mux.Vars(r) //use vars to obtain data from db

	var category CategoryWithTopics

	Server.Database.GetCategory(vars["category"], &category, sessPtr)

	modulo := func(a, b int) int {
		return a % b
	}

	tmpl := template.Must(template.New("category.html").Funcs(template.FuncMap{"mod": modulo}).ParseFiles("./web/category.html"))
	err = tmpl.ExecuteTemplate(w, "category.html", category)
	if err != nil {
		fmt.Println(err)
	}

}

// GenerateTopicPage generates a topic site with comments from the database
func GenerateTopicPage(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	fmt.Println("Generating Topic Page")
	vars := mux.Vars(r) //use vars to obtain data from db

	var topic TopicAndCategory

	if !bson.IsObjectIdHex(vars["topicID"]) {
		NotFoundHandler(w, r)
		return
	}

	sessPtr, err := Server.Database.CreateSessionPtr()
	defer sessPtr.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	Server.Database.GetTopic(vars["category"], vars["topicID"], &topic, sessPtr)

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
	err = tmpl.ExecuteTemplate(w, "topic.html", topic)
	if err != nil {
		fmt.Println(err)
	}

}

// NotFoundHandler creates a default error site
func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.WriteHeader(http.StatusNotFound)
	tmpl := template.Must(template.ParseFiles("./web/no_content.html"))
	tmpl.Execute(w, nil) //TODO: generating actual static pages is kinda bad...
}
