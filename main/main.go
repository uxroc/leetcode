package main

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// gin framework configs
	s, err := NewService()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.LoadHTMLFiles("home.html")

	r.POST("/problem", func(c *gin.Context) {
		err := s.CreateProblem(c.Request)
		if err != nil {
			c.Error(err)
		} else {
			c.JSON(http.StatusOK, gin.H{})
		}
	})

	r.PUT("/problem", func(c *gin.Context) {
		err := s.DeleteProblem(c.Request)
		if err != nil {
			c.Error(err)
		} else {
			c.JSON(http.StatusOK, gin.H{})
		}
	})

	r.GET("/events", func(c *gin.Context) {
		sse := NewSSEClient(s)
		c.Stream(func(w io.Writer) bool {
			select {
			case p := <-sse.problemChan:
				c.SSEvent("message", *p)
			}
			return true
		})
	})

	r.GET("/problem", func(c *gin.Context) {
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
