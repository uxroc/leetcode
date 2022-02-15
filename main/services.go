package main

import (
	"context"
	"encoding/json"
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
	client   *mongo.Client
	db       *mongo.Database
	problems *mongo.Collection

	sse        map[*SSEClient]bool
	register   chan *SSEClient
	unregister chan *SSEClient
	broadcast  chan *Problem
}

type SSEClient struct {
	problemChan chan *Problem
	service     *Service
}

func NewSSEClient(s *Service) (c *SSEClient) {
	c = &SSEClient{
		problemChan: make(chan *Problem, 150),
		service:     s,
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
	_, err = s.problems.Indexes().CreateOne(
		context.TODO(),
		mod,
	)
	if err != nil {
		return
	}

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

func (s *Service) Close() error {
	return s.client.Disconnect(context.TODO())
}

func (s *Service) GetData() (problems []Problem, err error) {
	var cursor *mongo.Cursor

	options := options.Find()
	options.SetSort(bson.D{{"lastattempted", 1}})

	cursor, err = s.problems.Find(context.TODO(), bson.D{}, options)
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

	//log.Println(len(problems))

	if err = cursor.Err(); err != nil {
		return
	}

	return
}

func (s *Service) parseAttempt(r *http.Request) (p Problem, solved bool, err error) {
	var b []byte
	b, err = ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}

	var objmap map[string]json.RawMessage
	err = json.Unmarshal(b, &objmap)
	if err != nil {
		return
	}

	err = json.Unmarshal(objmap["Solved"], &solved)
	if err != nil {
		return
	}

	err = json.Unmarshal(objmap["Problem"], &p)
	if err != nil {
		return
	}

	return
}

func (s *Service) CreateProblem(r *http.Request) (err error) {
	p, solved, err := s.parseAttempt(r)
	if err != nil {
		return
	}

	p.LastAttempted = time.Now()
	p.Attempts = 1
	p.Fails = 1
	if solved {
		p.Fails = 0
	}
	p.Hide = false

	result, err := s.problems.UpdateOne(
		context.TODO(),
		bson.M{"id": p.Id},
		bson.D{
			{"$inc", bson.D{{"attempts", 1}}},
			{"$inc", bson.D{{"fails", p.Fails}}},
			{"$set", bson.D{{"hide", false}}},
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

	err = s.notifyAll(p.Id)
	return
}

func (s *Service) parseDelete(r *http.Request) (p Problem, err error) {
	var b []byte
	b, err = ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(b, &p)
	if err != nil {
		return
	}

	return
}

func (s *Service) DeleteProblem(r *http.Request) (err error) {
	p, err := s.parseDelete(r)
	if err != nil {
		return
	}

	updates := p.Bson()
	_, err = s.problems.UpdateOne(
		context.TODO(),
		bson.M{"id": p.Id},
		updates,
	)

	err = s.notifyAll(p.Id)
	return nil
}

func (s *Service) notifyAll(id int) error {
	var data Problem
	err := s.problems.FindOne(
		context.TODO(),
		bson.M{"id": id},
	).Decode(&data)

	if err != nil {
		return err
	}

	data.LastAttempted = data.LastAttempted.Local()
	s.broadcast <- &data
	return nil
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
