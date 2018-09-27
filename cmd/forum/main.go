package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	// "gopkg.in/mgo.v2/bson"
	"github.com/globalsign/mgo/bson"
	"github.com/gorilla/mux"
	"github.com/subosito/gotenv"
	"gitlab.com/avokadoen/softsecoblig2/cmd/forum/app"
)

func init() {
	gotenv.Load("./cmd/forum/.env") //this path is relative to working dir upon go install
}

type Category struct {
	id    bson.ObjectId `bson:"_id,omitempty"`
	Name  string        `json:"name"`
	Posts int           `json:"posts"`
}

var Server *app.Server

func main() {

	Server = &app.Server{
		Port: os.Getenv("PORT"),
	}
	Server.Database.InitState() // TODO: move to handler or cookie

	router := mux.NewRouter().StrictSlash(false)

	fs := http.FileServer(http.Dir("web"))
	router.Handle("/", fs)
	router.HandleFunc("/postlogin", LoginAuthHandler)
	router.HandleFunc("/test", IndexHandler)

	fmt.Printf("\nListening through port %v...\n", Server.Port)
	http.ListenAndServe(":"+Server.Port, router)
}

func LoginAuthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		defer r.Body.Close()
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		body, _ := ioutil.ReadAll(r.Body)

		w.Write(body)

	} else {
		http.Error(w, "invalid method used", http.StatusMethodNotAllowed)
	}
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(1)
	err := Server.Database.CreateSession()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(2)
	// fmt.Print(Server.Database.Session)
	c := Server.Database.GetCollection("category")
	fmt.Println(3)
	var m Category

	err = c.Find(bson.M{"name": "fishing"}).One(&m)
	fmt.Println(4)
	if err != nil {
		fmt.Errorf("Failed Counting! %+v", err)
	}
	fmt.Fprintf(w, "found %+v", m)
}
