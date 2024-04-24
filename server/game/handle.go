package game

import (
	"net"

	"github.com/pemmel/gameserver/server"
	"google.golang.org/protobuf/proto"
)

type handleT struct {
	session     *server.Session
	addr        *net.UDPAddr
	requestCode uint8
	payload     []byte
}

func handle(h *handleT) {
	switch h.requestCode {
	case RequestCode_SyncPos:
		proto.Unmarshal(h.payload, nil)

	case RequestCode_Logout:
		proto.Unmarshal(h.payload, nil)

	case RequestCode_CreateLobby:
		proto.Unmarshal(h.payload, nil)

	default:
		break
	}
}
