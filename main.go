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

	"coder-with-a-bushido.in/neralai/internal/hls"
	"coder-with-a-bushido.in/neralai/internal/whip"
	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
)

func main() {
	whip.Init()
	hls.Init()

	r := chi.NewRouter()
	// Adds CORS headers
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"https://*", "http://*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders: []string{"*"},
	}))

	// WHIP endpoint - Start new stream
	r.Post("/stream", startNewStream)
	// Reserved WHIP endpoint methods for future
	// ref: `draft-ietf-wish-whip-01`
	r.Get("/stream", r.MethodNotAllowedHandler())
	r.Head("/stream", r.MethodNotAllowedHandler())
	r.Put("/stream", r.MethodNotAllowedHandler())

	// WHIP resources endpoint - operations on existing stream
	r.Delete("/stream/{resourceId}", stopStream)
	// TODO: WHIP resources endpoint - Trickle ICE
	r.Patch("/stream/{resourceId}", r.MethodNotAllowedHandler())
	// Reserved WHIP resources endpoint methods for future
	// ref: `draft-ietf-wish-whip-01`
	r.Get("/stream/{resourceId}", r.MethodNotAllowedHandler())
	r.Head("/stream/{resourceId}", r.MethodNotAllowedHandler())
	r.Post("/stream/{resourceId}", r.MethodNotAllowedHandler())
	r.Put("/stream/{resourceId}", r.MethodNotAllowedHandler())

	// Serve HLS files for resources from the output directory
	r.Get("/stream/{resourceId}/hls/*", serveHLSFiles)

	log.Println("Starting server at port 8080")
	log.Fatal((&http.Server{
		Handler: r,
		Addr:    ":8080",
	}).ListenAndServe())
	defer cleanup()

	// Cleanup on exit
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
	hls.CleanUp()
}

func startNewStream(w http.ResponseWriter, r *http.Request) {
	offerSDP, err := io.ReadAll(r.Body)
	if err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	disconnect := make(chan struct{})

	answerSDP, resourceID, err := whip.NewConnection(ctx, string(offerSDP), disconnect)
	if err != nil {
		writeInternalServerError(w, err.Error())
		cancel()
		return
	}

	w.Header().Set("Location", string("http://"+r.Host+"/stream/"+resourceID))
	w.Header().Set("Content-Type", "application/sdp")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, answerSDP)

	hls.NewStreamFromWHIPResource(ctx, resourceID)

	go func() {
		<-disconnect
		cancel()
	}()
}

func stopStream(w http.ResponseWriter, r *http.Request) {
	resourceId := chi.URLParam(r, "resourceId")
	resourse := whip.GetResource(resourceId)
	if resourse == nil {
		writeBadRequest(w, "Invalid Resource ID")
		return
	}

	resourse.Disconnect <- struct{}{}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w)
}

func serveHLSFiles(w http.ResponseWriter, r *http.Request) {
	resourceId := chi.URLParam(r, "resourceId")
	resourse := whip.GetResource(resourceId)
	if resourse == nil {
		writeBadRequest(w, "Invalid Resource ID")
		return
	}

	filePath := "output" + strings.TrimPrefix(r.URL.Path, "/stream")
	http.ServeFile(w, r, filePath)
}

func writeBadRequest(w http.ResponseWriter, errorStr string) {
	writeHTTPError(w, errorStr, http.StatusBadRequest)
}
func writeInternalServerError(w http.ResponseWriter, errorStr string) {
	writeHTTPError(w, errorStr, http.StatusInternalServerError)
}
func writeHTTPError(w http.ResponseWriter, err string, code int) {
	log.Println(err)
	http.Error(w, err, code)
}
