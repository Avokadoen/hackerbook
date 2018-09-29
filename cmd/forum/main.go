package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	validator "github.com/asaskevich/govalidator"
	"gitlab.com/avokadoen/softsecoblig2/lib/database"

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
		Port:     os.Getenv("PORT"),
		Database: &database.DbState{},
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
		fmt.Println("Start signup!")
		// if r.Header.Get("Content-Type") //TODO: handle different Content-Types? - or validate Content-Type
		var rawUserData database.SignUpUser
		rBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
		}
		err = json.Unmarshal(rBody, &rawUserData)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "unable to login")
			fmt.Println(string(rBody))
		}

		if valid, err := validator.ValidateStruct(rawUserData); !valid {
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintln(w, "unable to validate user")
				fmt.Printf("unable to validate user: %v", err)
			}
			fmt.Fprint(w, "invalid user")
			return
		}
		fmt.Println("user signup validated!")
		hashedPass := app.ConvertPlainPassword(rawUserData.Username, rawUserData.Password)

		fmt.Println("hashed password!")

		err = Server.Database.ValidateSession()

		fmt.Println("got through session validation!")
		if err != nil {
			fmt.Println(err)
		}
		user := database.SignUpUser{
			Email:    rawUserData.Email,
			Username: rawUserData.Username,
			Password: hashedPass,
		}

		fmt.Println("created user for database insertion!")

		Server.Database.InsertToCollection(app.TableUsers, user)

		fmt.Println("user inserted in database!")
	} else {
		http.Error(w, "invalid method used", http.StatusMethodNotAllowed)
	}
}

func LoginAuthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		defer r.Body.Close()
		// if r.Header.Get("Content-Type") //TODO: handle different Content-Types? - or validate Content-Type

		var rawUserData database.LoginUser
		rBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
		}
		err = json.Unmarshal(rBody, &rawUserData)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "unable to login")
		}

		if valid, err := validator.ValidateStruct(rawUserData); !valid {
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintln(w, "unable to validate user")
				fmt.Printf("unable to validate user: %v", err)
			}
			fmt.Fprint(w, "invalid user")
			return
		}

		hashedPass := app.ConvertPlainPassword(rawUserData.Username, rawUserData.Password)

		err = Server.Database.ValidateSession()
		if err != nil {
			fmt.Println(err)
		}
		user := database.LoginUser{
			Username: rawUserData.Username,
			Password: hashedPass,
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
		Name:  "hentai",
		Posts: 99999,
	}
	Server.Database.InsertToCollection(app.TableCategory, category)
}
