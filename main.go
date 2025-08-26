package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Shu-AFK/WawiIC/cmd/defines"
	"github.com/Shu-AFK/WawiIC/cmd/gui"
	"github.com/Shu-AFK/WawiIC/cmd/openai"
	"github.com/Shu-AFK/WawiIC/cmd/wawi"
	"github.com/Shu-AFK/WawiIC/cmd/wawi/wawi_registration"
)

func main() {
	defaultPath := defines.ConfigPath
	cfgFlag := flag.String("config", defaultPath, "config file path")
	flag.Parse()

	cfgPath := *cfgFlag
	if cfgPath == "" {
		cfgPath = defaultPath
	}

	if !strings.EqualFold(filepath.Ext(cfgPath), ".json") {
		fmt.Fprintf(os.Stderr, "error: -c must point to a .json file (got %q)\n", cfgPath)
		os.Exit(2)
	}
	if _, err := os.Stat(cfgPath); err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "error: config file not found: %s\n", cfgPath)
		} else {
			fmt.Fprintf(os.Stderr, "error: cannot access config file %s: %v\n", cfgPath, err)
		}
		os.Exit(2)
	}

	err := wawi.LoadConfig(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

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

	fmt.Println(wawi.QuerySalesChannels())

	err = openai.CheckForAPIKey()
	if err != nil {
		panic(err)
	}

	gui.RunGUI()
}
