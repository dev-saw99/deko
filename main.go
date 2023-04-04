package main

import (
	"os"
	"strconv"

	"github.com/dev-saw99/deko/compiler"
	"github.com/dev-saw99/deko/handler"
	"github.com/dev-saw99/deko/router"
	"github.com/dev-saw99/deko/utils"
	constants "github.com/dev-saw99/deko/utils/constant"
)

func main() {

	utils.InitializeLogger(constants.DEKO_LOG_FILE)
	utils.Logger.Infow("Starting Deko Server")

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
	handler.CompilerClient = compiler.NewCompiler(compilerDNS)
	utils.Logger.Infow("Initialising Compiler Client")
	handler.CompilerClient.Init()

	utils.Logger.Infow("Creating New Router",
		"port", ":9000")
	r := router.NewRouter()
	r.Run(":9000")

}
