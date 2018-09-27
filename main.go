package main

import (
	"SoftwareSecurity2/db"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"net/http"
	"os"
)

type masseData struct {
	id bson.ObjectId `bson:"_id,omitempty"`
	name string `json:"name"`
}

var dbState db.DbState

func main() {
	dbState.Url = os.Getenv("DBURL")
	dbState.Name = os.Getenv("DBNAME")
	port := os.Getenv("PORT")

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)
	http.HandleFunc("/postlogin", LoginAuthHandler)
	http.HandleFunc("/test", IndexHandler)
	http.ListenAndServe(":"+port, nil)
}

func LoginAuthHandler(w http.ResponseWriter, r *http.Request){
	if r.Method == "POST" {
		defer r.Body.Close()
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		body, _ := ioutil.ReadAll(r.Body)

		w.Write(body)

	} else {
		http.Error(w, "invalid method used", http.StatusMethodNotAllowed)
	}
}

func IndexHandler(w http.ResponseWriter, r *http.Request){
	session, _ := dbState.CreateSession()
	c := dbState.GetCollection(session, "Category")
	//m := masseData{}

	count, _ := c.Count()
	fmt.Printf("%+v\n", count)
}




