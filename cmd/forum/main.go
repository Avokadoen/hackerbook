package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/globalsign/mgo/bson"

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
var SecureCookie app.SCManager// TODO: make sure there always is a securecookie
func main() {
	validator.SetFieldsRequiredByDefault(true)

	Server = &app.Server{
		Port:     os.Getenv("PORT"),
		Database: &database.DbState{},
	}
	Server.Database.InitState() // TODO: move to handler or cookie
	SecureCookie.Init()
	router := mux.NewRouter().StrictSlash(false)
	fs := http.FileServer(http.Dir("./web"))
	fmt.Printf("%+v\n", fs)

	router.PathPrefix("/web/").Handler(http.StripPrefix("/web/", fs))

	router.HandleFunc("/cookielogin", CookieLoginHandler).Methods(http.MethodPost)
	router.HandleFunc("/postlogin", ManualLoginHandler).Methods(http.MethodPost).Headers("Content-Type", "application/json")
	router.HandleFunc("/signup", SignUpHandler).Methods(http.MethodPost).Headers("Content-Type", "application/json")
	router.HandleFunc("/postcomment", PostCommentHandler).Methods(http.MethodPost).Headers("Content-Type", "application/json")
	router.HandleFunc("/signout", SignOutHandler).Methods(http.MethodPost).Headers("Content-Type", "application/json")
	//router.HandleFunc("/signup", SignUpHandler).Methods(http.MethodGet)

	router.HandleFunc("/test", IndexHandler)

	// router.HandleFunc("/", fs.ServeHTTP)
	// PAGE HANDLES
	router.HandleFunc("/", GenerateHomePage)
	router.HandleFunc("/r/{category}", GenerateCategoryPage)
	router.HandleFunc("/r/{category}/{topicID}", GenerateTopicPage)

	fmt.Printf("\nListening through port %v...\n", Server.Port)
	http.ListenAndServe(":"+Server.Port, router)
}

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


func SignOutHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	cookie := SecureCookie.FetchCookie(r)
	if len(cookie.Token) <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("failed login"))
		return
	}
	encodedDbCookie := new(database.CookieData)

	Server.Database.GetCookie(cookie, encodedDbCookie)
	dbData := SecureCookie.DecodeDBCookieData(*encodedDbCookie)

	SecureCookie.DeleteCookie(w, r.URL.Path)
	if dbData != cookie {
		w.WriteHeader(http.StatusBadRequest)
		//w.Write([]byte("failed to validate cookie"))
		return
	}
	Server.Database.DeleteCookie(dbData.Id)


}

func CookieLoginHandler(w http.ResponseWriter, r *http.Request)(){

	defer r.Body.Close()
	cookie := SecureCookie.FetchCookie(r)
	if len(cookie.Token) <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("failed login"))
		return
	}
	encodedDbCookie := new(database.CookieData)

	Server.Database.GetCookie(cookie, encodedDbCookie)
	dbData := SecureCookie.DecodeDBCookieData(*encodedDbCookie)

	if dbData != cookie {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("failed to validate cookie"))
		return
	}
	username := Server.Database.GetUsername(cookie.Id)
	w.Write([]byte(username))
}

func ManualLoginHandler(w http.ResponseWriter, r *http.Request) {
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
	userDBId := Server.Database.AuthenticateUser(user)
	if userDBId != bson.ObjectId(0) {
		body = []byte("login successful")
		encoded := SecureCookie.CreateCookie(w, userDBId, r.URL.Path)
		if encoded == "" {
			fmt.Println("failed to create cookie from main")
			return
		}
		dbCookie := database.CookieData{
			Id:    userDBId,
			Token: encoded,
		}
		Server.Database.DeleteCookie(dbCookie.Id)
		Server.Database.InsertToCollection(database.TableCookie, dbCookie)
	}
	w.Write(body)

}

func IndexHandler(w http.ResponseWriter, r *http.Request) {

	err := Server.Database.CreateSession()
	if err != nil {
		fmt.Println(err)
	}
	category := Category{
		Name: "hentai",
	}
	Server.Database.InsertToCollection(database.TableCategory, category)
}

func PostCommentHandler (w http.ResponseWriter, r *http.Request){
	// TODO Get topic id
	err := Server.Database.ValidateSession()
	if err != nil{
		//error piss
	}
	// TODO Aksel fiks get user
	var commentRaw database.Comment
	rBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintln(w, "unable to read")
	}
	err = json.Unmarshal(rBody, &commentRaw)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "unable to unmarshal")
	}
	comment := database.Comment{
		CommentID: "12309712", 		// TODO generate id
		Username: "kek", 			// TODO Aksel fiks get user
		Text: "Hello world",		// TODO hent den her fra r på en måte
	}

	Server.Database.InsertToCollection(database.TableComment, comment)
	// TODO få lagt den inn i topic?
}