package database

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/globalsign/mgo/bson"

	"github.com/globalsign/mgo"
	// mgo "gopkg.in/mgo.v2"
)

const (
	//DATABASE TABLES
	TableCategory = "category"
	TableUsers    = "users"
)

type Db interface {
	InitState()
	CreateSession() error
	ValidateSession() error
	InsertToCollection(collectionName string, data interface{}) error
	AuthenticateUser(user LoginUser) bool
	IsExistingUser(user SignUpUser) (*string, error)
}

type DbState struct {
	Hosts    string
	DbName   string
	Username string
	Password string
	Session  *mgo.Session
}

type Category struct {
	Id    bson.ObjectId `bson:"_id,omitempty" valid:"-"`
	Name  string        `json:"name" valid:"-"`
	Posts int           `json:"posts" valid:"-"`
}

type SignUpUser struct {
	Id       bson.ObjectId `bson:"_id,omitempty" valid:"-, optional"`
	Email    string        `json:"email" valid:"email, required"`
	Username string        `json:"username" valid:"alphanum, required"`
	Password string        `json:"password" valid:"alphanum, required"`
}

type LoginUser struct {
	Username string `json:"username" valid:"alphanum, required"`
	Password string `json:"password" valid:"alphanum, required"`
}

func (db *DbState) InitState() {
	db.Hosts = os.Getenv("DBURL")
	db.DbName = os.Getenv("DBNAME")
	db.Username = os.Getenv("DBUSERNAME")
	db.Password = os.Getenv("DBPASSWORD")

	fmt.Printf("%+v\n", db.Hosts)
	fmt.Printf("%+v\n", db.DbName)
	fmt.Printf("%+v\n", db.Username)
	fmt.Printf("%+v\n", db.Password)
}

func (db *DbState) CreateSession() (err error) {

	url := fmt.Sprintf("mongodb://%s:%s@%s/%s", db.Username, db.Password, db.Hosts, db.DbName)

	fmt.Println("Dialing the database!")

	fmt.Println(url)
	db.Session, err = mgo.Dial(url)

	if db.Session == nil {
		log.Fatal("Session was nil")
	}

	if err != nil {
		err = fmt.Errorf("died on error: %+v", err)
	}

	return err
}

func (db *DbState) ValidateSession() error {
	if db.Session != nil {
		return nil
	}
	err := db.CreateSession()
	if err != nil {
		return err
	}
	return nil
}

func (db *DbState) GetCollection(collectionName string) *mgo.Collection {
	return db.Session.DB(db.DbName).C(strings.ToLower(collectionName))
}

func (db *DbState) InsertToCollection(collectionName string, data interface{}) error {
	collection := db.GetCollection(collectionName)
	return collection.Insert(data)
}

func (db *DbState) AuthenticateUser(user LoginUser) bool {
	collection := db.GetCollection(TableUsers)
	//var storedUser User
	count, err := collection.Find(bson.M{"username": user.Username, "password": user.Password}).Count() //.One(&storedUser)
	if count < 1 || err != nil {
		return false
	}
	return true
}

func (db *DbState) IsExistingUser(user SignUpUser) (*string, error) {
	collection := db.GetCollection(TableUsers)
	rtrString := new(string)
	rtrNil := true
	count, err := collection.Find(bson.M{"username": user.Username}).Count()
	if count > 0 {
		*rtrString = "username"
		rtrNil = false
	} else if err != nil {
		return nil, err
	}
	count, err = collection.Find(bson.M{"email": user.Email}).Count()
	if count > 0 || err != nil {
		*rtrString = *rtrString + "email"
		rtrNil = false
	} else if err != nil {
		return nil, err
	}
	if rtrNil {
		return nil, nil
	}
	return rtrString, nil
}
