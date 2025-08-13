package main

import (
	"WawiIC/defines"
	"WawiIC/registration"
	"os"
	"os/exec"
)

func main() {
	_, exists := os.LookupEnv(defines.APIKeyVarName)

	if !exists {
		apiKey, err := registration.Register()
		if err != nil {
			panic(err)
		}

		cmd := exec.Command("setx", defines.APIKeyVarName, apiKey)
		err = cmd.Run()
		if err != nil {
			panic(err)
		}
	}
}
