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

//DATABASE TABLES
const (

	// TableCategory ...
	TableCategory = "category"
	// TableUser ...
	TableUser     = "user"
	// TableTopic ...
	TableTopic    = "topic"
	// TableCookie ...
	TableCookie   = "cookie"
	// TableComment ...
	TableComment  = "comment"
	//TableEmailToken = "eToken"
	// TableAdmin ...
	TableAdmin = "admin"
)

// COOKIE CONST
const (
	// CookieName ...
	CookieName       = "HackerBook"
	// CookieExpiration defines lifespan of cookie
	CookieExpiration = time.Hour * 24
)

// Db defines our DB API
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

// DbState contains the connection data for our mongodb
type DbState struct {
	Hosts    string
	DbName   string
	Username string
	Password string
	Session  *mgo.Session
}

// SignUpUser is a json struct for sign up form post
type SignUpUser struct {
	ID       bson.ObjectId `bson:"_id,omitempty" valid:"-"`
	Email    string        `json:"email" valid:"email, required"`
	Username string        `json:"username" valid:"alphanum, required"`
	Password string        `json:"password" valid:"alphanum, required"`
	Response string        `json:"captcha" valid:"ascii, required"`
}

// AdminUser is json struct for a user that also is an admin
type AdminUser struct {
	ID     bson.ObjectId `bson:"_id,omitempty" valid:"-"`
	UserID bson.ObjectId `json:"userID" valid:"-, required"`
}

// EmailToken unused
/*
type EmailToken struct { // Unverified emails
	Username string `json:"username" valid:"alphanum, required"`
	Token    string `json:"token" valid:"alphanum, required"`
}*/

// LoginUser is json struct for users manually login in
type LoginUser struct {
	Username string `json:"username" valid:"alphanum, required"`
	Password string `json:"password" valid:"alphanum, required"`
}

// CookieData is json struct for cookie data to be stored
type CookieData struct {
	ID    bson.ObjectId `json:"userid" valid:"-, required"`
	Token string        `json:"token" valid:"alphanum, required"`
}

// Category is json struct to contain data relevant for a forum category
type Category struct {
	ID     bson.ObjectId   `bson:"_id,omitempty" valid:"-"`
	Name   string          `json:"name" valid:"printableascii, required"`
	Topics []bson.ObjectId `json:"topics" valid:"-"`
}

/*type Category struct {
	ID       bson.ObjectId `bson:"_id,omitempty" valid:"-, optional"`
	Name	 string		   `json:"name" valid:"alphanum, required"`
}*/

// Topic is json struct to contain data relevant for a forum topic
type Topic struct {
	ID       bson.ObjectId `bson:"_id" valid:"-"`
	Category string        `json:"name" valid:"alphanum, required"`
	Username string        `json:"username" valid:"alphanum, required"`
	Title    string        `json:"title" valid:"printableascii, required"`
	Content  string        `json:"content" valid:"halfwidth"`
}

// Comment is json struct to contain data relevant for a forum comment
type Comment struct {
	CommentID bson.ObjectId `bson:"_id,omitempty" valid:"-"`
	Username  string        `json:"username" valid:"alphanum, required"`
	Text      string        `json:"text" valid:"halfwidth"`
	ReplyTo   int           `json:"replyto" valid:"-"`
}

// InitState retrieves environment variables and stores them in a DbState
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

// CreateMainSession dials the mongodb with DbState data and create main session
// Returns error if failed
func (db *DbState) CreateMainSession() (err error) {

	url := fmt.Sprintf("mongodb://%s:%s@%s/%s", db.Username, db.Password, db.Hosts, db.DbName)

	db.Session, err = mgo.Dial(url)

	if db.Session == nil {
		log.Fatal("Session was nil")
	}
	if err != nil {
		return fmt.Errorf("died on error: %+v", err)
	}

	err = db.EnsureAllIndices()
	if err != nil {
		return fmt.Errorf("died on error: %+v", err)
	}

	return nil
}

// CreateSessionPtr copies the DbState session. Remember to close returned
// pointer value when done with it
// Returns the copy of the session and an error if failed
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

// EnsureAllIndices defines index values in the DB to ensure
// certain properties of the db
// Returns an error if failed
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

	return nil
}

// ValidateMainSession check if DbState session is still valid
// Will also attempt to recover a new session if invalid
// Returns error if failed to validate
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

// getCollection will return a collection pointer it found said collection
func (db *DbState) getCollection(collectionName string, session *mgo.Session) *mgo.Collection {
	if session == nil {
		return nil
	}
	return session.DB(db.DbName).C(strings.ToLower(collectionName))
}

// InsertToCollection will attempt to insert interface into collection with collectionName as name
// Returns error if failed to insert
func (db *DbState) InsertToCollection(collectionName string, data interface{}, session *mgo.Session) error {
	if session == nil {
		return fmt.Errorf("session was nil in InsertToCollection")
	}
	collection := db.getCollection(collectionName, session)
	return collection.Insert(data)
}

// AuthenticateUser will verify if sent user actually is a stored user
// Returns id of user if user was authentic
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

// AuthenticateAdmin will verify if sent admin actually is a stored admin
// Returns id of admin if admin was authentic
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

// IsExistingUser checks if certain parts of a sign-up form is already used
// by another user
// Returns a string with what is in use and error if something went wrong
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

// GetCookie tries to insert cookie data into the entry parameter
func (db *DbState) GetCookie(cookie CookieData, entry *CookieData, session *mgo.Session) {

	collection := db.getCollection(TableCookie, session)
	err := collection.Find(bson.M{"id": bson.ObjectIdHex(cookie.ID.Hex())}).One(&entry)
	if err != nil {
		fmt.Printf("when retrieving cookie error: %+v", err)
	}

}

// DeleteCookie deletes all cookies that share id with parameter id
func (db *DbState) DeleteCookie(id bson.ObjectId, session *mgo.Session) {
	collection := db.getCollection(TableCookie, session)
	// TODO: log?
	_, err := collection.RemoveAll(bson.M{"id": bson.ObjectIdHex(id.Hex())})
	if err != nil {
		fmt.Printf("when deleting cookies error: %+v", err)
	}
}

// GetUsername retrieves username of user that share id with parameter id
// Returns username or <bad boi> if failed
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

// GetCategories retrieves all categories in the collection
// Returns error if session was nil
func (db *DbState) GetCategories(categories interface{}, session *mgo.Session) error {
	if session == nil {
		return fmt.Errorf("nil session in get getcategories")
	}
	return db.getCollection(TableCategory, session).Find(nil).All(categories)
}

// GetCategory retrieves a category from the collection
// Returns error if it failed
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

// IsExistingCategory verifies that said category does not already exist
// Returns false if it does not exist, true otherwise
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

// GetTopic retrieves a topic from collection
// Returns an error if session was nil
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

// PushTopicComment inserts comment into topic collection
// Returns error if session is nil or if it failed to update collection
func (db *DbState) PushTopicComment(topicID string, comment Comment, session *mgo.Session) error {
	if session == nil {
		return fmt.Errorf("nil session in get PushTopicComment")
	}
	selector := bson.M{"_id": bson.ObjectIdHex(topicID)}

	update := bson.M{"$push": bson.M{"comments": bson.M{"$each": []Comment{comment}}}}

	return db.getCollection(TableTopic, session).Update(selector, update)
}

// CreateTopic creates a new topic in the db
// Returns error if session is nil or if it failed to update collection
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
