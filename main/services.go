package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	MONGO = "mongodb://localhost:27017"
)

type Service struct {
	client 	 *mongo.Client
	db       *mongo.Database
	problems *mongo.Collection
}

func NewService() (s *Service, err error) {
	s = &Service{}
	s.client, err = connect()
	if err != nil {
		return
	}

	s.db = s.client.Database("leetcode")
	s.problems = s.db.Collection("problems")

	mod := mongo.IndexModel{
		Keys: bson.M{"id": 1},
	}
	s.problems.Indexes().CreateOne(
		context.TODO(),
		mod,
	)
	return
}

func connect() (client *mongo.Client, err error) {
	clientOptions := options.Client().ApplyURI(MONGO)

	// Connect to MongoDB
	client, err = mongo.Connect(context.TODO(), clientOptions)

	if err != nil {
		return
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)

	if err != nil {
		return
	}

	log.Println("Connected to MongoDB!")

	return
}

func (s *Service) Close() {
	s.client.Disconnect(context.TODO())
}

func date(time time.Time) string {
	return fmt.Sprintf("%v %v, %v", time.Month().String()[:3], time.Day(), time.Year())
}

func (s *Service) GetData() (problems []Problem, err error) {
	var cursor *mongo.Cursor

	options := options.Find()
	options.SetSort(bson.D{{ "lastattempted", -1 }})

	cursor, err = s.problems.Find(context.TODO(), bson.D{{}}, options)
	if err != nil {
		return
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var p Problem
		if err = cursor.Decode(&p); err != nil {
			return
		}

		problems = append(problems, p)
	}

	if err = cursor.Err(); err != nil {
		return
	}

	return
}

func (s *Service) ServeAttempt(r *http.Request) (err error) {
	var b []byte
	b, err = ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}

	var p Problem
	err = json.Unmarshal(b, &p)
	if err != nil {
		return
	}

	p.LastAttempted = time.Now()
	p.Attempts = 1

	result, err := s.problems.UpdateOne(
		context.TODO(),
		bson.M{"id": p.Id},
		bson.D{
			{"$inc", bson.D{{"attempts", 1}}},
			{"$set", bson.D{{"lastattempted", time.Now()}}},
		},
	)

	if err != nil {
		return
	}

	if result.MatchedCount == 0 {
		//it is a new problem
		_, err = s.problems.InsertOne(context.TODO(), p)
		if err != nil {
			return
		}
	}
	return
}