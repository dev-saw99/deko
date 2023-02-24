package compiler

import (
	"context"
	"fmt"
	"io"
	"sync"

	pb "github.com/dev-saw99/deko/proto"
	"github.com/dev-saw99/deko/utils"
	constants "github.com/dev-saw99/deko/utils/constant"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Compiler struct {
	DNS    string
	Client pb.CompileServiceClient
}

func (c *Compiler) Init() {
	utils.Logger.Infow("Connecting with DekoBridge Service",
		"port", c.DNS)
	conn, err := grpc.Dial(c.DNS, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		utils.Logger.Fatalw("Can't connect with server",
			"error", err)
	}
	c.Client = pb.NewCompileServiceClient(conn)
}

func NewCompiler(dns string) *Compiler {
	return &Compiler{
		DNS: dns,
	}
}

func (c *Compiler) CompileCode(input *pb.WSInputInterface, outputChan chan<- *pb.WSOutputInterface, ctx context.Context) {
	stream, err := c.Client.CompileSource(ctx, input)
	if err != nil {
		utils.Logger.Infow("Unable to open stream for compilation",
			"message_id", input.MessageId)
	}
	fmt.Println(stream)
	var wg sync.WaitGroup

	wg.Add(1)
	go func(outCh chan<- *pb.WSOutputInterface, wg *sync.WaitGroup) {
		defer wg.Done()
	loop:
		for {
			select {
			case <-ctx.Done():
				utils.Logger.Infow("Context done invoked",
					"message_id", input.MessageId)
				break loop
			default:
				resp, err := stream.Recv()
				if err == io.EOF {
					// TODO: Add close connection message
					break loop
				}
				if err != nil {
					utils.Logger.Infow("Error occured while reading from the stream",
						"message_id", input.MessageId,
						"error", err)
				}
				if resp != nil && resp.StatusCode == constants.STATUS_COMPILE_DONE {
					// TODO: Add close connection message
					outputChan <- resp
					break loop
				}
				outputChan <- resp
			}

		}
	}(outputChan, &wg)

	wg.Wait()
	utils.Logger.Infow("Compilation Done",
		"message_id", input.MessageId)
}
