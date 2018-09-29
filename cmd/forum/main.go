package main

import (
	"fmt"
	validator "github.com/asaskevich/govalidator"
	"gitlab.com/avokadoen/softsecoblig2/lib/database"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/subosito/gotenv"
	"gitlab.com/avokadoen/softsecoblig2/cmd/forum/app"
)

// sources: https://www.thepolyglotdeveloper.com/2018/02/encrypt-decrypt-data-golang-application-crypto-packages/

func init() {
	gotenv.Load("./cmd/forum/.env") //this path is relative to working dir upon go install
}


var Server *app.Server

func main() {
	validator.SetFieldsRequiredByDefault(true)

	Server = &app.Server{
		Port: os.Getenv("PORT"),
		Database:&database.DbState{},
	}
	Server.Database.InitState() // TODO: move to handler or cookie

	router := mux.NewRouter().StrictSlash(false)
	fs := http.FileServer(http.Dir("./web/"))
	fmt.Printf("%+v\n", fs)
	router.PathPrefix("/web/").Handler(http.StripPrefix("/web/", fs))

	router.Handle("/", fs)

	router.HandleFunc("/postlogin", LoginAuthHandler)
	router.HandleFunc("/signup", SignUpHandler)

	router.HandleFunc("/test", IndexHandler)

	fmt.Printf("\nListening through port %v...\n", Server.Port)
	http.ListenAndServe(":"+Server.Port, router)
}

// TODO: Javascript deal with invalid messages
func SignUpHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		defer r.Body.Close()
		r.ParseForm()

		rawEmail 	 := r.FormValue("email")
		rawUsername := r.FormValue("username")
		rawPassword := r.FormValue("password")


		if !validator.IsExistingEmail(rawEmail) {
			fmt.Fprint(w, "invalid email") // TODO: replace
			return
		}
		if !validator.IsAlphanumeric(rawUsername) {
			fmt.Fprint(w, "invalid username") // TODO: replace
			return
		}
		if !validator.IsAlphanumeric(rawPassword){
			fmt.Fprint(w, "invalid password") // TODO: replace
			return
		}


		password := app.ConvertPlainPassword(rawUsername, rawPassword)

		err := Server.Database.ValidateSession()
		if err != nil {
			fmt.Println(err)
		}
		user := database.SignUpUser{
			Email:rawEmail,
			Username:rawUsername,
			Password:password,
		}
		
		Server.Database.InsertToCollection("users", user)

	} else {
		http.Error(w, "invalid method used", http.StatusMethodNotAllowed)
	}
}

func LoginAuthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		defer r.Body.Close()
		r.ParseForm()
		rawUsername := r.FormValue("username")
		rawPassword := r.FormValue("password")

		if !validator.IsAlphanumeric(rawUsername) {
			fmt.Fprint(w, "invalid username") // TODO: replace
			return
		}
		if !validator.IsAlphanumeric(rawPassword){
			fmt.Fprint(w, "invalid password") // TODO: replace
			return
		}

		password := app.ConvertPlainPassword(rawUsername, rawPassword)

		err := Server.Database.ValidateSession()
		if err != nil {
			fmt.Println(err)
		}
		user := database.LoginUser{
			Username:rawUsername,
			Password:password,
		}
		var body []byte
		body = []byte("login failed")
		if Server.Database.ValidateUser(user) {
			body = []byte("login successful")
		}
		w.Write(body)

	} else {
		http.Error(w, "invalid method used", http.StatusMethodNotAllowed)
	}
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {

	err := Server.Database.CreateSession()
	if err != nil {
		fmt.Println(err)
	}
	category := database.Category{
		Name:"hentai",
		Posts:99999,
	}
	Server.Database.InsertToCollection("CatEgory", category)
}
