package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	"coder-with-a-bushido.in/neralai/internal/whip"
)

var resourcesPath = "/resources/"

// Adds CORS headers to the response to allow everything
func enableCors(res *http.ResponseWriter) {
	(*res).Header().Set("Access-Control-Allow-Origin", "*")
	(*res).Header().Set("Access-Control-Allow-Methods", "*")
	(*res).Header().Set("Access-Control-Allow-Headers", "*")
	(*res).Header().Set("Access-Control-Expose-Headers", "*")
}

func logHTTPError(w http.ResponseWriter, err string, code int) {
	log.Println(err)
	http.Error(w, err, code)
}

func handleWHIPConn(res http.ResponseWriter, req *http.Request) {
	enableCors(&res)
	//TODO: authentication with bearer token

	// Reserve other methods for future use according to `draft-ietf-wish-whip-01`
	if req.Method == http.MethodGet || req.Method == http.MethodHead || req.Method == http.MethodPut {
		logHTTPError(res, "Unsupported Method", http.StatusMethodNotAllowed)
		return
	}

	offerSDP, err := io.ReadAll(req.Body)
	if err != nil {
		logHTTPError(res, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	disconnect := make(chan struct{})

	answerSDP, resourceID, err := whip.NewWHIPConnection(ctx, string(offerSDP), disconnect)
	if err != nil {
		logHTTPError(res, err.Error(), http.StatusBadRequest)
		cancel()
		return
	}

	res.WriteHeader(http.StatusCreated)
	res.Header().Add("Location", req.Host+resourcesPath+resourceID)
	res.Header().Add("Content-Type", "application/sdp")
	fmt.Fprint(res, answerSDP)

	select {
	case <-disconnect:
		cancel()
	}
}

func handleWHIPClose(res http.ResponseWriter, req *http.Request) {
	enableCors(&res)
	// TODO: authentication with bearer token

	// Reserve other methods for future use according to `draft-ietf-wish-whip-01`
	// Trickle ICE(PATCH method) not supported
	if req.Method == http.MethodGet || req.Method == http.MethodHead || req.Method == http.MethodPost || req.Method == http.MethodPut || req.Method == http.MethodPatch {
		logHTTPError(res, "Unsupported Method", http.StatusMethodNotAllowed)
		return
	}

	resourse := whip.GetResource(req.URL.Path)
	resourse.Disconnect <- struct{}{}

	res.WriteHeader(http.StatusOK)
	fmt.Fprint(res)
}

func main() {
	whip.Init()
	mux := http.NewServeMux()
	// for creating a new resource
	mux.HandleFunc("/", handleWHIPConn)
	// for closing an existing resource
	mux.Handle(
		resourcesPath,
		http.StripPrefix(resourcesPath, http.HandlerFunc(handleWHIPClose)),
	)

	log.Println("Starting server at port 8080")

	log.Fatal((&http.Server{
		Handler: mux,
		Addr:    ":8080",
	}).ListenAndServe())
}
