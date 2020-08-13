package main

import (
	"fmt"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
)

func main() {
	// Load configuration
	viper.SetConfigName("secret")
	viper.SetConfigName("preferences")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://api.raindrop.io/rest/v1/raindrops/0", nil)
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error in building request: %s \n", err))
	}
	// Raindrop API doesn't rigorously respect RFC 3986 so commas shouldn't be encoded, Go will always encode them so
	// the URL encoding must be handled specifically for this case
	req.URL.Opaque = `/rest/v1/raindrops/0?search=%5B%7B%22key%22:%22tag%22,%22val%22:%22` +
		viper.GetString("LABEL_NAME") + `%22%7D%5D`

	req.Header.Add("Authorization", "Bearer 6da4f2c4-f5bb-4626-9fb2-733e743fac1a")

	res, err := client.Do(req)
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error sending request: %s \n", err))
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	fmt.Println(string(body))
}