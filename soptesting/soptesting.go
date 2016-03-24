package main

import (
	"log"
	"net/http"
)

func jsonHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		//https://www.w3.org/TR/cors
		//JSONP?
		w.Header().Set("Access-Control-Allow-Origin", "*")
		//w.Header().Set("Access-Control-Allow-Credentials", "*")
		//w.Header().Set("Access-Control-Expose-Headers", "*")
		w.Header().Set("Access-Control-Max-Age", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET")
		w.Header().Set("Access-Control-Allow-Headers", "complex")
	} else {
		value := r.FormValue("value")
		cors := r.FormValue("cors")

		if len(cors) > 0 {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{\"key\":\"" + value + "\"}"))
	}
}

func startFirst() {
	err := http.ListenAndServe(":9090", nil)

	if err != nil {
		log.Fatal("ListenAndServe (9090): ", err)
	}

}

func startSecond() {
	err := http.ListenAndServe(":9091", nil)

	if err != nil {
		log.Fatal("ListenAndServe (9091): ", err)
	}

}

func main() {
	// serve static from fileserver
	fs := http.FileServer(http.Dir("."))
	http.HandleFunc("/json", jsonHandler)
	http.Handle("/", fs)
	go startFirst()
	go startSecond()

	for {
		// wait forever
	}
}
