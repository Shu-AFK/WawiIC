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

func usage() {
	out := flag.CommandLine.Output()

	fmt.Fprintf(out, "WawiIC %s - %s\n\n", defines.Version, defines.Description)
	fmt.Fprint(out, `Aufruf:
  WawiIC.exe [-config <datei>]                 Programm mit Oberfläche starten
  WawiIC.exe -backfill-prices [optionen]       Verkaufskanalpreise nachtragen

Verkaufskanalpreise nachtragen
------------------------------
Vaterartikel, die vor Version 1.0.2 erstellt wurden, haben nur den Standardpreis
bekommen, nicht die Preise der einzelnen Verkaufskanäle. Der Backfill trägt diese
nach.

Ablauf pro Artikel:
  1. Artikel wird direkt über seinen Internen Schlüssel geladen.
  2. Es wird geprüft, ob es wirklich ein Vaterartikel ist. Artikel ohne
     Kindartikel und Artikel, die selbst Kindartikel sind, werden übersprungen.
  3. Alle Kindartikel werden geladen und deren Verkaufskanalpreise verglichen.
  4. Pro Verkaufskanal, Kundengruppe und Staffelmenge wird der niedrigste Preis
     ermittelt und auf den Vaterartikel geschrieben.

Der günstigste Kanalpreis kann also von einem anderen Kindartikel kommen als der
günstigste Preis eines anderen Kanals. Ohne -apply wird nichts geschrieben, der
Lauf listet für jeden Kanal den gefundenen Preis und das Kind, aus dem er stammt.
Ein Lauf kann gefahrlos wiederholt werden, die Preise werden absolut gesetzt.

Beim Zusammenführen von Artikeln in der Oberfläche gilt ab 1.0.4 dieselbe Regel,
neu angelegte und nachgetragene Vaterartikel bekommen also dieselben Preise.

Welche Artikel bearbeitet werden, eine der drei Varianten:

  -backfill-csv <datei>   JTL-Ameise Export. Gelesen wird die Spalte
                          "Interner Schlüssel" (auch kArtikel, Interne ID,
                          Artikel-ID, ItemId, Id). Die Spalte muss nicht die
                          erste sein. Trennzeichen (;  ,  Tab  |) und Kodierung
                          (UTF-16, UTF-8, ANSI) werden erkannt. Der Dateiname
                          ist beliebig.

  -backfill-ids <datei>   Textdatei mit einem Internen Schlüssel pro Zeile.
                          Leerzeilen und Zeilen mit # werden übersprungen.

  Wird eine der beiden Dateien angegeben, läuft der Backfill automatisch,
  -backfill-prices muss dann nicht zusätzlich gesetzt werden.

  ohne beides             Die Kategorie aus der config wird durchsucht. Nur
                          Artikel mit der Anmerkung "Mit API erstellt" werden
                          angefasst. Bei großen Artikelstämmen deutlich langsamer
                          als eine Liste.

Beispiele:
  WawiIC.exe -backfill-prices -backfill-csv export.csv
  WawiIC.exe -backfill-prices -backfill-csv export.csv -apply
  WawiIC.exe -backfill-prices -backfill-ids ids.txt -apply
  WawiIC.exe -backfill-prices -backfill-category 0
  WawiIC.exe -config "D:\configs\custom.json"

Voraussetzung:
  Die App muss in JTL-Wawi registriert sein. Die Berechtigungen für
  Verkaufskanalpreise sind in 1.0.2 dazugekommen. Registrierungen aus älteren
  Versionen müssen einmalig erneuert werden: App-Autorisierung in Wawi
  entfernen, Umgebungsvariable WAWIIC_APIKEY löschen, neues Fenster öffnen,
  Programm starten.

Optionen:
`)
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	defaultPath := defines.ConfigPath
	cfgFlag := flag.String("config", defaultPath, "Pfad zur config.json")
	pauseFlag := flag.Bool("pause", false, "am Ende auf Enter warten, bevor das Fenster schließt")
	backfillFlag := flag.Bool("backfill-prices", false, "Verkaufskanalpreise auf bestehende Vaterartikel nachtragen")
	applyFlag := flag.Bool("apply", false, "zusammen mit -backfill-prices: Änderungen wirklich schreiben (sonst nur Testlauf)")
	backfillCatFlag := flag.Int("backfill-category", -1, "zusammen mit -backfill-prices: zu durchsuchende Kategorie (Standard: die aus der config, 0 = alle Artikel)")
	backfillIDsFlag := flag.String("backfill-ids", "", "Datei mit einem Internen Schlüssel pro Zeile (impliziert -backfill-prices)")
	backfillCSVFlag := flag.String("backfill-csv", "", "JTL-Ameise Export, die Internen Schlüssel daraus werden bearbeitet (impliziert -backfill-prices)")
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

	// Read the lists before anything touches the network, so a bad file surfaces
	// immediately instead of after the registration round trip.
	var itemIDs []int
	if *backfillCSVFlag != "" {
		data, err := os.ReadFile(*backfillCSVFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot read %s: %v\n", *backfillCSVFlag, err)
			exit(2, *pauseFlag)
		}
		itemIDs, err = wawi.ParseAmeiseCSV(data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", *backfillCSVFlag, err)
			exit(2, *pauseFlag)
		}
		fmt.Printf("%d Interne Schlüssel aus %s gelesen.\n", len(itemIDs), *backfillCSVFlag)
	}

	if *backfillIDsFlag != "" {
		data, err := os.ReadFile(*backfillIDsFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot read %s: %v\n", *backfillIDsFlag, err)
			exit(2, *pauseFlag)
		}
		itemIDs, err = wawi.ParseItemIDs(string(data))
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", *backfillIDsFlag, err)
			exit(2, *pauseFlag)
		}
		fmt.Printf("%d Artikel-IDs aus %s gelesen.\n", len(itemIDs), *backfillIDsFlag)
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

	// Handing over a list can only mean the backfill. Falling through to the GUI
	// here would look like the program simply ignored the file.
	if *backfillCSVFlag != "" || *backfillIDsFlag != "" {
		*backfillFlag = true
	}

	if *backfillFlag {
		category := *backfillCatFlag
		if category < 0 {
			category = wawi.ConfiguredCategoryID()
		}

		if err := runBackfill(itemIDs, category, *applyFlag); err != nil {
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

func runBackfill(itemIDs []int, categoryID int, apply bool) error {
	if apply {
		fmt.Println("Backfill wird ausgeführt, Änderungen werden geschrieben.")
	} else {
		fmt.Println("Backfill als Testlauf, es wird nichts geschrieben. Mit -apply ausführen.")
	}
	if len(itemIDs) == 0 && categoryID == 0 {
		fmt.Println("Warnung: ohne Kategorie wird der gesamte Artikelstamm durchsucht, das kann sehr lange dauern.")
	}

	results, err := wawi.BackfillSalesChannelPrices(wawi.BackfillOptions{
		ItemIDs:    itemIDs,
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
			verb := "würden gesetzt"
			if apply {
				verb = "gesetzt"
			}
			fmt.Printf("  %s: %d Preise aus %d Kindartikeln %s\n", res.ParentSKU, res.Prices, res.Children, verb)

			// In a test run the detail is the whole point: it shows which child
			// won each shop, so a few can be checked against Wawi before writing.
			if !apply {
				for _, detail := range res.Details {
					fmt.Printf("      %s\n", formatPriceDetail(detail))
				}
			}
		}
	}

	fmt.Printf("\nFertig: %d aktualisiert, %d übersprungen, %d fehlgeschlagen.\n", updated, skipped, failed)
	if failed > 0 {
		return fmt.Errorf("%d Artikel konnten nicht aktualisiert werden", failed)
	}

	return nil
}

func formatPriceDetail(detail wawi.BackfillPrice) string {
	price := detail.Price

	amount := "-"
	switch {
	case price.NetPrice != nil:
		amount = fmt.Sprintf("%.2f netto", *price.NetPrice)
	case price.ReduceStandardPriceByPercent != nil:
		amount = fmt.Sprintf("-%.2f%%", *price.ReduceStandardPriceByPercent)
	}

	tier := ""
	if price.FromQuantity > 0 {
		tier = fmt.Sprintf(" ab %d Stück", price.FromQuantity)
	}

	return fmt.Sprintf("Kanal %s, Kundengruppe %d%s: %s (von %s)",
		price.SalesChannelId, price.CustomerGroupId, tier, amount, detail.SourceSKU)
}
