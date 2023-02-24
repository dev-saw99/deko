package handler

import (
	"encoding/json"
	"net/http"

	pb "github.com/dev-saw99/deko/proto"
	"github.com/dev-saw99/deko/utils"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WSClient struct {
	C              *websocket.Conn
	ConnectionLive bool
	ConnectionUUID string
	upgrader       *websocket.Upgrader
}

func (ws *WSClient) Close() {
	ws.ConnectionLive = false
	ws.C.Close()
}

func (ws *WSClient) SendMessage(msg *pb.WSOutputInterface) {
	if !ws.ConnectionLive {
		utils.Logger.Errorw("Attempt to write into a closed websocket connection",
			"message_id", msg.MessageId,
			"message", msg)
	}
	err := ws.C.WriteJSON(msg)
	if err != nil {
		utils.Logger.Errorw("Error occurred while writing into the websocket connection",
			"message_id", msg.MessageId,
			"error", err)
	}
}

func (ws *WSClient) GetMessage() *pb.WSInputInterface {
	t, msg, err := ws.C.ReadMessage()
	if err != nil {
		if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseInternalServerErr, websocket.CloseNoStatusReceived) {
			ws.ConnectionLive = false
			utils.Logger.Infow("Connection closed from the client side",
				"connection-id", ws.ConnectionUUID,
			)
			return nil
		}
		utils.Logger.Errorw("Error occured while reading message from websocket connection",
			"connection-id", ws.ConnectionUUID,
			"error", err.Error())
		return nil
	}

	var inputMessage pb.WSInputInterface
	err = json.Unmarshal(msg, &inputMessage)
	if err != nil {
		utils.Logger.Errorw("Error occured while unmarshal of websocket message",
			"connection-id", ws.ConnectionUUID,
			"error", err.Error())
		return nil
	}
	inputMessage.MessageType = int64(t)
	return &inputMessage
}

func UpgradeToWSClient(ctx *gin.Context) *WSClient {
	var client WSClient
	client.ConnectionUUID = ctx.Param("connid")
	client.upgrader = &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	c, err := client.upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		utils.Logger.Errorw("Error occured while upgrading the http connection",
			"connection-id", client.ConnectionUUID,
			"err", err.Error(),
		)
		return nil
	}
	client.C = c
	client.ConnectionLive = true
	return &client
}
