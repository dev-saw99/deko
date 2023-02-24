package main

import (
	dekobridge "github.com/dev-saw99/deko/dekobridge/dekoBridgeServer"
	"github.com/dev-saw99/deko/utils"
	constants "github.com/dev-saw99/deko/utils/constant"
)

func main() {
	utils.InitializeLogger(constants.DEKO_BRIDGE_LOG_FILE)
	utils.Logger.Infow("Intialising Deko Bridge")
	dekobridge.Process()
}
