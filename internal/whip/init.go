package whip

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/pion/interceptor"
	"github.com/pion/interceptor/pkg/intervalpli"
	"github.com/pion/webrtc/v3"
)

var api *webrtc.API

func getPublicIP() string {
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

func populateInterceptorRegistry(interceptorRegistry *interceptor.Registry) error {
	intervalPliFactory, err := intervalpli.NewReceiverInterceptor()
	if err != nil {
		return err
	}
	interceptorRegistry.Add(intervalPliFactory)
	return nil
}

func populateMediaEngine(mediaEngine *webrtc.MediaEngine) error {
	if err := mediaEngine.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType:     webrtc.MimeTypeVP8,
			ClockRate:    90000,
			Channels:     0,
			SDPFmtpLine:  "",
			RTCPFeedback: nil,
		},
	}, webrtc.RTPCodecTypeVideo); err != nil {
		return err
	}
	if err := mediaEngine.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType:     webrtc.MimeTypeOpus,
			ClockRate:    48000,
			Channels:     0,
			SDPFmtpLine:  "",
			RTCPFeedback: nil,
		},
	}, webrtc.RTPCodecTypeAudio); err != nil {
		return err
	}

	return nil
}

func populateSettingEngine(settingEngine *webrtc.SettingEngine) {
	NAT1To1IPs := []string{}

	if os.Getenv("INCLUDE_PUBLIC_IP_IN_NAT_1_TO_1_IP") != "" {
		NAT1To1IPs = append(NAT1To1IPs, getPublicIP())
	}

	if os.Getenv("NAT_1_TO_1_IP") != "" {
		NAT1To1IPs = append(NAT1To1IPs, os.Getenv("NAT_1_TO_1_IP"))
	}

	if len(NAT1To1IPs) != 0 {
		settingEngine.SetNAT1To1IPs(NAT1To1IPs, webrtc.ICECandidateTypeHost)
	}
}

func Init() {
	// Configure mediaEngine with support for VP8 and Opus
	mediaEngine := &webrtc.MediaEngine{}
	if err := populateMediaEngine(mediaEngine); err != nil {
		panic(err)
	}

	// Configure interceptorRegistry to send PLI every 3 seconds
	interceptorRegistry := &interceptor.Registry{}
	if err := populateInterceptorRegistry(interceptorRegistry); err != nil {
		panic(err)
	}

	// Set up interceptorRegistry and register it with mediaEngine
	if err := webrtc.RegisterDefaultInterceptors(mediaEngine, interceptorRegistry); err != nil {
		log.Fatal(err)
	}

	// Configure settingEngine to use public IP and avoid STUN server
	// refer: https://pkg.go.dev/github.com/pion/webrtc/v3@v3.1.60#SettingEngine.SetNAT1To1IPs
	settingEngine := webrtc.SettingEngine{}
	populateSettingEngine(&settingEngine)

	// Create the API object with configured mediaEngine, interceptorRegistry, settingEngine
	api = webrtc.NewAPI(
		webrtc.WithMediaEngine(mediaEngine),
		webrtc.WithInterceptorRegistry(interceptorRegistry),
		webrtc.WithSettingEngine(settingEngine),
	)
}
