package main

import (
	"log"
	"net/http"
)

// Adds CORS headers to the response to allow everything
func enableCors(res *http.ResponseWriter) {
	(*res).Header().Set("Access-Control-Allow-Origin", "*")
	(*res).Header().Set("Access-Control-Allow-Methods", "*")
	(*res).Header().Set("Access-Control-Allow-Headers", "*")
	(*res).Header().Set("Access-Control-Expose-Headers", "*")
}

func handleWHIP(res http.ResponseWriter, req *http.Request) {

}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleWHIP)

	log.Println("Starting server at port 8080")

	log.Fatal((&http.Server{
		Handler: mux,
		Addr:    ":8080",
	}).ListenAndServe())
}
