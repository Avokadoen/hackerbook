package db

import (
	"github.com/globalsign/mgo"
	"log"
	"os"
)

type Db interface {
	InitDb()
	CreateSession(url string) (*mgo.Session, error)
	GetCollection(session *mgo.Session) (mgo.Collection, error)
	ValidateSession(session *mgo.Session) (*mgo.Session, error)
}

type DbState struct {
	Url []string
	DbName string
	Username string
	Password string
}

func (db *DbState) InitDb() {
	db.Url = append(db.Url, os.Getenv("DBURL1"))
	db.Url = append(db.Url, os.Getenv("DBURL2"))
	db.Url = append(db.Url, os.Getenv("DBURL3"))
	db.DbName = os.Getenv("DBNAME")
	db.Username = os.Getenv("DBUSERNAME")
	db.Password = os.Getenv("DBPASSWORD")
}

func (db *DbState) CreateSession() (*mgo.Session, error){
	dialInfo :=
		&mgo.DialInfo{
			Addrs:    db.Url,
			Username: db.Username,
			Password: db.Password,
			Timeout: 2000,
	}
	session, err := mgo.DialWithInfo(dialInfo)//.Dial(db.Url)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return session, nil
}

func (db *DbState) GetCollection(session *mgo.Session, collectionName string) (*mgo.Collection) {
	return session.DB(db.DbName).C(collectionName)
}

func (db *DbState) ValidateSession(session *mgo.Session) (*mgo.Session, error) {
	if session != nil {
		return session, nil
	}
	session, err := db.CreateSession()
	if err != nil{
		return nil, err
	}
	return session, nil
}