package model

import (
	"fmt"
	"os"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Node struct {
	Id      bson.ObjectId `bson:"_id"`
	Key     string        `bson:"key"`
	Choices []string      `bson:"choices"`
}

func connect() (sess *mgo.Session, err error) {
	uri := os.Getenv("MONGODB_URL")
	if uri == "" {
		fmt.Fprintf(os.Stderr, "no connection string provided: using localhost\n")
		uri = "localhost:27017"
	}

	sess, err = mgo.Dial(uri)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't connect to mongo, go error %v\n", err)
		os.Exit(1)
	}

	sess.SetSafe(&mgo.Safe{})
	return
}

func Collections(db *mgo.Database) (names []string, err error) {
	all, err := db.CollectionNames()
	if err != nil {
		return
	}

	for _, x := range all {
		if x != "system.indexes" {
			names = append(names, x)
		}
	}
	return
}

func Connect(cn string) (sess *mgo.Session, coll *mgo.Collection) {
	sess, _ = connect()
	coll = sess.DB("markovianomatic").C(cn)
	return
}

func Database() *mgo.Database {
	sess, _ := connect()
	return sess.DB("markovianomatic")
}

func NewNode(k string, v []string) *Node {
	return &Node{
		Id:      bson.NewObjectId(),
		Key:     k,
		Choices: v}
}

func (n *Node) Save(coll *mgo.Collection) bool {
	var nn *Node
	err := coll.Find(
		bson.M{"key": n.Key}).One(&nn)

	var operr error
	if err == nil {
		operr = coll.Update(
			bson.M{"key": n.Key},
			bson.M{"$push": bson.M{"choices": bson.M{"$each": n.Choices}}})
	} else {
		operr = coll.Insert(n)
	}

	if operr != nil {
		fmt.Fprintf(os.Stderr, "Error upserting node entry: %s\n", err.Error())
	}
	return true
}

//func withDBContext(fn func(db DB)) error {
// get a db connection from the connection pool
// dbConn := NewDB()

// return fn(dbConn)
// }
