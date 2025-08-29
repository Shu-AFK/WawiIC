package main

import (
	"bufio"
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

func pauseIfNeeded(enabled bool) {
	if !enabled {
		if fi, err := os.Stdin.Stat(); err == nil && (fi.Mode()&os.ModeCharDevice) == 0 {
			return
		}
	}
	fmt.Print("Press Enter to exit...")
	_, _ = bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func exit(code int, pause bool) {
	pauseIfNeeded(pause)
	os.Exit(code)
}

func main() {
	defaultPath := defines.ConfigPath
	cfgFlag := flag.String("config", defaultPath, "config file path")
	pauseFlag := flag.Bool("pause", false, "wait for Enter before exit")
	flag.Parse()

	cfgPath := *cfgFlag
	if cfgPath == "" {
		cfgPath = defaultPath
	}
	defines.ConfigPath = cfgPath

	if !strings.EqualFold(filepath.Ext(cfgPath), ".json") {
		fmt.Fprintf(os.Stderr, "error: -config must point to a .json file (got %q)\n", cfgPath)
		exit(2, *pauseFlag)
	}
	if _, err := os.Stat(cfgPath); err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "error: config file not found: %s\n", cfgPath)
		} else {
			fmt.Fprintf(os.Stderr, "error: cannot access config file %s: %v\n", cfgPath, err)
		}
		exit(2, *pauseFlag)
	}

	fmt.Printf("Loading config from %s...\n", cfgPath)
	err := wawi.LoadConfig(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		exit(1, *pauseFlag)
	}

	fmt.Println("Checking for Wawi API key...")
	_, exists := os.LookupEnv(defines.APIKeyVarName)
	if !exists {
		fmt.Println("Wawi API key not found. Registering...")
		apiKey, err := wawi_registration.Register()
		if err != nil {
			fmt.Fprintf(os.Stderr, "registration failed: %v\n", err)
			exit(1, *pauseFlag)
		}

		cmd := exec.Command("setx", defines.APIKeyVarName, apiKey)
		err = cmd.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to set environment variable: %v\n", err)
			exit(1, *pauseFlag)
		}

		err = os.Setenv(defines.APIKeyVarName, apiKey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to set environment variable: %v\n", err)
			exit(1, *pauseFlag)
		}
	} else {
		fmt.Println("Wawi API key found.")
	}

	fmt.Println("Checking for OpenAI API key...")
	err = openai.CheckForAPIKey()
	if err != nil {
		fmt.Fprintf(os.Stderr, "OpenAI API key check failed: %v\n", err)
		exit(1, *pauseFlag)
	}
	fmt.Println("OpenAI API key found.")

	gui.RunGUI()
	pauseIfNeeded(*pauseFlag)
}
