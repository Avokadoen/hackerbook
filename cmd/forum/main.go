package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"log"
	"strings"

	validator "github.com/asaskevich/govalidator"
	"github.com/gorilla/mux"
	"github.com/subosito/gotenv"
	"gitlab.com/avokadoen/softsecoblig2/cmd/forum/app"
	"gitlab.com/avokadoen/softsecoblig2/lib/database"
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
	fs := http.FileServer(http.Dir("./web"))
	fmt.Printf("%+v\n", fs)

	router.PathPrefix("/web/").Handler(http.StripPrefix("/web/", fs))

	//API ENDPOINTS
	router.HandleFunc("/postlogin", LoginAuthHandler).Methods(http.MethodPost).Headers("Content-Type", "application/json")
	router.HandleFunc("/signup", SignUpHandler).Methods(http.MethodPost).Headers("Content-Type", "application/json")
	router.HandleFunc("/test", IndexHandler)

	// router.HandleFunc("/", fs.ServeHTTP)
	// PAGE HANDLES
	router.HandleFunc("/", app.GenerateHomePage)
	router.HandleFunc("/r/{category}", app.GenerateCategoryPage)
	router.HandleFunc("/r/{category}/{topicID}", app.GenerateTopicPage)

	fmt.Printf("\nListening through port %v...\n", Server.Port)
	http.ListenAndServe(":"+Server.Port, router)
}

// TODO: Javascript deal with invalid messages
func SignUpHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fmt.Println("Start signup!")
	var rawUserData database.SignUpUser
	rBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
	}
	err = json.Unmarshal(rBody, &rawUserData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "unable to sign up")
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
	userStatus, err := Server.Database.IsExistingUser(user)
	if err != nil {
		log.Printf("failed to check user in sign up. error: %v+", err)
	} else if userStatus != nil {
		if strings.Contains(*userStatus, "username") {
			w.Write([]byte("username already exist"))
		}
		if strings.Contains(*userStatus, "email") {
			w.Write([]byte("\nemail already exist"))
		}
		return
	}

	Server.Database.InsertToCollection(database.TableUsers, user)

	fmt.Println("user inserted in database!")

}

func LoginAuthHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

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
	if Server.Database.AuthenticateUser(user) {
		body = []byte("login successful")
	}
	w.Write(body)
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
	Server.Database.InsertToCollection(database.TableCategory, category)
}
