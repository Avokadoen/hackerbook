package main

import "gopkg.in/mgo.v2/bson"

//Category - Shallow category, not containing other than id to reference topics
type Category struct {
	ID     bson.ObjectId   `bson:"_id,omitempty"`
	Name   string          `json:"name"`
	Topics []bson.ObjectId `json:"topics"`
	//MORE?
}

type CategoryWithTopics struct {
	ID     bson.ObjectId `bson:"_id,omitempty"`
	Name   string        `json:"name"`
	Topics []Topic       `json:"topics"`
	//MORE?
}

/*
	"_id" : ObjectId("5bb175765499851637a9379d"),
	"name" : "phishing",
	"topic" : {
		"_id" : ObjectId("5bb177bc5499851637a9379e"),
		"title" : "Test Post Pls Ignore",
		"content" : "test ok",
		"comments" : [ ],
		"createdBy" : ObjectId("5bb0ed24ed8bad61aa93bd85"),
		"creationTime" : ISODate("2018-10-01T01:26:20.214Z")
	}

*/
type TopicAndCategory struct {
	ID   bson.ObjectId `bson:"_id,omitempty"`
	Name string        `json:"name"`
	Topic
}

//Topic within a category
type Topic struct { //TODO unify with database structs
	ID       bson.ObjectId `bson:"_id,omitempty"`
	Title    string        `json:"title"`
	Content  string        `json:"content"`
	Comments []Comment     `json:"comments"`
	Username string        `json:"username"` //user
}

//Comment within a post
type Comment struct {
	Text     string `json:"text"`
	Username string `json:"username"`
	ReplyTo  int    `json:"replyto"`
}

//Struct for captcha
type ReCaptcha struct {
	Secret   string `json:"secret"`
	Response string `json:"response"`
}

type ReCaptchaResponse struct {
	Success   bool     `json:"success"`
	Errorcode []string `json:"error-codes"`
}
