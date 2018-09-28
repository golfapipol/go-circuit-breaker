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

func main() {
	settings := gobreaker.Settings{
		Name: "HTTP GET", // ชื่อของ circuit breaker
		//MaxRequests: 0, จำนวน request ที่ให้ผ่านได้เวลา half-open
		//Interval: 0, // เวลาที่จะ reset count ตอนเป็น close
		Timeout:     5 * time.Second,                // เวลาที่ open จะเปลี่ยนเป็น half-open
		ReadyToTrip: ConsecutiveFailureReachedLimit, // สามารถ custom ได้ว่าเงื่อนไขของการ open คืออะไร (default เป็น failed ติดกันมากกว่า 5 ครั้ง)
		//OnStateChange: func(name string, from State, to State)// ไว้ใช้อาจจะส่งเมล์หรือ alert เวลา state ถูกเปลี่ยน
	}

	breaker = gobreaker.NewCircuitBreaker(settings)

	http.HandleFunc("/", logger(handle))
	log.Println("listening on :8080")
	http.ListenAndServe(":8080", nil)
}

func handle(w http.ResponseWriter, r *http.Request) {
	body, err := Get("http://localhost:3000/v1/quizzes")
	if err != nil {
		log.Println(err)
		w.WriteHeader(503)
		w.Write([]byte("failed"))
	} else {
		log.Println("success ", string(body))
		w.WriteHeader(200)
		w.Write([]byte("success"))
	}
}

func logger(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.URL.Path, r.Method)
		fn(w, r)
	}
}

func ConsecutiveFailureReachedLimit(counts gobreaker.Counts) bool {
	return counts.ConsecutiveFailures > requestLimit
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
