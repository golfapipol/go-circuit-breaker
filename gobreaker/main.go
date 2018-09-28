package main

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/sony/gobreaker"
)

const (
	requestLimit = 3
	failureRatio = 0.6
)

var breaker *gobreaker.CircuitBreaker

func init() {
	var settings gobreaker.Settings
	settings.Name = "HTTP GET"
	settings.ReadyToTrip = func(counts gobreaker.Counts) bool {
		log.Println("counts.TotalFailures", counts.TotalFailures, "counts.Requests", counts.Requests)
		failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
		return counts.Requests >= requestLimit && failureRatio >= failureRatio
	}

	breaker = gobreaker.NewCircuitBreaker(settings)
}

func Get(url string) ([]byte, error) {

	body, err := breaker.Execute(func() (interface{}, error) {
		return DoHTTPGet(url)
	})

	if err != nil {
		return nil, err
	}

	return body.([]byte), nil
}

func DoHTTPGet(url string) (interface{}, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func main() {
	_, err := Get("http://www.google.com/robots.txt")
	if err != nil {
		log.Fatal(err)
	}

	// fmt.Println(string(body))
}
