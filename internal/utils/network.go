package utils

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
)

func GetPublicIP() string {
	req, err := http.Get("http://ip-api.com/json/")
	if err != nil {
		log.Fatal(err)
	}
	defer req.Body.Close()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		log.Fatal(err)
	}

	ip := struct {
		Query string
	}{}
	if err = json.Unmarshal(body, &ip); err != nil {
		log.Fatal(err)
	}

	if ip.Query == "" {
		log.Fatal("Query entry was not populated")
	}

	return ip.Query
}

// find a even-numbered free port
// refer: https://www.ietf.org/rfc/rfc2327.txt#:~:text=For%20RTP%20compliance%20it%20should%20be%20an%20even%0A%20%20%20%20%20number.
func GetFreePortEven() (port int, err error) {
	for port == 0 || port%2 == 1 {
		var addr *net.TCPAddr
		if addr, err = net.ResolveTCPAddr("tcp", "127.0.0.1:"); err != nil {
			return
		}
		var l *net.TCPListener
		if l, err = net.ListenTCP("tcp", addr); err != nil {
			return
		} else if err = l.Close(); err != nil {
			return
		}
		port = l.Addr().(*net.TCPAddr).Port
	}
	return
}
