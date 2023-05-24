package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"coder-with-a-bushido.in/neralai/internal/outputs"
	"coder-with-a-bushido.in/neralai/internal/whip"
)

var resourcePath = "/resources/"

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
	log.Println("Request for new WHIP conn")
	enableCors(&res)
	//TODO: authentication with bearer token

	switch req.Method {
	case http.MethodPost:
		break
	// preflight request
	case http.MethodOptions:
		res.WriteHeader(http.StatusOK)
		fmt.Fprint(res)
		return
	// Reserve other methods for future use according to `draft-ietf-wish-whip-01`
	default:
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

	disconnect := make(chan struct{})

	answerSDP, resourceID, err := whip.NewConnection(ctx, string(offerSDP), disconnect)
	if err != nil {
		logHTTPError(res, err.Error(), http.StatusInternalServerError)
		cancel()
		return
	}

	res.Header().Set("Location", string("http://"+req.Host+resourcePath+resourceID))
	res.Header().Set("Content-Type", "application/sdp")
	res.WriteHeader(http.StatusCreated)
	fmt.Fprint(res, answerSDP)

	outputOptions := outputs.Options{Recording: false, HLSStream: true}
	outputs.StartFromWHIPResource(ctx, resourceID, outputOptions)

	go func() {
		<-disconnect
		cancel()
	}()
}

func handleWHIPResource(res http.ResponseWriter, req *http.Request) {
	enableCors(&res)
	// TODO: authentication with bearer token

	switch req.Method {
	case http.MethodDelete:
		break
	// TODO: Trickle ICE(PATCH method) not supported
	case http.MethodPatch:
		logHTTPError(res, "Unsupported Method", http.StatusMethodNotAllowed)
		return
	// preflight request
	case http.MethodOptions:
		res.WriteHeader(http.StatusOK)
		fmt.Fprint(res)
		return
	// Reserve other methods for future use according to `draft-ietf-wish-whip-01`
	default:
		logHTTPError(res, "Unsupported Method", http.StatusMethodNotAllowed)
		return
	}

	resourceId := strings.TrimPrefix(req.URL.Path, resourcePath)
	resourse := whip.GetResource(resourceId)
	if resourse != nil {
		resourse.Disconnect <- struct{}{}
	}

	res.WriteHeader(http.StatusOK)
	fmt.Fprint(res)
}

func main() {
	whip.Init()
	outputs.Init()
	mux := http.NewServeMux()
	// for creating a new WHIP resource
	mux.HandleFunc("/start", handleWHIPConn)
	// for operating on an existing WHIP resource
	mux.HandleFunc(resourcePath, handleWHIPResource)

	log.Println("Starting server at port 8080")

	log.Fatal((&http.Server{
		Handler: mux,
		Addr:    ":8080",
	}).ListenAndServe())
	defer cleanup()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		cleanup()
		os.Exit(0)
	}()
}

func cleanup() {
	whip.CleanUp()
	outputs.CleanUp()
}
