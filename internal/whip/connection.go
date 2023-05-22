package whip

import (
	"context"
	// "io"
	"log"

	// "github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
)

func NewWHIPConnection(ctx context.Context, offerSDP string, disconnect chan<- struct{},

// audioPackets chan *rtp.Packet, videoPackets chan *rtp.Packet
) (answerSDP string, resourceId string, err error) {
	peerConnection, err := api.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return "", "", nil
	}

	// make answer SDP use "recvonly" attribute
	if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo, webrtc.RTPTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionRecvonly,
	}); err != nil {
		return "", "", nil
	}
	if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio, webrtc.RTPTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionRecvonly,
	}); err != nil {
		return "", "", nil
	}

	peerConnection.OnTrack(
		func(
			remoteTrack *webrtc.TrackRemote,
			rtpReceiver *webrtc.RTPReceiver,
		) {
			// Read RTP packets of the track
			// for {
			// 	rtpPacket, _, readErr := remoteTrack.ReadRTP()
			// 	if readErr != nil {
			// 		if readErr == io.EOF {
			// 			return
			// 		}
			// 		panic(readErr)
			// 	}
			// 	switch remoteTrack.Kind() {
			// 	// If it's audio, send to audioPackets
			// 	case webrtc.RTPCodecTypeAudio:
			// 		audioPackets <- rtpPacket
			// 	// If it's video, send to videoPackets
			// 	case webrtc.RTPCodecTypeVideo:
			// 		videoPackets <- rtpPacket
			// 	}
			// }
		})

	peerConnection.OnICEConnectionStateChange(func(i webrtc.ICEConnectionState) {
		// close connection
		if i == webrtc.ICEConnectionStateFailed {
			if err := peerConnection.Close(); err != nil {
				log.Println(err)
			}

			disconnect <- struct{}{}
		}
	})

	if err := peerConnection.SetRemoteDescription(
		webrtc.SessionDescription{
			Type: webrtc.SDPTypeOffer,
			SDP:  offerSDP,
		},
	); err != nil {
		return "", "", nil
	}

	iceGatherComplete := webrtc.GatheringCompletePromise(peerConnection)
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		return "", "", err
	}
	if err = peerConnection.SetLocalDescription(answer); err != nil {
		return "", "", err
	}

	<-iceGatherComplete
	resourceId = addNewResource(
		&Resource{
			peerConnection: peerConnection,
			ctx:            ctx,
			Disconnect:     disconnect,
		},
	)
	return peerConnection.LocalDescription().SDP, resourceId, nil
}
