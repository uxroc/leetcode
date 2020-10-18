package main

import (
	"encoding/json"
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
}

func (p Problem) String() (string, error) {
	str, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	return string(str), nil
}


