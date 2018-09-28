package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/sony/gobreaker"
)

const (
	requestLimit = 3
	failureRatio = 0.6
)

var breaker *gobreaker.CircuitBreaker

func ConsecutiveFailureReachedLimit(counts gobreaker.Counts) bool {
	return counts.ConsecutiveFailures > requestLimit
}
func init() {
	settings := gobreaker.Settings{
		Name: "HTTP GET",
		Timeout: 5 * time.Second,
		ReadyToTrip: ConsecutiveFailureReachedLimit
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
	for {
		select {
		case t := <-time.Tick(time.Second):
			if t.Unix()-time.Now().Unix() > 60 {
				return
			}

			body, err := Get("http://localhost:3000/v1/quizzes")
			if err != nil {
				log.Println(err)
			} else {
				log.Println("success ", string(body))
			}
		}
	}
}
