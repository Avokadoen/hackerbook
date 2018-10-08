package database

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

const (
	//DATABASE TABLES
	TableCategory   = "category"
	TableUser       = "user"
	TableTopic      = "topic"
	TableCookie     = "cookie"
	TableComment    = "comment"
	//TableEmailToken = "eToken"
	TableAdmin      = "admin"
)

const (
	// COOKIE CONST
	CookieName       = "HackerBook"
	CookieExpiration = time.Hour * 24
)

type Db interface { //TODO: split interface on type of access
	InitState()
	CreateMainSession() error
	CreateSessionPtr() (*mgo.Session, error)
	ValidateMainSession() error
	InsertToCollection(collectionName string, data interface{}, session *mgo.Session) error
	AuthenticateUser(user LoginUser, session *mgo.Session) bson.ObjectId
	AuthenticateAdmin(userID bson.ObjectId, session *mgo.Session) bson.ObjectId
	IsExistingUser(user SignUpUser, session *mgo.Session) (*string, error)
	GetCookie(cookie CookieData, entry *CookieData, session *mgo.Session)
	DeleteCookie(id bson.ObjectId, session *mgo.Session)
	GetUsername(id bson.ObjectId, session *mgo.Session) string
	GetCategories(categories interface{}, session *mgo.Session) error
	GetCategory(categoryName string, category interface{}, session *mgo.Session) error
	IsExistingCategory(categoryName string, session *mgo.Session) bool
	GetTopic(categoryName string, topicID string, topic interface{}, session *mgo.Session) error
	CreateTopic(categoryName string, topic Topic, session *mgo.Session) error
	PushTopicComment(topicID string, comment Comment, session *mgo.Session) error
}

type DbState struct {
	Hosts    string
	DbName   string
	Username string
	Password string
	Session  *mgo.Session
}

type SignUpUser struct {
	ID       bson.ObjectId `bson:"_id,omitempty" valid:"-"`
	Email    string        `json:"email" valid:"email, required"`
	Username string        `json:"username" valid:"alphanum, required"`
	Password string        `json:"password" valid:"alphanum, required"`
	Response string        `json:"captcha" valid:"ascii, required"`
}

type AdminUser struct {
	ID     bson.ObjectId `bson:"_id,omitempty" valid:"-"`
	UserID bson.ObjectId `json:"userID" valid:"-, required"`
}

type EmailToken struct { // Unverified emails
	Username string `json:"username" valid:"alphanum, required"`
	Token    string `json:"token" valid:"alphanum, required"`
}

type LoginUser struct {
	Username string `json:"username" valid:"alphanum, required"`
	Password string `json:"password" valid:"alphanum, required"`
}

type CookieData struct {
	ID    bson.ObjectId `json:"userid" valid:"-, required"`
	Token string        `json:"token" valid:"alphanum, required"`
}

type Category struct {
	ID     bson.ObjectId   `bson:"_id,omitempty"`
	Name   string          `json:"name"`
	Topics []bson.ObjectId `json:"topics"`
	//MORE?
}

/*type Category struct {
	ID       bson.ObjectId `bson:"_id,omitempty" valid:"-, optional"`
	Name	 string		   `json:"name" valid:"alphanum, required"`
}*/

type Topic struct {
	ID       bson.ObjectId `bson:"_id" valid:"-"`
	Category string        `json:"name" valid:"alphanum, required"`
	Username string        `json:"username" valid:"alphanum, required"`
	Title    string        `json:"title" valid:"printableascii, required"`
	Content  string        `json:"content" valid:"halfwidth"`
}

type Comment struct {
	CommentID bson.ObjectId `bson:"_id,omitempty" valid:"-"`
	Username  string        `json:"username" valid:"alphanum, required"`
	Text      string        `json:"text" valid:"halfwidth"`
	ReplyTo   int           `json:"replyto" valid:"-"`
}

func (db *DbState) InitState() {
	db.Hosts = os.Getenv("DBURL")
	db.DbName = os.Getenv("DBNAME")
	db.Username = os.Getenv("DBUSERNAME")
	db.Password = os.Getenv("DBPASSWORD")

	//
	/*log.Printf("%+v\n", db.Hosts)
	log.Printf("%+v\n", db.DbName)
	log.Printf("%+v\n", db.Username)
	log.Printf("%+v\n", db.Password)*/
}

func (db *DbState) CreateMainSession() (err error) {

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

	err = db.EnsureAllIndices()
	if err != nil {
		return fmt.Errorf("died on error: %+v", err)
	}

	return nil
}

func (db *DbState) CreateSessionPtr() (*mgo.Session, error) {
	if db.Session == nil {
		db.CreateMainSession()
		if db.Session == nil {
			return nil, fmt.Errorf("failed to recover session in createsessionptr")
		}
	}
	sessionPtr := db.Session.Copy()
	return sessionPtr, nil
}

func (db *DbState) EnsureAllIndices() error {

	categoryIndex := mgo.Index{
		Key:        []string{"name"},
		Unique:     true,
		DropDups:   true,
		Background: false,
		Sparse:     false,
	}
	collCategory := db.getCollection(TableCategory, db.Session)
	err := collCategory.DropAllIndexes()
	if err != nil {
		return fmt.Errorf("DropAllIndexes\n category failed, err: %+v", err)
	}
	err = collCategory.EnsureIndex(categoryIndex)
	if err != nil {
		return fmt.Errorf("EnsureAllIndices\n category failed, err: %+v", err)
	}
	userIndex := mgo.Index{
		Key:        []string{"username"},
		Unique:     true,
		DropDups:   true,
		Background: false,
		Sparse:     false,
	}
	collUser := db.getCollection(TableUser, db.Session)
	err = collUser.DropAllIndexes()
	if err != nil {
		return fmt.Errorf("DropAllIndexes\n user failed, err: %+v", err)
	}
	err = collUser.EnsureIndex(userIndex)
	if err != nil {
		return fmt.Errorf("EnsureAllIndices\n user failed, err: %+v", err)
	}
	cookieIndex := mgo.Index{
		Key:         []string{"token"},
		Unique:      true,
		DropDups:    true,
		Background:  false,
		Sparse:      false,
		ExpireAfter: CookieExpiration,
	}
	collCook := db.getCollection(TableCookie, db.Session)
	err = collCook.DropAllIndexes()
	if err != nil {
		return fmt.Errorf("DropAllIndexes\n cookie failed, err: %+v", err)

	}
	err = collCook.EnsureIndex(cookieIndex)
	if err != nil {
		return fmt.Errorf("EnsureAllIndices\n cookie failed, err: %+v", err)
	}

	/*topicIndex := mgo.Index{
		Key: []string{"_id"},
		Unique: true,
		DropDups: true,
		Background: false,
		Sparse:     false,
	}
	collTopic := db.getCollection(TableTopic, db.Session)
	err = collCook.DropAllIndexes()
	if err != nil {
		return fmt.Errorf("DropAllIndexes\n topic failed, err: %+v", err)
	}
	err = collTopic.EnsureIndex(topicIndex)
	if err != nil {
		return fmt.Errorf("EnsureAllIndices\n topic failed, err: %+v", err)
	}*/
	return nil
}

func (db *DbState) ValidateMainSession() error {
	if db.Session != nil {
		return nil
	}
	err := db.CreateMainSession()
	if err != nil {
		return err
	}
	return nil
}

func (db *DbState) getCollection(collectionName string, session *mgo.Session) *mgo.Collection {
	if session == nil {
		println("session was nil in InsertToCollection")
		return nil
	}
	return session.DB(db.DbName).C(strings.ToLower(collectionName))
}

func (db *DbState) InsertToCollection(collectionName string, data interface{}, session *mgo.Session) error {
	if session == nil {
		return fmt.Errorf("session was nil in InsertToCollection")
	}
	collection := db.getCollection(collectionName, session)
	return collection.Insert(data)
}

func (db *DbState) AuthenticateUser(user LoginUser, session *mgo.Session) bson.ObjectId {
	collection := db.getCollection(TableUser, session)
	var storedUser SignUpUser
	err := collection.Find(bson.M{"username": user.Username, "password": user.Password}).One(&storedUser)
	if err != nil {
		log.Printf("%+v", err)
		return bson.ObjectId(0)
	}
	return storedUser.ID
}

func (db *DbState) AuthenticateAdmin(userID bson.ObjectId, session *mgo.Session) bson.ObjectId {
	collection := db.getCollection(TableAdmin, session)
	var adminUser AdminUser
	err := collection.Find(bson.M{"userID": userID.Hex()}).One(&adminUser)
	if err != nil {
		log.Printf("%+v", err)
		return bson.ObjectId(0)
	}
	return adminUser.ID
}

func (db *DbState) IsExistingUser(user SignUpUser, session *mgo.Session) (*string, error) {
	collection := db.getCollection(TableUser, session)
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

func (db *DbState) GetCookie(cookie CookieData, entry *CookieData, session *mgo.Session) {

	collection := db.getCollection(TableCookie, session)
	err := collection.Find(bson.M{"id": bson.ObjectIdHex(cookie.ID.Hex())}).One(&entry)
	if err != nil {
		fmt.Printf("when retrieving cookie error: %+v", err)
	}

}

func (db *DbState) DeleteCookie(id bson.ObjectId, session *mgo.Session) {
	collection := db.getCollection(TableCookie, session)
	// TODO: log?
	_, err := collection.RemoveAll(bson.M{"id": bson.ObjectIdHex(id.Hex())})
	if err != nil {
		fmt.Printf("when deleting cookies error: %+v", err)
	}
}

func (db *DbState) GetUsername(id bson.ObjectId, session *mgo.Session) string {
	if session == nil {
		return "<bad boi>"
	}
	user := LoginUser{Username: "<bad boi>"}
	collection := db.getCollection(TableUser, session)
	err := collection.FindId(bson.ObjectIdHex(id.Hex())).One(&user)
	if err != nil {
		fmt.Printf("when retrieving username error: %+v, id: %+v", err, bson.ObjectIdHex(id.Hex()))
	}
	return user.Username
}

//TODO sepparate into other file
func (db *DbState) GetCategories(categories interface{}, session *mgo.Session) error {
	if session == nil {
		return fmt.Errorf("nil session in get getcategories")
	}
	return db.getCollection(TableCategory, session).Find(nil).All(categories)
}
func (db *DbState) GetCategory(categoryName string, category interface{}, session *mgo.Session) error {
	if session == nil {
		return fmt.Errorf("nil session in get getcategory")
	}
	pipe := db.getCollection(TableCategory, session).Pipe(
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
	return pipe.One(category)
}

func (db *DbState) IsExistingCategory(categoryName string, session *mgo.Session) bool {
	if session == nil {
		fmt.Printf("session was nil")
		return true
	}

	collection := db.getCollection(TableCategory, session)

	//TODO: Fix find function? It says category already exists?
	count, err := collection.Find(bson.M{"name": categoryName}).Count()
	if err != nil {
		log.Printf("Category doesn't exist, err?: %+v", err)
		return true
	}

	if count > 0 {
		fmt.Printf("Category already exists!")
		return true
	} else {
		fmt.Printf("Category doesn't exist")
		return false
	}

}

func (db *DbState) GetTopic(categoryName string, topicID string, topic interface{}, session *mgo.Session) error {
	if session == nil {
		return fmt.Errorf("nil session in get GetTopic")
	}
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
	pipe := db.getCollection(TableCategory, session).Pipe(pipeline)
	return pipe.One(topic)
}

func (db *DbState) PushTopicComment(topicID string, comment Comment, session *mgo.Session) error {
	if session == nil {
		return fmt.Errorf("nil session in get PushTopicComment")
	}
	selector := bson.M{"_id": bson.ObjectIdHex(topicID)}

	update := bson.M{"$push": bson.M{"comments": bson.M{"$each": []Comment{comment}}}}

	return db.getCollection(TableTopic, session).Update(selector, update)
}
func (db *DbState) CreateTopic(categoryName string, topic Topic, session *mgo.Session) error {

	if session == nil {
		return fmt.Errorf("nil session in get CreateTopic")
	}
	topic.ID = bson.NewObjectId()
	db.InsertToCollection(TableTopic, topic, session)

	selector := bson.M{"name": categoryName}
	update := bson.M{"$push": bson.M{"topics": bson.M{"$each": []bson.ObjectId{topic.ID}}}}

	return db.getCollection(TableCategory, session).Update(selector, update)
}
