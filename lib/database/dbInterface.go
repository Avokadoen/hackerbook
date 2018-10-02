package database

import (
	"encoding/json"
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
	TableTopic    = "topic"
	TableCookie   = "cookie"
)

type Db interface {
	InitState()
	CreateSession() error
	ValidateSession() error
	InsertToCollection(collectionName string, data interface{}) error
	AuthenticateUser(user LoginUser) bson.ObjectId
	//AuthenticateUserCookie(cookie http.Cookie) bson.ObjectId
	IsExistingUser(user SignUpUser) (*string, error)
	GetCookie(cookie CookieData, entry *CookieData)
	DeleteCookie(id bson.ObjectId)
	GetUsername(id bson.ObjectId) string
	GetCategories(categories interface{})
	GetCategory(categoryName string, category interface{})
	GetTopic(categoryName string, topicID string, topic interface{})
}

type DbState struct {
	Hosts    string
	DbName   string
	Username string
	Password string
	Session  *mgo.Session
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

type CookieData struct {
	Id    bson.ObjectId `json:"token" valid:"-, required"`
	Token string        `json:"token" valid:"alphanum, required"`
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

func (db *DbState) getCollection(collectionName string) *mgo.Collection {
	return db.Session.DB(db.DbName).C(strings.ToLower(collectionName))
}

func (db *DbState) InsertToCollection(collectionName string, data interface{}) error {
	collection := db.getCollection(collectionName)
	return collection.Insert(data) //TODO ensureIndex
}

func (db *DbState) AuthenticateUser(user LoginUser) bson.ObjectId {
	collection := db.getCollection(TableUsers)
	var storedUser SignUpUser
	err := collection.Find(bson.M{"username": user.Username, "password": user.Password}).One(&storedUser)
	if err != nil {
		log.Printf("%+v", err)
		return bson.ObjectId(0)
	}
	return storedUser.Id
}

/*func (db *DbState) AuthenticateUserCookie(cookie CookieData) bson.ObjectId {
	collection := db.getCollection(TableCookie)
	var dbCookieData CookieData
	err := collection.Find(bson.M{"id":cookie.Id, "token":cookie.Token}).One(&dbCookieData)
	if err != nil {
		log.Printf("%+v", err)
		return bson.ObjectId(0)
	}

}*/

func (db *DbState) IsExistingUser(user SignUpUser) (*string, error) {
	collection := db.getCollection(TableUsers)
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

func (db *DbState) GetCookie(cookie CookieData, entry *CookieData) {

	collection := db.getCollection(TableCookie)
	err := collection.Find(bson.M{"id": bson.ObjectIdHex(cookie.Id.Hex())}).One(&entry)
	if err != nil {
		fmt.Printf("when retrieving cookie error: %+v", err)
	}

}

func (db *DbState) DeleteCookie(id bson.ObjectId) {
	collection := db.getCollection(TableCookie)
	// TODO: log?
	_, err := collection.RemoveAll(bson.M{"id": bson.ObjectIdHex(id.Hex())})
	if err != nil {
		fmt.Printf("when deleting cookies error: %+v", err)
	}
}

func (db *DbState) GetUsername(id bson.ObjectId) string {
	user := LoginUser{}
	collection := db.getCollection(TableUsers)
	err := collection.FindId(bson.ObjectIdHex(id.Hex())).One(&user)
	if err != nil {
		fmt.Printf("when retrieving username error: %+v, id: %+v", err, bson.ObjectIdHex(id.Hex()))
	}
	return user.Username
}

//TODO sepparate into other file
func (db *DbState) GetCategories(categories interface{}) {
	db.ValidateSession()
	db.getCollection(TableCategory).Find(nil).All(categories)
}
func (db *DbState) GetCategory(categoryName string, category interface{}) {
	db.ValidateSession()

	pipe := db.getCollection(TableCategory).Pipe(
		[]bson.M{
			{"$match": bson.M{"name": bson.M{"$eq": categoryName}}},
			{
				"$lookup": bson.M{
					"from":         TableTopic,
					"localField":   "topics",
					"foreignField": "_id",
					"as":           "topics",
				},
			},
		},
	)
	err := pipe.One(category)
	if err != nil {
		fmt.Printf(err.Error())
		return
	}
	fmt.Printf("\n\n%+v\n\n", category)
}

func (db *DbState) GetTopic(categoryName string, topicID string, topic interface{}) {
	db.ValidateSession()

	pipeline := []bson.M{
		{"$unwind": "$topics"},
		{"$match": bson.M{"name": bson.M{"$eq": categoryName}, "topics": bson.M{"$eq": bson.ObjectIdHex(topicID)}}},
		{"$lookup": bson.M{
			"from":         TableTopic,
			"localField":   "topics",
			"foreignField": "_id",
			"as":           "topic",
		},
		},
		{"$project": bson.M{"topics": 0}},
		{"$unwind": "$topic"},
	}
	/*
	   	RESULTS IN THIS KIND OF STRUCTURE
	   {
	   	Id:ObjectIdHex("5bb175765499851637a9379d")
	   	Name:phishing
	   	Topic:{
	   		Id:ObjectIdHex("5bb177bc5499851637a9379e")
	   		Title:Test Post Pls Ignore
	   		Content:test ok
	   		Comments:[]
	   		CreatedBy:ObjectIdHex("")
	   	}
	   }
	*/

	b, err := json.MarshalIndent(pipeline, "", "  ")
	if err != nil {
		fmt.Printf("%+v\n", pipeline)
	}
	fmt.Print(string(b))

	//The following line does not validate against category... which might not really be an issue
	pipe := db.getCollection(TableCategory).Pipe(pipeline)
	pipe.One(topic)
}
