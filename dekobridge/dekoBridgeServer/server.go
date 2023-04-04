package dekobridge

import (
	"context"
	"errors"
	"io"
	"net"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	v1 "github.com/dev-saw99/deko/proto"
	"github.com/dev-saw99/deko/utils"
	constants "github.com/dev-saw99/deko/utils/constant"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type compileServiceServer struct {
	v1.UnimplementedCompileServiceServer
}

func (c *compileServiceServer) CompileSource(input *v1.WSInputInterface, stream v1.CompileService_CompileSourceServer) error {
	utils.Logger.Infow("Running Source Code")
	outputChan, err := parseAndExecute(input)
	if err != nil {
		wsOutput := &v1.WSOutputInterface{
			Error:      err.Error(),
			MessageId:  input.MessageId,
			StatusCode: constants.STATUS_LANGUAGE_INVALID,
		}
		err := stream.SendMsg(&wsOutput)
		if err != nil {
			utils.Logger.Infow("Error occured while sending the wsOutput",
				"error", err,
				"message_id", input.GetMessageId(),
				"wsOutput", wsOutput)
		}
		return err
	}

	for {
		select {
		case <-stream.Context().Done():
			utils.Logger.Infow("Stream has ended",
				"message_id", input.GetMessageId())
			return status.Error(codes.Canceled, "Stream has ended")
		case output := <-outputChan:
			err := stream.SendMsg(output)
			if err != nil {
				utils.Logger.Infow("Error occured while sending the wsOutput",
					"error", err,
					"message_id", input.GetMessageId(),
					"wsOutput", output)
				return err
			}
		}
	}

}

func parseAndExecute(input *v1.WSInputInterface) (<-chan *v1.WSOutputInterface, error) {
	outputChan := make(chan *v1.WSOutputInterface, 5)
	var runCommand string
	var runArguments []string
	var fileExtension string
	switch input.GetLanguage() {
	case constants.PYTHON:
		runCommand = constants.PYTHON_CMD
		runArguments = append(runArguments, "-u")
		fileExtension = constants.PYTHON_EXT
	case constants.GO:
		runCommand = constants.GO_CMD
		runArguments = append(runArguments, "run")
		fileExtension = constants.GO_EXT
	default:
		utils.Logger.Infow("Language not supported yet",
			"language", input.GetLanguage())
		return nil, errors.New(constants.MSG_LANGUAGE_NOT_SUPPORTED)
	}
	go runCode(outputChan, input.GetCode(), input.GetMessageId(), fileExtension, runCommand, runArguments)

	return outputChan, nil
}

func killProcess(ctx context.Context, cancel context.CancelFunc, outputChan chan<- *v1.WSOutputInterface, messageId string, wg *sync.WaitGroup) {
	defer wg.Done()
	var timeOut bool

loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case <-time.After(constants.TIMEOUT_DURATION):
			utils.Logger.Infow("TIMEOUT, Killing the process",
				"message_id", messageId)
			timeOut = true
			cancel()
			break loop
		}
	}
	if timeOut {
		outputChan <- &v1.WSOutputInterface{
			Message:    "TIMEOUT",
			MessageId:  messageId,
			StatusCode: constants.STATUS_TIMEOUT,
		}
	}
}

func deleteFileAndFolder(folderName string, messageId string) {
	err := os.RemoveAll(folderName)
	if err != nil {
		utils.Logger.Infow("Error while removing file",
			"error", err,
			"foldername", folderName,
			"message_id", messageId)
	}
	utils.Logger.Infow("Removed folder",
		"foldername", folderName,
		"message_id", messageId,
	)
}

func createFile(filename string, messageId string, sourceCode string) error {
	file, err := os.Create(filename)
	if err != nil {
		utils.Logger.Infow("Unable to create a file for compile and run",
			"error", err,
			"message_id", messageId)
		return err
	}
	io.WriteString(file, sourceCode)
	utils.Logger.Infow("created file",
		"filename", filename,
		"message_id", messageId,
	)
	return nil
}

func runCode(outputChan chan<- *v1.WSOutputInterface, sourceCode string, messageId string, fileExtension string, runCmd string, runArgs []string) {

	// create a file
	rootDir := constants.SOURCE_CODE_DIR
	jailDir := rootDir + "/" + messageId
	err := os.MkdirAll(jailDir, 0755)
	if err != nil {
		utils.Logger.Infow("Unable to create a file for compile and run",
			"error", err,
			"message_id", messageId)

		outputChan <- &v1.WSOutputInterface{
			Message:    "Internal Server Error",
			StatusCode: constants.STATUS_DEFAULT_ERROR,
			MessageId:  messageId,
		}
		return
	}
	filename := jailDir + "/" + messageId + fileExtension
	if err := createFile(filename, messageId, sourceCode); err != nil {
		outputChan <- &v1.WSOutputInterface{
			Message:    "Internal Server Error",
			StatusCode: constants.STATUS_DEFAULT_ERROR,
			MessageId:  messageId,
		}
		return
	} // delete files and folders once done.
	defer deleteFileAndFolder(jailDir, messageId)

	runArgs = append(runArgs, filename)
	// create command to execute
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "pwd")

	// set chroot jail, chroot jail is a way to isolate a process from the rest of the system by setting its root directory to a specific directory.
	// cmd.SysProcAttr = &syscall.SysProcAttr{
	// 	Chroot: rootDir,
	// }
	// cmd.Dir = jailDir
	// save the pipes to read ouput and error
	outputPipe, err := cmd.StdoutPipe()
	if err != nil {
		utils.Logger.Infow("Error configuring the command to run",
			"error", err,
			"message_id", messageId)
		cancel()
		return
	}
	errPipe, err := cmd.StderrPipe()
	if err != nil {
		utils.Logger.Infow("Error configuring the command to run",
			"error", err,
			"message_id", messageId)
		cancel()
		return
	}

	// start the command
	err = cmd.Start()
	if err != nil {
		utils.Logger.Infow("Error starting the command",
			"error", err,
			"cmd", cmd,
			"message_id", messageId)
		cancel()
		return
	}

	utils.Logger.Infow("Compilation started",
		"message_id", messageId)

	outputChan <- &v1.WSOutputInterface{
		StatusCode: constants.STATUS_COMPILE_START,
		MessageId:  messageId,
		Message:    constants.MSG_COMPILE_STARTED,
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go readDataFromPipe(errPipe, messageId, constants.PIPE_TYPE_ERROR, outputChan, ctx, &wg)
	wg.Add(1)
	go readDataFromPipe(outputPipe, messageId, constants.PIPE_TYPE_OUTPUT, outputChan, ctx, &wg)
	wg.Add(1)
	go killProcess(ctx, cancel, outputChan, messageId, &wg)

	wg.Wait()
	utils.Logger.Infow("Compilation completed",
		"message_id", messageId)
	outputChan <- &v1.WSOutputInterface{
		StatusCode: constants.STATUS_COMPILE_DONE,
		MessageId:  messageId,
		Message:    constants.MSG_COMPILE_CODE,
	}

	err = cmd.Wait()
	if err != nil {
		utils.Logger.Infow("Error exiting the command",
			"error", err,
			"cmd", cmd,
			"message_id", messageId)
	}
}

func readDataFromPipe(pipe io.ReadCloser, messageId string, pipeType string, outputChan chan<- *v1.WSOutputInterface, ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			utils.Logger.Infow("Context done called",
				"pipeType", pipeType,
				"message_id", messageId,
			)
			return
		default:
			buff := make([]byte, 1024)
			n, err := pipe.Read(buff)
			// line, err := reader.ReadString('\n')
			// fmt.Println("asdasd")
			if err != nil {
				if err == io.EOF {
					utils.Logger.Infow("Done reading from the pipe",
						"pipeType", pipeType,
						"message_id", messageId,
					)
					return
				} else {
					utils.Logger.Infow("Error while reading from pipe",
						"pipeType", pipeType,
						"error", err,
						"message_id", messageId)
					return
				}
			}
			if err == nil && n > 0 {
				utils.Logger.Infow("asd",
					"pipeType", pipeType,
					"message_id", messageId,
				)
				wsOutput := v1.WSOutputInterface{}
				if pipeType == constants.PIPE_TYPE_ERROR {
					wsOutput.Error = string(buff[:n])
				} else if pipeType == constants.PIPE_TYPE_OUTPUT {
					wsOutput.Output = string(buff[:n])
				}
				wsOutput.StatusCode = constants.STATUS_COMPILE_SUCCESS
				wsOutput.MessageId = messageId
				outputChan <- &wsOutput
			}
		}
	}
}

// deployServer start a gRPC server and listens for the connections at port 50051
func deployServer() {
	isSandboxEnv, err := strconv.ParseBool(os.Getenv("SANDBOX_ENV"))
	if err != nil {
		utils.Logger.Infow("SANDBOX_ENV env variable not found, setting isSandboxEnv to False")
		isSandboxEnv = false
	}
	var compilerDNS string

	if isSandboxEnv {
		compilerDNS = constants.DEKO_BRIDGE_SANDBOX_CONTAINER_HOST_PORT
	} else {
		compilerDNS = constants.DEKO_BRIDGE_LOCALHOST_CONTAINER_HOST_PORT
	}

	utils.Logger.Infow("Creating listener for gRPC server",
		"port", compilerDNS)

	lis, err := net.Listen("tcp", compilerDNS)
	if err != nil {
		utils.Logger.Errorw("Unable to create the listener gRPC server",
			"error", err)
		return
	}

	server := grpc.NewServer()

	v1.RegisterCompileServiceServer(server, &compileServiceServer{})

	utils.Logger.Infow("Successfully registered Compiler Service gRPC server")

	err = server.Serve(lis)
	if err != nil {
		utils.Logger.Errorw("Unable to start the gRPC server",
			"error", err)
		return
	}

}

func Process() {
	utils.Logger.Infow("Starting DekoBridge ...")
	os.MkdirAll(constants.SOURCE_CODE_DIR, 0755)
	defer os.RemoveAll(constants.SOURCE_CODE_DIR)
	deployServer()
}
