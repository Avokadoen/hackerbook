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

//var (
//	LogT = SetLogger()
//)

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
	log.Print("1")
	validator.SetFieldsRequiredByDefault(true)

	Server = &app.Server{
		Port:     os.Getenv("PORT"),
		Database: &database.DbState{},
	}
	
	Server.Database.InitState() // TODO: use session copies instead of main pointer
	err := Server.Database.CreateMainSession()
	if err != nil {
		log.Fatal("failed to create db session\n" + err.Error())
	}

	LogT := SetLogger()	//TODO fix it, does not work outside SetLogger function?!
	LogT.Println("MainTest Println")
	LogT.Print("MainTest Print\r\n")
	LogT.Printf("MainTest printf %v","please")

	SecureCookie.Init()

	router := mux.NewRouter().StrictSlash(false)
	fs := http.FileServer(http.Dir("./web"))
	fmt.Printf("%+v\n", fs)

	router.PathPrefix("/web/").Handler(http.StripPrefix("/web/", fs))

	// POST HANDLES
	router.HandleFunc("/cookielogin", CookieLoginHandler).Methods(http.MethodPost)
	router.HandleFunc("/verifyadmin", AuthenticateAdminHandler).Methods(http.MethodPost)
	router.HandleFunc("/postlogin", ManualLoginHandler).Methods(http.MethodPost).Headers("Content-Type", "application/json")
	router.HandleFunc("/signup", SignUpHandler).Methods(http.MethodPost).Headers("Content-Type", "application/json")
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
	router.HandleFunc("/admincreatenewcategory", CreateNewCategoryHandler).Methods(http.MethodPost)
	router.NotFoundHandler = http.HandlerFunc(NotFoundHandler) //set 404 default handle

	log.Printf("\nListening through port %v...\n", Server.Port)
	// secure false: only when http, don't use in production
	//Csrf := csrf.Protect(securecookie.GenerateRandomKey(32),csrf.Secure(false))
	log.Fatal(http.ListenAndServe(":"+Server.Port, router)) //Csrf(
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

func SetLogger()(*log.Logger){ // Testing setting new loggers.
	errorLog, err := os.OpenFile("info.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND,0644)
	if err!= nil{
		fmt.Printf("error opening file: %v", err)
	}
	defer errorLog.Close()
	//logger.SetOutput(errorLog)
	//logger.Print("Loggertest\r\n")
	logger := log.New(errorLog,"logtest: ", 1)
	logger.Println("Setlogger\r\n")
	//mgo.SetLogger(logger) // Gjør mgo nå me dt?!?
	//mgo.SetDebug(true)
	return logger
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

	validateRecaptcha := ValidateReCaptcha(rawUserData.Response)

	if validateRecaptcha == false {
		fmt.Println("Captcha not validated successfully!")
		w.Write([]byte("Captcha not validated successfully!"))
		return
	} else {
	fmt.Println("Got through captcha validation!")
	}

	hashedPass := app.ConvertPlainPassword(rawUserData.Username, rawUserData.Password)

	fmt.Println("hashed password!")

	sessPtr, err := Server.Database.CreateSessionPtr()
	defer sessPtr.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	user := database.SignUpUser{
		Email:    rawUserData.Email,
		Username: rawUserData.Username,
		Password: hashedPass,
	}
	userStatus, err := Server.Database.IsExistingUser(user, sessPtr)
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

	Server.Database.InsertToCollection(database.TableUser, user, sessPtr)
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
	sessPtr, err := Server.Database.CreateSessionPtr()
	defer sessPtr.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	err = SecureCookie.AuthenticateCookie(w, Server, cookie, sessPtr)
	if err != nil {
		fmt.Printf("main failed to delete cookie, err: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	Server.Database.DeleteCookie(cookie.ID, sessPtr)

}

func CookieLoginHandler(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()
	cookie, err := SecureCookie.FetchCookie(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Printf("unable to fetch cookie, err: %v\n", err)
		return
	}
	sessPtr, err := Server.Database.CreateSessionPtr()
	defer sessPtr.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	err = SecureCookie.AuthenticateCookie(w, Server, cookie, sessPtr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Printf("\nunable to validate cookie, err: %v\n", err)
		return
	}
	sessPtr, err = Server.Database.CreateSessionPtr()
	defer sessPtr.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	username := Server.Database.GetUsername(cookie.ID, sessPtr)
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

	sessPtr, err := Server.Database.CreateSessionPtr()
	defer sessPtr.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	user := database.LoginUser{
		Username: rawUserData.Username,
		Password: hashedPass,
	}
	var body []byte
	body = []byte("login failed")
	userDBId := Server.Database.AuthenticateUser(user, sessPtr)
	if userDBId != bson.ObjectId(0) {
		body = []byte("login successful")
		encoded := SecureCookie.CreateCookie(w, userDBId, r.URL.Path)
		//w.Header().Set("X-CSRF-Token", csrf.Token(r))
		if encoded == "" {
			fmt.Println("failed to create cookie from main")
			return
		}
		dbCookie := database.CookieData{
			ID:    userDBId,
			Token: encoded,
		}
		Server.Database.DeleteCookie(dbCookie.ID, sessPtr)
		Server.Database.InsertToCollection(database.TableCookie, dbCookie, sessPtr)
	}
	w.Write(body)

}

func PostCommentHandler(w http.ResponseWriter, r *http.Request) {
	// TODO Get topic id
	sessPtr, err := Server.Database.CreateSessionPtr()
	defer sessPtr.Close()
	if err != nil {
		fmt.Println(err)
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
	username := Server.Database.GetUsername(cookie.ID, sessPtr)

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

	Server.Database.InsertToCollection(database.TableComment, comment, sessPtr)
	// TODO få lagt den inn i topic?
}


