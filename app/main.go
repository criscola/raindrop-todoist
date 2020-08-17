package main

import (
	"context"
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/jackc/pgx/v4"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var conn *pgx.Conn

func main() {
	// Load configuration
	viper.SetConfigName("preferences")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	// Connect to DB
	// Try connecting to the db with 2 sec sleep between retries for a maximum of 10 times
	for i := 0; i < 10; i++ {
		conn, err = pgx.Connect(context.Background(), os.Getenv("POSTGRES_URL"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
			fmt.Println("Retry num: ", i+1)
			time.Sleep(time.Second * 2)
		} else {
			fmt.Println("DB successfully connected.")
			break
		}
	}
	defer conn.Close(context.Background())

	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://api.raindrop.io/rest/v1/raindrops/0", nil)
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error in building request: %s \n", err))
	}
	// Raindrop API doesn't rigorously respect RFC 3986 so commas shouldn't be encoded, Go will always encode them so
	// the URL encoding must be handled specifically for this case
	req.URL.Opaque = `/rest/v1/raindrops/0?search=%5B%7B%22key%22:%22tag%22,%22val%22:%22` +
		viper.GetString("LABEL_NAME") + `%22%7D%5D`

	fmt.Println(viper.GetString("RAINDROP_TOKEN"))
	req.Header.Add("Authorization", "Bearer " + os.Getenv("RAINDROP_TOKEN"))

	res, err := client.Do(req)
	if err != nil {
		panic(fmt.Errorf("Fatal error sending request: %s \n", err))
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	// Connect to Todoist

	// For each id obtained from the raindrop request
		// Match id against records in db
			// If id exists then move on
			// else add task to todoist and then add record
	// For each id obtained from db
		// Get todoist task with that id
		// If attribute "checked" is true then remove label from raindrop and then remove entry from db
	// Wait 5 seconds

	test, _ := jsonparser.GetBoolean(body, "result")
	fmt.Println("test is: ", test)

	// For each bookmark marked as "Read it later"
	_, err = jsonparser.ArrayEach(body, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		bookmarkId, _ := jsonparser.GetInt(value, "_id")

		// Check if bookmark is already present in db
		var todoId int64
		err = conn.QueryRow(context.Background(),"SELECT id_todo FROM bookmark_with_todo WHERE id_bookmark=$1", strconv.FormatInt(bookmarkId, 10)).Scan(&todoId)
		fmt.Println("RESULT FROM QUERY: ", err)
		if err == pgx.ErrNoRows {
			// If bookmark doesn't exist, create the relative Todoist task and add IDs to the db
			domain, _ := jsonparser.GetString(value, "domain")
			url, _ := jsonparser.GetString(value, "url")
			title, _ := jsonparser.GetString(value, "title")

			// Add task to todoist and save the id
			// TODO: Cut strings too long
			postBody := strings.NewReader(`{"content": "Read [` + title  + `](` + url + `) on `+ domain + `"}`)
			req, err = http.NewRequest("POST", "https://api.todoist.com/rest/v1/tasks", postBody)

			if err != nil {
				panic(fmt.Errorf("Fatal error in building request: %s \n", err))
			}
			req.Header.Add("Content-Type", "application/json")
			req.Header.Add("Authorization", "Bearer " + os.Getenv("TODOIST_TOKEN"))

			res, err = client.Do(req)
			if err != nil {
				panic(fmt.Errorf("Fatal error sending request: %s \n", err))
			}

			taskCreatedRes, _ := ioutil.ReadAll(res.Body)
			// Get task id from the response
			todoId, _ = jsonparser.GetInt(taskCreatedRes, "id")
			defer res.Body.Close()

			// Now add elements to db
			_, err := conn.Exec(context.Background(), "INSERT INTO bookmark_with_todo (id_bookmark, id_todo) VALUES ($1, $2)",
				bookmarkId, todoId)
			if err != nil {
				fmt.Println("Error inserting entry into bookmark_with_todo: ", err)
			}
		} else {
			if err != nil {
				fmt.Println("Error querying the database for id_bookmark entries: ", err)
			}
		}
	}, "items")

	// Now check if the user has checked a "Read it later" task

	if err != nil {
		fmt.Println("Error is: ", err)
	}
}