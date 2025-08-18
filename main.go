package main

import (
	"os"
	"os/exec"

	"github.com/Shu-AFK/WawiIC/cmd/defines"
	"github.com/Shu-AFK/WawiIC/cmd/gui"
	"github.com/Shu-AFK/WawiIC/cmd/openai"
	"github.com/Shu-AFK/WawiIC/cmd/wawi/wawi_registration"
)

func main() {
	_, exists := os.LookupEnv(defines.APIKeyVarName)

	if !exists {
		apiKey, err := wawi_registration.Register()
		if err != nil {
			panic(err)
		}

		cmd := exec.Command("setx", defines.APIKeyVarName, apiKey)
		err = cmd.Run()
		if err != nil {
			panic(err)
		}

		err = os.Setenv(defines.APIKeyVarName, apiKey)
		if err != nil {
			panic(err)
		}
	}

	err := openai.CheckForAPIKey()
	if err != nil {
		panic(err)
	}

	gui.RunGUI()
}
