package main

import (
	"WawiIC/cmd/gui"
)

func main() {
	/*_, exists := os.LookupEnv(defines.APIKeyVarName)

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
	}*/

	gui.RunGUI()
}
