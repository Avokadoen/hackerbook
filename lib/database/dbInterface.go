package database

import (
	"fmt"
	"log"
	"os"

	"github.com/globalsign/mgo"
	// mgo "gopkg.in/mgo.v2"
)

type Db interface {
	InitState()
	CreateSession(url string) (*mgo.Session, error)
	GetCollection(session *mgo.Session) (mgo.Collection, error)
	ValidateSession(session *mgo.Session) (*mgo.Session, error)
}

type DbState struct {
	Hosts    []string
	DbName   string
	Username string
	Password string
	Session  *mgo.Session
}

func (db *DbState) InitState() {
	db.Hosts = append(db.Hosts, os.Getenv("DBURL"))
	// db.Hosts = append(db.Hosts, os.Getenv("DBURL2"))
	// db.Hosts = append(db.Hosts, os.Getenv("DBURL3"))
	db.DbName = os.Getenv("DBNAME")
	db.Username = os.Getenv("DBUSERNAME")
	db.Password = os.Getenv("DBPASSWORD")

	fmt.Printf("%+v\n", db.Hosts)
	fmt.Printf("%+v\n", db.DbName)
	fmt.Printf("%+v\n", db.Username)
	fmt.Printf("%+v\n", db.Password)
}

func (db *DbState) CreateSession() (err error) {
	url := fmt.Sprintf("mongodb://%s:%s@%s/%s?ssl=true", db.Username, db.Password, db.Hosts[0], db.DbName)
	// dialInfo, err := mgo.ParseURL(url)
	if err != nil {
		fmt.Println(err)
	}
	db.Session, err = mgo.Dial(url)
	if db.Session == nil {
		log.Fatal("Session was nil")
	}
	fmt.Println(1.25)
	// fmt.Println(db.Session)
	fmt.Println(1.3)
	if err != nil {
		// fmt.Print(dialInfo)
		fmt.Errorf("died on error: %+v", err)
	}
	fmt.Println(1.4)
	return err
}

func (db *DbState) GetCollection(collectionName string) *mgo.Collection {
	return db.Session.DB(db.DbName).C(collectionName)
}

func (db *DbState) ValidateSession() error {
	if db.Session != nil {
		return nil
	}
	err := db.CreateSession()
	if err != nil {
		return err
	}
	return fmt.Errorf("what the fuk, session: %+v", db.Session)
}
