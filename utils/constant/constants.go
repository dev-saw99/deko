package constants

import "time"

const (
	STATUS_ALREADY_PROCESSING = iota + 3000
	STATUS_COMPILE_START
	STATUS_COMPILE_SUCCESS
	STATUS_COMPILE_DONE
	STATUS_LANGUAGE_INVALID
	STATUS_CODE_INVALID
	STATUS_TIMEOUT
	STATUS_CONN_CLOSE
	STATUS_DEFAULT_ERROR
	STATUS_INVALID_MESSAGE
)

const (
	MSG_CLOSE_CONNECTION = "close-conn"
	MSG_COMPILE_CODE     = "compile-code"
)

const (
	SOURCE_CODE_EMPTY = "Source Code Empty"
	LANGUAGE_EMPTY    = "Language Empty"
)

const (
	PYTHON     = "python"
	PYTHON_CMD = "python3"
	PYTHON_EXT = ".py"

	GO     = "go"
	GO_EXT = ".go"
	GO_CMD = "go"

	PIPE_TYPE_ERROR  = "pipe-error"
	PIPE_TYPE_OUTPUT = "pipe-output"

	SOURCE_CODE_DIR = "sourceCodeDir"
)

const (
	MSG_LANGUAGE_NOT_SUPPORTED = "language not supported"
	MSG_COMPILE_STARTED        = "Executing ..."
	MSG_COMPILE_COMPLETED      = "Done ..."
)

const (
	TIMEOUT_DURATION = time.Second * 30
)

const (
	// configuration
	DEKO_BRIDGE_CONTAINER_HOST_PORT = "deko-bridge:50051"
	DEKO_BRIDGE_LOG_FILE            = "deko-bridge.log"
	DEKO_LOG_FILE                   = "deko.log"
)
