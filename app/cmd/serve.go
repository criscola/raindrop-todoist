/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

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

	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("serve called")
	},
}

var conn *pgx.Conn

func init() {
	rootCmd.AddCommand(serveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	// Connect to DB
	// Try connecting to the db with 2 sec sleep between retries for a maximum of 10 times

	for i := 0; i < 10; i++ {
		temp, err := pgx.Connect(context.Background(), os.Getenv("POSTGRES_URL"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
			fmt.Println("Retry num: ", i+1)
			time.Sleep(time.Second * 2)
		} else {
			fmt.Println("DB successfully connected.")
			conn = temp
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