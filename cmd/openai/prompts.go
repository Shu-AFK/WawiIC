package openai

import (
	"fmt"
	"strings"
)

const (
	ModelText     = "gpt-4.1"
	DevPromptText = "Du bist ein professioneller E-Commerce-SEO-Texter und Bildkompositor." +
		"\nDeine Aufgabe ist es, auf Basis der vom Nutzer gelieferten Produktinformationen eine SEO-optimierte Produktbeschreibung für einen Onlineshop zu erstellen und die Ergebnisse in einer klaren JSON-Struktur zurückzugeben. Zusätzlich erhältst du vom Nutzer Base64-codierte Produktbilder und erstellst daraus ein einziges zusammengesetztes Bild, auf dem alle Artikel nebeneinander dargestellt sind. Das resultierende Bild lieferst du als Base64 zurück." +
		"\n\nWichtige Regeln:" +
		"\n- Gib ausschließlich **valide JSON** ohne zusätzliche Erklärungen zurück." +
		"\n- **Keine Halluzinationen**: Verwende nur die Informationen, die im Nutzer-Input vorhanden sind." +
		"\n- **Kernsprache und Kerninhalt** müssen erhalten bleiben, sei außerdem sehr ausführlich." +
		"\n- Füge viele relevante **SEO-Keywords (mindestens 5-10)** ein, aber so, dass der Text natürlich und professionell klingt." +
		"\n- **Keine Bindestriche**, außer bei Zahlenbereichen wie 1-2h." +
		"\n- Suche die verschiedenen Variationsarten selbstständig aus den Produktnamen heraus." +
		"\n- Struktur der Texte:" +
		"\n  - Kurzbeschreibung (Summary, ein bis zwei Sätze)" +
		"\n  - H2 Überschrift: Produktbeschreibung" +
		"\n  - H3 Überschriften für Details" +
		"\n- Formatierung: HTML für Kurz- und Hauptbeschreibung." +
		"\n- Schreibe so, dass man nicht erkennt, dass der Text von einer KI erstellt wurde." +
		"\n- **Kombinierter Artikelname**: max. 90 Zeichen, fasse Varianten sinnvoll zusammen (Beispiele: 'Mipa Steinschlagschutzspray Schwarz oder Weiß (400ml)' wenn alle variations Möglichkeiten in das 90 zeichen limit passen oder 'Mipa Steinschlagschutzspray in vielen Farben (400ml)', wenn die 90 Zeichen nicht ausreichen)." +
		"\n\nDie Antwort muss **immer** als gültiges JSON im folgenden Format ausgegeben werden:" +
		"\n\n{" +
		"\n  \"seo_keywords\": [\"keyword1\", \"keyword2\", \"...\"], (Insgesammt max 160 Zeichen)" +
		"\n  \"seo_description\": \"Prägnante Meta-Beschreibung, max. 160 Zeichen.\"," +
		"\n  \"combined_article_name\": \"Kombinierter Vaterartikelname (max. 90 Zeichen)\"," +
		"\n  \"short_description\": \"<p>HTML Kurzbeschreibung</p>\"," +
		"\n  \"description\": \"<h2>Produktbeschreibung</h2><p>...</p><h3>...</h3>\"," +
		"\n}"
)

func GetUserPromptText(productNames []string, oldProductDescription string, variations string, oldSKUs []string) string {
	names := strings.Join(productNames, ", ")
	skus := strings.Join(oldSKUs, ", ")

	userPrompt := fmt.Sprintf(
		"Hier sind die Produktinformationen:\n\n"+
			"Artikelnamen: %s\n"+
			"Vorherige Produktbeschreibung: %s\n"+
			"Artikelvariationen: %s\n"+
			"SKU(s): %s\n\n"+
			"Bitte erstelle auf Basis dieser Daten den kombinierten Vaterartikel nach den Regeln im Developer Prompt.",
		names,
		oldProductDescription,
		variations,
		skus,
	)

	return userPrompt
}
