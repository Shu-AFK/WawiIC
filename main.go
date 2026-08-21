package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/Shu-AFK/WawiIC/cmd/defines"
	"github.com/Shu-AFK/WawiIC/cmd/gui"
	"github.com/Shu-AFK/WawiIC/cmd/openai"
	"github.com/Shu-AFK/WawiIC/cmd/wawi"
	"github.com/Shu-AFK/WawiIC/cmd/wawi/wawi_registration"
	"github.com/Shu-AFK/WawiIC/cmd/wawi/wawi_structs"
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
     ermittelt und auf den Vaterartikel geschrieben. Fehlt die Preiszeile am
     Vaterartikel noch, wird sie angelegt.
  5. Zusätzlich bekommt der Vaterartikel den niedrigsten Standard-, eBay- und
     Amazonpreis seiner Kindartikel. Ohne das würde ein Shop für eine
     Kundengruppe ohne eigene Preiszeile weiter den alten Preis zeigen.
  6. Geschrieben wird von allgemein nach speziell: erst die Artikelpreise, dann
     der Kanal 9-7-1-2, dann die übrigen Kanäle. So überschreibt der allgemeine
     Preis nicht den spezifischeren.

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
  Die App muss in JTL-Wawi registriert sein. Ab 1.0.7 meldet sie sich als
  WawiIC/v2 an, weil neue Berechtigungen dazugekommen sind. Beim ersten Start
  einfach die Umgebungsvariable WAWIIC_APIKEY löschen, ein neues Fenster öffnen
  und das Programm starten, dann läuft die Registrierung neu. Die alte
  Autorisierung WawiIC/v1 kann in Wawi stehen bleiben.

Diagnose:
  -sales-channels         Listet alle Verkaufskanäle des Systems mit ID, Typ,
                          Name und ob sie Artikelpreise annehmen. Hilfreich,
                          wenn ein Preis abgelehnt wird.

  -inspect <id>           Gibt zu einem Internen Schlüssel den Artikel und alle
                          seine Kindartikel aus, jeweils mit sämtlichen
                          Verkaufskanalpreisen so, wie die API sie sieht. Zeigt,
                          unter welchem Kanal, welcher Kundengruppe und welcher
                          Staffelmenge ein Preis tatsächlich liegt.

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
	channelsFlag := flag.Bool("sales-channels", false, "die Verkaufskanäle des Systems auflisten und beenden")
	inspectFlag := flag.Int("inspect", 0, "einen Artikel und seine Kindartikel mit allen Verkaufskanalpreisen ausgeben")
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

	if *inspectFlag != 0 {
		if err := inspectItem(*inspectFlag); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			exit(1, *pauseFlag)
		}
		exit(0, *pauseFlag)
	}

	if *channelsFlag {
		if err := printSalesChannels(); err != nil {
			fmt.Fprintf(os.Stderr, "Verkaufskanäle konnten nicht geladen werden: %v\n", err)
			exit(1, *pauseFlag)
		}
		exit(0, *pauseFlag)
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

	// Printed once so a rejected price can be matched against the channel it was
	// meant for.
	if err := printSalesChannels(); err != nil {
		fmt.Fprintf(os.Stderr, "Verkaufskanäle konnten nicht geladen werden: %v\n", err)
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
				if res.PriceData != nil {
					fmt.Printf("      %s\n", formatPriceData(*res.PriceData))
				}
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

// printSalesChannels lists what the system actually has. Sales channel ids are
// not guessable and differ per installation, so a rejected price is only
// explainable next to this list.
func printSalesChannels() error {
	channels, err := wawi.QuerySalesChannels()
	if err != nil {
		return err
	}

	sort.Slice(channels, func(i, j int) bool { return channels[i].Id < channels[j].Id })

	fmt.Printf("\nVerkaufskanäle (%d):\n", len(channels))
	fmt.Printf("  %-12s %-6s %-8s %s\n", "ID", "Typ", "Preise", "Name")
	for _, channel := range channels {
		prices := "nein"
		if channel.ItemCapabilities.Prices {
			prices = "ja"
		}
		fmt.Printf("  %-12s %-6d %-8s %s\n", channel.Id, channel.Type, prices, channel.Name)
	}
	fmt.Println()

	return nil
}

func formatPriceData(data wawi_structs.ItemPriceData) string {
	parts := make([]string, 0, 3)
	for _, field := range []struct {
		name  string
		value *float64
	}{
		{"Standardpreis", data.SalesPriceNet},
		{"eBay", data.EbayPrice},
		{"Amazon", data.AmazonPrice},
	} {
		if field.value != nil {
			parts = append(parts, fmt.Sprintf("%s %.2f netto", field.name, *field.value))
		}
	}

	if len(parts) == 0 {
		return "Artikelpreise: kein Kindartikel hat einen Preis"
	}

	return "Artikelpreise: " + strings.Join(parts, ", ")
}

// inspectItem prints what the API reports for one item and its children. Wawi's
// own screen and the API do not always agree on which key a price sits under,
// and only the API view explains why a write is refused.
func inspectItem(id int) error {
	parent, found, err := wawi.GetItemByID(id)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("Artikel %d nicht gefunden", id)
	}

	printItemPrices("VATER", *parent)

	for _, childID := range parent.ChildItems {
		child, found, err := wawi.GetItemByID(childID)
		if err != nil {
			return err
		}
		if !found {
			fmt.Printf("\nKIND %d nicht gefunden\n", childID)
			continue
		}
		printItemPrices("KIND", *child)
	}

	return nil
}

func printItemPrices(label string, item wawi_structs.GetItem) {
	fmt.Printf("\n%s %d %s\n", label, item.ID, item.SKU)
	fmt.Printf("  %s\n", formatPriceData(item.ItemPriceData))

	prices, err := wawi.QueryItemSalesChannelPrices(strconv.Itoa(item.ID))
	if err != nil {
		fmt.Printf("  Verkaufskanalpreise: FEHLER %v\n", err)
		return
	}
	if len(prices) == 0 {
		fmt.Printf("  Verkaufskanalpreise: keine\n")
		return
	}

	fmt.Printf("  %-12s %-7s %-7s %-12s %s\n", "Kanal", "Gruppe", "abMenge", "Netto", "Rabatt%")
	for _, price := range prices {
		fmt.Printf("  %-12s %-7d %-7d %-12s %s\n",
			price.SalesChannelId,
			price.CustomerGroupId,
			price.FromQuantity,
			formatOptionalPrice(price.NetPrice),
			formatOptionalPrice(price.ReduceStandardPriceByPercent),
		)
	}

	// The parsed view hides any field the model does not know about, and one of
	// those is what tells two rows with the same key apart.
	if raw, err := wawi.QueryItemSalesChannelPricesRaw(strconv.Itoa(item.ID)); err == nil {
		fmt.Printf("  Rohdaten: %s\n", string(raw))
	}
}

func formatOptionalPrice(value *float64) string {
	if value == nil {
		return "-"
	}

	return fmt.Sprintf("%g", *value)
}
