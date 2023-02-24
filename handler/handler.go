package handler

import (
	"context"
	"sync"

	"github.com/dev-saw99/deko/compiler"
	pb "github.com/dev-saw99/deko/proto"
	"github.com/dev-saw99/deko/utils"
	constants "github.com/dev-saw99/deko/utils/constant"
	"github.com/gin-gonic/gin"
)

var CompilerClient *compiler.Compiler
var ExecutingMeassageMAP sync.Map

func CodeCompiler(ctx *gin.Context) {
	connectionID := ctx.Param("connid")
	utils.Logger.Infow("CodeCompiler Handler",
		"connection-id", connectionID,
	)

	// initiating websocket connection with the client
	utils.Logger.Infow("Upgrading to websocket connection",
		"connection-id", connectionID,
	)
	wsClient := UpgradeToWSClient(ctx)

	var wg sync.WaitGroup
	inputMessageChannel := make(chan *pb.WSInputInterface)
	// making buffer channel, so that other threads don't get blocked on output channel
	outputMessageChannel := make(chan *pb.WSOutputInterface, 5)
	processMessageChannel := make(chan *pb.WSInputInterface)
	ct, cancel := context.WithCancel(context.Background())
	wg.Add(1)
	go readMessage(ct, cancel, wsClient, &wg, inputMessageChannel)
	wg.Add(1)
	go processMessage(ct, cancel, connectionID, wsClient, &wg, processMessageChannel, outputMessageChannel)

loop:
	for {
		select {
		case input := <-inputMessageChannel:
			// fmt.Println("INPUT", input)
			processMessageChannel <- input
		case output := <-outputMessageChannel:
			// fmt.Println("OUTPUT", output)
			wsClient.SendMessage(output)
			if isCompileCompleted(int(output.StatusCode)) {
				wsClient.Close()
				cancel()
				break loop
			}
		case <-ct.Done():
			utils.Logger.Infow("Connection Closed",
				"connection-id", connectionID)
			break loop

		}
	}
	wg.Wait()
}

func isCompileCompleted(statusCode int) bool {

	switch statusCode {
	case constants.STATUS_COMPILE_DONE:
		return true
	case constants.STATUS_CODE_INVALID:
		return true
	case constants.STATUS_LANGUAGE_INVALID:
		return true
	case constants.STATUS_TIMEOUT:
		return true
	case constants.STATUS_CONN_CLOSE:
		return true
	case constants.STATUS_DEFAULT_ERROR:
		return true
	case constants.STATUS_INVALID_MESSAGE:
		return true
	default:
		return false
	}

}

func processMessage(ctx context.Context, cancel context.CancelFunc, connectionID string, wsClient *WSClient, wg *sync.WaitGroup, processMessageChannel <-chan *pb.WSInputInterface, outputMessageChannel chan<- *pb.WSOutputInterface) {
	defer wg.Done()

	for {

		msg := <-processMessageChannel
		msg.MessageId = connectionID
		// check if message is to close connection
		if msg.Message == constants.MSG_CLOSE_CONNECTION {
			replyAndCloseConnection(wsClient, msg.MessageId)
			cancel()
			break
		}
		if _, ok := ExecutingMeassageMAP.Load(msg.MessageId); ok {
			outputMessageChannel <- &pb.WSOutputInterface{
				MessageId:  msg.MessageId,
				Message:    "Already Processing Code",
				StatusCode: constants.STATUS_ALREADY_PROCESSING,
			}
			continue
		}
		// Validate message for compilation
		if msg.Code == "" {
			outputMessageChannel <- &pb.WSOutputInterface{
				MessageId:  msg.MessageId,
				Message:    "Invalid Code",
				StatusCode: constants.STATUS_CODE_INVALID,
			}
			continue
		} else if msg.Language == "" {
			outputMessageChannel <- &pb.WSOutputInterface{
				MessageId:  msg.MessageId,
				Message:    "Invalid Language",
				StatusCode: constants.STATUS_LANGUAGE_INVALID,
			}
			continue
		} else if msg.Message != constants.MSG_COMPILE_CODE {
			outputMessageChannel <- &pb.WSOutputInterface{
				MessageId:  msg.MessageId,
				Message:    "Invalid Message",
				StatusCode: constants.STATUS_INVALID_MESSAGE,
			}
			continue
		}

		// TODO: Add logic to send this message to compiler
		ExecutingMeassageMAP.Store(msg.MessageId, true)
		go CompilerClient.CompileCode(msg, outputMessageChannel, ctx)
		// TODO: Add logic to recieve message from the compiler
	}
	ExecutingMeassageMAP.Delete(connectionID)
}

func readMessage(ctx context.Context, cancel context.CancelFunc, wsClient *WSClient, wg *sync.WaitGroup, inputMessageChannel chan<- *pb.WSInputInterface) {
	defer wg.Done()
	defer cancel()
	for wsClient.ConnectionLive {
		select {
		case <-ctx.Done():
			return
		default:
			msg := wsClient.GetMessage()
			if msg != nil {
				inputMessageChannel <- msg
			}
		}
	}

}

func replyAndCloseConnection(ws *WSClient, msgID string) {
	ws.SendMessage(&pb.WSOutputInterface{
		MessageId:  msgID,
		Message:    "Connection Closed",
		StatusCode: constants.STATUS_CONN_CLOSE,
	})
	ws.Close()
}
