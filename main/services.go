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

	sse	map[*SSEClient]bool
	register chan *SSEClient
	unregister chan *SSEClient
	broadcast chan *Problem
}

type SSEClient struct {
	problemChan chan *Problem
	service *Service
}

func NewSSEClient(s *Service) (c *SSEClient) {
	c = &SSEClient{
		problemChan: make(chan *Problem, 150),
		service:      s,
	}

	c.service.register <- c

	return
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

	s.sse = make(map[*SSEClient]bool)
	s.register = make(chan *SSEClient)
	s.unregister = make(chan *SSEClient)
	s.broadcast = make(chan *Problem)

	go s.runSSE()
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
	options.SetSort(bson.D{{ "lastattempted", 1 }})

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
		p.LastAttempted = p.LastAttempted.Local()
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
			{"$push", bson.D{{"tags", bson.D{{"$each", p.Tags}}}}},
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

	var data Problem
	s.problems.FindOne(
		context.TODO(),
		bson.M{"id": p.Id},
	).Decode(&data)

	data.LastAttempted = data.LastAttempted.Local()
	s.broadcast <- &data

	return
}

func (s *Service) Broadcast(p *Problem) {
	for c := range s.sse {
		select {
		case c.problemChan <- p:
		default:
			close(c.problemChan)
			delete(s.sse, c)
		}
	}
}

func (s *Service) runSSE() {
	for {
		select {
		case c := <-s.register:
			s.sse[c] = true
		case c := <-s.unregister:
			close(c.problemChan)
			delete(s.sse, c)
		case p := <-s.broadcast:
			s.Broadcast(p)
		}
	}
}
