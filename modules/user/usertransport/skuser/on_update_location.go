package skuser

import (
	"demo/common"
	"demo/component"
	socketio "github.com/googollee/go-socket.io"
	"log"
)

//func(s socketio.Conn, token string) {}

type LocationData struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

func OnUserUpdateLocation(appCtx component.AppContext, requester common.Requester) func(s socketio.Conn, location LocationData) {
	return func(s socketio.Conn, location LocationData) {

		// location belong to user ???
		log.Println("User update location: user id is", requester.GetUserId(), "at location", location)
	}
}
