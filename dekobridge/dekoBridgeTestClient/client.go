package main

import (
	"context"
	"fmt"
	"io"
	"log"

	pb "github.com/dev-saw99/deko/proto"
	constants "github.com/dev-saw99/deko/utils/constant"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

func main() {
	// dial server
	conn, err := grpc.Dial(":50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("can not connect with server %v", err)
	}

	// create stream
	client := pb.NewCompileServiceClient(conn)
	in := &pb.WSInputInterface{
		MessageId: uuid.Must(uuid.NewRandom()).String(),
		Code:      "package main\n\nimport (\n\t\"fmt\"\n\t\"time\"\n)\n\nfunc main() {\n\tfmt.Println(\"hello world 1\")\n\ttime.Sleep(1 * time.Minute);time.Sleep(1 * time.Second)\n\tfmt.Println(\"hello world 2\")\n\ttime.Sleep(1 * time.Second)\n\tfmt.Println(\"hello world 3\")\n\ttime.Sleep(1 * time.Second)\n\n}\n",
		Language:  "go",
	}
	log.Println("new message created", in)
	stream, err := client.CompileSource(context.Background(), in)

	if err != nil {
		log.Fatalf("open stream error %v", err)
	}

	done := make(chan bool)

	go func() {
		for {
			resp, err := stream.Recv()

			if err == io.EOF {
				done <- true //means stream is finished
				return
			}

			if resp != nil && resp.StatusCode == constants.STATUS_COMPILE_DONE {
				done <- true
				return
			}
			if err != nil {
				log.Fatalf("cannot receive %v", err)
			}
			fmt.Println("Response", "Output", resp.Output, "Error", resp.Error, "MessageID", resp.MessageId, "StatusCode", resp.StatusCode)
		}
	}()

	<-done //we will wait until all response is received
	log.Printf("finished")
}
