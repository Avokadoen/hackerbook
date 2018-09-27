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
	db.Hosts = append(db.Hosts, os.Getenv("DBURL1"))
	db.Hosts = append(db.Hosts, os.Getenv("DBURL2"))
	db.Hosts = append(db.Hosts, os.Getenv("DBURL3"))
	db.DbName = os.Getenv("DBNAME")
	db.Username = os.Getenv("DBUSERNAME")
	db.Password = os.Getenv("DBPASSWORD")

	fmt.Printf("%+v\n", db.Hosts)
	fmt.Printf("%+v\n", db.DbName)
	fmt.Printf("%+v\n", db.Username)
	fmt.Printf("%+v\n", db.Password)
}

func (db *DbState) CreateSession() (err error) {
	url := "mongodb://localhost"
	// url := "mongodb+srv://master:FQjHYATFwBhpOT8t@cluster0-jgrwo.mongodb.net/forum?ssl=true"
	dialInfo, err := mgo.ParseURL(url)
	dialInfo.Database = "forum"
	if err != nil {
		log.Fatalf("Parsing failed with error: %+v", err)
	}
	// url := "mongodb+srv://master:3tm1BK2II9plEqL3@cluster0-jgrwo.mongodb.net/forum?ssl=true"
	// fmt.Println(1.1)
	// dialInfo := &mgo.DialInfo{
	// 	Addrs:    db.Hosts,
	// 	Username: "master",
	// 	Password: "3tm1BK2II9plEqL3",
	// 	Database: "forum",
	//
	// 	DialServer: func(addr *mgo.ServerAddr) (net.Conn, error) {
	// 		return tls.Dial("tcp", addr.String(), &tls.Config{})
	// 	},
	// 	Timeout: time.Second * 10,
	// }
	// fmt.Printf("dialInfo:\n%+v", dialInfo)

	fmt.Println(1.2)
	fmt.Printf("\n\nDialInfo:\n\n%+v", dialInfo)
	db.Session, err = mgo.DialWithInfo(dialInfo)

	//END OF DIGGING INTO DATA FOR VALIDATING

	dbs, err := db.Session.DatabaseNames()
	fmt.Println()
	for _, db := range dbs {
		fmt.Println(db)
	}
	cols, err := db.Session.DB(dbs[2]).CollectionNames()
	fmt.Println()
	for _, col := range cols {
		fmt.Println(col)
	}
	m := make(map[string]interface{})
	db.Session.DB(dbs[2]).C(cols[0]).Find(nil).One(&m)
	fmt.Println(m)

	//END OF DIGGING INTO DATA FOR VALIDATING

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
