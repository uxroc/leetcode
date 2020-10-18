package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	// gin framework configs
	s, err := NewService()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	r := gin.Default()
	r.LoadHTMLFiles("home.html")

	r.POST("/attempt", func(c *gin.Context) {
		err := s.ServeAttempt(c.Request)
		if err != nil {
			c.Error(err)
		} else {
			c.JSON(http.StatusOK, gin.H{})
		}
	})

	r.GET("/data", func(c *gin.Context) {
		data, err := s.GetData()
		if err != nil {
			c.Error(err)
		} else {
			c.JSON(http.StatusOK, data)
		}
	})

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "home.html", nil)
	})

	r.Run()
}