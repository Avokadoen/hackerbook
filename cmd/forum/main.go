package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/dchest/captcha"
	"github.com/globalsign/mgo/bson"

	"log"
	"strings"

	validator "github.com/asaskevich/govalidator"
	"github.com/gorilla/mux"
	"github.com/subosito/gotenv"
	"gitlab.com/avokadoen/softsecoblig2/cmd/forum/app"
	"gitlab.com/avokadoen/softsecoblig2/lib/database"
)

/* sources:
https://www.thepolyglotdeveloper.com/2018/02/encrypt-decrypt-data-golang-application-crypto-packages/
https://www.kaihag.com/https-and-go/
*/
func init() {
	gotenv.Load("./cmd/forum/.env") //this path is relative to working dir upon go install
}

var Server *app.Server
var SecureCookie app.SCManager // TODO: make sure there always is a securecookie
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
	router.HandleFunc("/createcaptcha", CreateCaptchaHandler)
	router.HandleFunc("/postcomment", PostCommentHandler).Methods(http.MethodPost).Headers("Content-Type", "application/json")
	router.HandleFunc("/signout", SignOutHandler).Methods(http.MethodPost).Headers("Content-Type", "application/json")

	//router.HandleFunc("/signup", SignUpHandler).Methods(http.MethodGet)

	// router.HandleFunc("/", fs.ServeHTTP)
	// PAGE HANDLES
	router.HandleFunc("/", GenerateHomePage)
	router.HandleFunc("/r/{category}", GenerateCategoryPage)
	router.HandleFunc("/r/{category}/newtopic", CreateNewTopic).Methods(http.MethodPost)

	router.HandleFunc("/r/{category}/{topicID}", GenerateTopicPage).Methods(http.MethodGet)

	router.HandleFunc("/r/{category}/{topicID}/comment", CreateNewComment).Methods(http.MethodPost)

	router.NotFoundHandler = http.HandlerFunc(NotFoundHandler) //set 404 default handle

	fmt.Printf("\nListening through port %v...\n", Server.Port)
	http.ListenAndServe(":"+Server.Port, router)
	//go http.ListenAndServeTLS(":"+Server.Port, "cert.pem", "key.pem", router)
	/*
		cfg := &tls.Config{
			MinVersion:               tls.VersionTLS12,
			CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			},
		}
		srv := &http.Server{
			Addr:         ":"+Server.Port,
			Handler:      router,
			TLSConfig:    cfg,
			TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
		}
		log.Fatal(srv.ListenAndServeTLS("server.crt", "server.key"))*/
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
		return
	}
	err = json.Unmarshal(rBody, &rawUserData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "unable to sign up")
		fmt.Println(string(rBody))
		return
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
		log.Printf("failed to check user in sign up. error: %+v", err)
		return
	} else if userStatus != nil {
		if strings.Contains(*userStatus, "username") {
			w.Write([]byte("username already exist"))
		}
		if strings.Contains(*userStatus, "email") {
			w.Write([]byte("\nemail already exist"))
		}
		return
	}

	Server.Database.InsertToCollection(database.TableUser, user)
	fmt.Println("user inserted in database!")
	EmailVerification(w, r, user)
}

func SignOutHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	cookie, err := SecureCookie.FetchCookie(r)
	if err != nil {
		fmt.Printf("main fetch cookie, err: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = SecureCookie.DeleteClientCookie(w, r.URL.Path)
	if err != nil {
		fmt.Printf("main failed to delete client cookie, err: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = SecureCookie.AuthenticateCookie(w, Server, cookie)
	if err != nil {
		fmt.Printf("main failed to delete cookie, err: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	Server.Database.DeleteCookie(cookie.Id)

}

func CookieLoginHandler(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()
	cookie, err := SecureCookie.FetchCookie(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Printf("unable to fetch cookie, err: %v", err)
		return
	}
	err = SecureCookie.AuthenticateCookie(w, Server, cookie)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Printf("unable to validate cookie, err: %v", err)
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
		return
	}
	err = json.Unmarshal(rBody, &rawUserData)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Printf("failed to unmarshal: %v", err)
		return
	}

	if valid, err := validator.ValidateStruct(rawUserData); !valid {
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
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

func PostCommentHandler(w http.ResponseWriter, r *http.Request) {
	// TODO Get topic id
	err := Server.Database.ValidateSession()
	if err != nil {
		fmt.Printf("unable to validate session, err: %v", err)
		return
	}
	// TODO: should use unique postComment struct
	var commentRaw database.Comment
	rBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		fmt.Printf("unable to read, err: %v", err)
		return
	}
	err = json.Unmarshal(rBody, &commentRaw)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Printf("unable to unmarshal, err: %v", err)
		return
	}
	cookie, err := SecureCookie.FetchCookie(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Printf("unable to fetch cookie: %v", err)
		return
	}
	username := Server.Database.GetUsername(cookie.Id)

	comment := database.Comment{
		Username: username,
		Text:     commentRaw.Text, // TODO hent den her fra r på en måte
	}

	if valid, err := validator.ValidateStruct(comment); !valid {
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Printf("unable to validate comment: %v", err)
		}
		fmt.Fprint(w, "invalid comment")
		return

	}

	Server.Database.InsertToCollection(database.TableComment, comment)
	// TODO få lagt den inn i topic?
}

func CreateCaptchaHandler(w http.ResponseWriter, r *http.Request) {

	sessionId := captcha.New()
	err := captcha.WriteImage(w, sessionId, 240, 80)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	/*if r.Method == http.MethodGet {
		sessionID := app.CreateHash(string(time.Now().UnixNano()))
		w.

	} else if r.Method == http.MethodPost {
		var signUpSession database.SignupSession
		rBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			fmt.Printf("unable to read, err: %v", err)
			return
		}
		err = json.Unmarshal(rBody, &signUpSession)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Printf("unable to unmarshal, err: %v", err)
			return
		}
	}
	*/
}
