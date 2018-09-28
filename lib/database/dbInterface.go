package database

import (
	"fmt"
	"github.com/globalsign/mgo/bson"
	"log"
	"os"
	"strings"

	"github.com/globalsign/mgo"
	// mgo "gopkg.in/mgo.v2"
)

type Db interface {
	InitState()
	CreateSession() (error)
	ValidateSession() (error)
	InsertToCollection(collectionName string, data interface{}) (error)
	ValidateUser(user User) (bool)
}

type DbState struct {
	Hosts    string
	DbName   string
	Username string
	Password string
	Session  *mgo.Session
}

type Category struct {
	Id    bson.ObjectId `bson:"_id,omitempty"`
	Name  string        `json:"name"`
	Posts int           `json:"posts"`
}

type User struct {
	Id    bson.ObjectId `bson:"_id,omitempty"`
	Email  string       `json:"email"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
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

	db.Session, err = mgo.Dial(url)

	if db.Session == nil {
		log.Fatal("Session was nil")
	}

	if err != nil {
		// fmt.Print(dialInfo)
		fmt.Errorf("died on error: %+v", err)
	}
	fmt.Println(1.4)
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

func (db *DbState) InsertToCollection(collectionName string, data interface{}) (error) {
	collection := db.GetCollection(collectionName)
	return collection.Insert(data)
}

func (db *DbState) ValidateUser(user User) (bool) {
	collection := db.GetCollection("users")
	//var storedUser User
	count, _ :=collection.Find(bson.M{"username": user.Username, "password": user.Password}).Count()//.One(&storedUser)
	if count < 1 {
		return false
	}
	return true
}