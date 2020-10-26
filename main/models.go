package main

import (
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

type Problem struct {
	Id 				int 		`json:"Id"`
	Title 			string 		`json:"Title"`
	Uname			string		`json:"Uname"`
	Difficulty 		string 		`json:"Difficulty"`
	LastAttempted  	time.Time 	`json:"LastAttempted"`
	Attempts 		int 		`json:"Attempts"`
	Tags 			[]string	`json:"Tags"`
	Url				string		`json:"Url"`
	Hide   			bool 		`json:"Hide"`
}

func (p Problem) Bson() bson.D {
	var setElements bson.D
	if len(p.Tags) > 0 {
		setElements = append(setElements, bson.E{Key: "tags", Value: p.Tags})
	}
	if p.Hide {
		setElements = append(setElements, bson.E{Key: "hide", Value: p.Hide})
	}
	return bson.D{{"$set", setElements}}
}

func (p Problem) String() (string, error) {
	str, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	return string(str), nil
}


