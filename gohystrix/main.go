package main

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/afex/hystrix-go/hystrix"
)

const commandName = "producer_api" // ชื่อของ circuit breaker

func main() {

	hystrix.ConfigureCommand(commandName, hystrix.CommandConfig{
		Timeout:                500,  // รอ command ทำงานนานแค่ไหน
		MaxConcurrentRequests:  100,  // จำนวน command ที่ให้ทำพร้อมกันได้
		ErrorPercentThreshold:  50,   // circuits จะเป็น open เมื่อจำนวน errors เกิน percent of requests
		RequestVolumeThreshold: 3,    // จำนวนน้อยสุดที่ต้องการก่อน circuit can be tripped due to health
		SleepWindow:            5000, // เวลาที่ open จะเปลี่ยนเป็น half-open
	})

	http.HandleFunc("/", logger(handle))
	log.Println("listening on :8080")
	http.ListenAndServe(":8080", nil)
}

func handle(w http.ResponseWriter, r *http.Request) {
	var response interface{}
	err := hystrix.Do(commandName, func() error {
		body, err := DoHTTPGet("http://www.google.com/robot.txt")
		if err != nil {
			return err
		}
		response = body.([]byte)
		return nil
	}, nil)

	if err != nil {
		log.Println(err)
		w.WriteHeader(503)
		w.Write([]byte("failed"))
		return
	}
	log.Println(string(response.([]byte)))
	w.WriteHeader(200)
	w.Write([]byte("success"))
}

func handleChan(w http.ResponseWriter, r *http.Request) {
	output := make(chan []byte, 1)
	errors := hystrix.Go(commandName, func() error {
		body, err := DoHTTPGet("http://")

		if err == nil {
			output <- body.([]byte)
		}
		return err
	}, nil)

	select {
	case out := <-output:
		log.Printf("success %s", out)
		w.WriteHeader(200)
		w.Write([]byte("success"))
	case err := <-errors:
		log.Printf("failed %s", err)
		w.WriteHeader(503)
		w.Write([]byte("failed"))
	}

}

func logger(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.URL.Path, r.Method)
		fn(w, r)
	}
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
