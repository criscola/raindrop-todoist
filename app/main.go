package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
)

func main() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	client := &http.Client{}
	
	req, _ := http.NewRequest("GET", "https://api.raindrop.io/rest/v1/collections", nil)
	req.Header.Set("Authorization", "Bearer aabc506e-5688-496e-80c2-fe8f6f185f2b")
	res, _ := client.Do(req)

	data, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()

	fmt.Printf("%s\n", data)
	r.Run() // listen and serve on 0.0.0.0:8080
}