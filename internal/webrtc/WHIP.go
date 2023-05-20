package webrtc

import (
	"log"
	"strings"

	"github.com/pion/webrtc/v3"
)

func (api *API) NewWHIPConnection() (string, error) {
	peerConnection, err := api.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return "", nil
	}

	peerConnection.OnTrack(
		func(
			remoteTrack *webrtc.TrackRemote,
			rtpReceiver *webrtc.RTPReceiver,
		) {
			if strings.HasPrefix(
				remoteTrack.Codec().RTPCodecCapability.MimeType,
				"audio",
			) {
				// TODO: handle audio packets
			} else {
				// TODO: handle video packets
			}
		})

	peerConnection.OnICEConnectionStateChange(func(i webrtc.ICEConnectionState) {
		if i == webrtc.ICEConnectionStateFailed {
			if err := peerConnection.Close(); err != nil {
				log.Println(err)
			}
		}
		// end stream
	})
}
