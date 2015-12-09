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

const database string = "markovianomatic"

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
		// TODO: add namespace to Collections
		if x != "system.indexes" {
			names = append(names, x)
		}
	}
	return
}

func Connect(cn string) (sess *mgo.Session, coll *mgo.Collection) {
	sess, _ = connect()
	coll = sess.DB(database).C(cn)

	index := mgo.Index{
		Key:        []string{"key"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     false,
	}

	err := coll.EnsureIndex(index)
	if err != nil {
		panic(err)
	}

	return
}

func Database() *mgo.Database {
	sess, _ := connect()
	return sess.DB("markovianomatic")
}

type NewNodeInfo struct {
	K string
	V []string
}

func NewNode(ni NewNodeInfo) *Node {
	return &Node{
		Id:      bson.NewObjectId(),
		Key:     ni.K,
		Choices: ni.V}
}

// Save persist a node into the given collection.
// If the node exists already, the choices are pushed, keeping the
// statistica weight (no unique/sort/deletion are made)
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
		fmt.Fprintf(os.Stderr, "Error upserting node entry: %s\n", operr.Error())
		return false
	}
	return true
}

//func withDBContext(fn func(db DB)) error {
// get a db connection from the connection pool
// dbConn := NewDB()

// return fn(dbConn)
// }
