package main

import "github.com/globalsign/mgo/bson"

// Category - Shallow category, not containing other than id to reference topics
type Category struct {
	ID     bson.ObjectId   `bson:"_id,omitempty"`
	Name   string          `json:"name"`
	Topics []bson.ObjectId `json:"topics"`
}

// CategoryWithTopics category struct as they are stored in db
type CategoryWithTopics struct {
	ID     bson.ObjectId `bson:"_id,omitempty"`
	Name   string        `json:"name"`
	Topics []Topic       `json:"topics"`
}

// TopicAndCategory Topic as they are stored in db
type TopicAndCategory struct {
	ID   bson.ObjectId `bson:"_id,omitempty"`
	Name string        `json:"name"`
	Topic
}

// Topic within a category
type Topic struct { //TODO unify with database structs
	ID       bson.ObjectId `bson:"_id,omitempty"`
	Title    string        `json:"title"`
	Content  string        `json:"content"`
	Comments []Comment     `json:"comments"`
	Username string        `json:"username"` //user
}

// Comment within a post
type Comment struct {
	Text     string `json:"text"`
	Username string `json:"username"`
	ReplyTo  int    `json:"replyto"`
}

// ReCaptchaResponse struct that contains response from google
type ReCaptchaResponse struct {
	Success   bool     `json:"success"`
	Errorcode []string `json:"error-codes"`
}
