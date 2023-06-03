package whip

import (
	"log"
	"os"
	"strconv"

	"github.com/coder-with-a-bushido/neralai/internal/utils"
	"github.com/pion/ice/v2"
	"github.com/pion/interceptor"
	"github.com/pion/interceptor/pkg/intervalpli"
	"github.com/pion/webrtc/v3"
)

var api *webrtc.API

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
	// Set IP to be used in candidates instead of using ICE servers
	// ref: https://pkg.go.dev/github.com/pion/webrtc/v3@v3.1.60#SettingEngine.SetNAT1To1IPs
	settingEngine := webrtc.SettingEngine{}
	populateSettingEngine(&settingEngine)

	// Create the API object with configured mediaEngine, interceptorRegistry, settingEngine
	api = webrtc.NewAPI(
		webrtc.WithMediaEngine(mediaEngine),
		webrtc.WithInterceptorRegistry(interceptorRegistry),
		webrtc.WithSettingEngine(settingEngine),
	)

	resourceMap = make(map[string]*Resource)
}

func CleanUp() {
	resourceMapLock.Lock()
	defer resourceMapLock.Unlock()

	// Delete all WHIP resources
	for _, resource := range resourceMap {
		resource.Disconnect <- struct{}{}
	}
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
		PayloadType: 96,
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
		PayloadType: 111,
	}, webrtc.RTPCodecTypeAudio); err != nil {
		return err
	}

	return nil
}

// credits: https://github.com/Glimesh/broadcast-box
func populateSettingEngine(settingEngine *webrtc.SettingEngine) {
	NAT1To1IPs := []string{}

	NAT1To1IPs = append(NAT1To1IPs, utils.GetPublicIP())

	if os.Getenv("NAT_1_TO_1_IP") != "" {
		NAT1To1IPs = append(NAT1To1IPs, os.Getenv("NAT_1_TO_1_IP"))
	}

	if len(NAT1To1IPs) != 0 {
		settingEngine.SetNAT1To1IPs(NAT1To1IPs, webrtc.ICECandidateTypeHost)
	}

	if os.Getenv("UDP_MUX_PORT") != "" {
		udpPort, err := strconv.Atoi(os.Getenv("UDP_MUX_PORT"))
		if err != nil {
			log.Fatal(err)
		}

		udpMux, err := ice.NewMultiUDPMuxFromPort(udpPort)
		if err != nil {
			log.Fatal(err)
		}

		settingEngine.SetICEUDPMux(udpMux)
	}
}
