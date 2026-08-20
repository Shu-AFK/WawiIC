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
	backfillFlag := flag.Bool("backfill-prices", false, "copy sales channel prices onto parent items created before this was supported")
	applyFlag := flag.Bool("apply", false, "with -backfill-prices: write the changes instead of only reporting them")
	backfillCatFlag := flag.Int("backfill-category", -1, "with -backfill-prices: category to search (default: the configured category, 0 searches every item)")
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

	if *backfillFlag {
		category := *backfillCatFlag
		if category < 0 {
			category = wawi.ConfiguredCategoryID()
		}
		if err := runBackfill(category, *applyFlag); err != nil {
			fmt.Fprintf(os.Stderr, "backfill failed: %v\n", err)
			exit(1, *pauseFlag)
		}
		exit(0, *pauseFlag)
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

func runBackfill(categoryID int, apply bool) error {
	if apply {
		fmt.Println("Backfill wird ausgeführt, Änderungen werden geschrieben.")
	} else {
		fmt.Println("Backfill als Testlauf, es wird nichts geschrieben. Mit -apply ausführen.")
	}
	if categoryID == 0 {
		fmt.Println("Warnung: ohne Kategorie wird der gesamte Artikelstamm durchsucht, das kann sehr lange dauern.")
	}

	results, err := wawi.BackfillSalesChannelPrices(wawi.BackfillOptions{
		CategoryID: categoryID,
		Apply:      apply,
	}, func(msg string) {
		fmt.Println(msg)
	})
	if err != nil {
		return err
	}

	var updated, skipped, failed int
	for _, res := range results {
		switch {
		case res.Err != nil:
			failed++
			fmt.Fprintf(os.Stderr, "  FEHLER %s: %v\n", res.ParentSKU, res.Err)
		case res.Skipped != "":
			skipped++
			fmt.Printf("  ÜBERSPRUNGEN %s: %s\n", res.ParentSKU, res.Skipped)
		default:
			updated++
			verb := "würde übernehmen"
			if apply {
				verb = "übernommen"
			}
			fmt.Printf("  %s: %d Preise von %s %s\n", res.ParentSKU, res.Prices, res.SourceSKU, verb)
		}
	}

	fmt.Printf("\nFertig: %d aktualisiert, %d übersprungen, %d fehlgeschlagen.\n", updated, skipped, failed)
	if failed > 0 {
		return fmt.Errorf("%d Artikel konnten nicht aktualisiert werden", failed)
	}

	return nil
}
