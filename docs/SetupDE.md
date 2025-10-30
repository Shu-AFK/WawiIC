### Starten des JTL-Wawi API servers
> Im Installationspfad von JTL-Wawi ist eine Datei names `JTL.Wawi.Rest.exe`. Dies ist der API server. Der standard Installationspfad ist `C:\Programm Files (x86)\JTL-Software`. Um den server zu starten muss man die folgenden befehle innerhalb einer Konsole ausführen:

```sh
> cd C:\Programm Files (x86)\JTL-Software
> JTL.Wawi.Rest.exe -w "Standard" -l 127.0.0.1
```

> Der Pfad nach `cd` ist mit dem richtigen Pfad zu JTL-Software zu ersetzen und falls man einen anderen Profilname für JTL-Wawi eingerichtet hat, muss man diesen anstatt `Standard` in die Anführungszeichen schreiben.

### Was man braucht
#### 1. Wawi API Key
> Falls noch keiner vorhanden ist in JTL Wawi Admin -> App Registrierung -> Hinzufügen Weiter, WawiIC, nachdem der API server von Wawi läuft, starten und dann die Berechtigungen akzeptieren und fertig, der API Key ist nun in den Umgebungsvariablen gespeichert. Um einen bereits existierenden key zu verwenden, diesen in die Umgebungsvariable "WAWIIC_APIKEY" speichern.

#### 2. OpenAi API Key 
> Ein OpenAi API Key kann [hier](https://platform.openai.com/api-keys) erstellt werden. Danach den key einfach in die Umgebungsvariable "OPENAI_API_KEY" speichern.

#### 3. Bilder
> Damit WawiIC die Bilder von Artikeln zusammenfassen kann, bzw. die Bilder für den Vaterartikel hochladen kann, braucht das Programm einen Export mit allen Bildern von den Artikeln in einem Ordner. Dabei sollten die Bilder alle folgend heißen: "[Artikelnummer]-[Bild nr.].[jpg/png]". Dies kann man am besten machen mit der JTL-Ameise. Nach dem Export muss man den Pfad zu diesem Ordner in der Config angeben, mehr dazu [hier](###config). 

### Config

- Standardpfad: `config/config.json`  
- Alternativ: Pfad per `-c` Flag beim Starten des Programms mitgeben  
  Beispiel:
  ```sh
  WawiIC.exe -c "D:\pfad\zu\meiner\config.json"
  ```

---
#### Aufbau
- Typ: JSON
- Inhalt:
	- `api base url`: string - Basis-URL der JTL-Wawi API. Standard: `"http://127.0.0.1:5883/api/eazybusiness/"` Die URL wird gezeigt, wenn der JTL Wawi API server gestartet wird. Bei älteren Versionen kann es sein, dass die URL `http://127.0.0.1:5883/api/eazybusiness/v1/` ist.
	- `search mode`: string - steuert wie Artikel ausgewählt werden. Erlaubt: `"category"`, `"supplier"` oder `"none"`, jedoch bringt die suche mit Kategorie oder Hersteller mit der nicht, jedoch kann es sein, dass bei neueren Versionen von JTL-Wawi die suche danach funktioniert.
	- `category id`: string - die id der Kategorie, welche zu dem Vaterartikel zum prüfen hinzugefügt werden soll.
	- `path to folder`: der Pfad zu dem Ordern mit den Bildern.
	- `activate sales channel` bool - wenn `true`, wird der Vaterartikel direkt auf allen sales channeln aktiviert, wird benötigt für die automatische Zuordnung von Kinderartikeln zu den angegeben Variationen.

**Wichtig:**
- Keys sind **case-sensitive** und müssen exakt stimmen (inkl. Leerzeichen).
- JSON erlaubt **keine** Kommentare und **keine** abschließenden Kommata.
- Datei sollte UTF-8 kodiert sein.
- `\`in Pfaden muss immer mit einem weiteren `\` escaped werden.

#### Beispiel Config
```json
{
  "api base url": "http://127.0.0.1:5883/api/eazybusiness/",
  "search mode": "category",
  "category id": "155",
  "path to folder": "C:\\Users\\your-username\\Pictures\\JTL-Wawi-Images",
  "activate sales channel": true
}
```

---

### Anwendung starten
> Man kann die Anwendung entweder durch einen Doppelclick starten oder, wenn man eine alternativen Pfad zur config angeben will dann durch 

```sh
WawiIC.exe -c "D:\\MeineConfigs\\custom.json"
```

---
### Wichtig zu wissen:

- **Search mode funktioniert nicht**: Momentan kann man nur durch die Artikelnummer oder den name suchen, aufgrund eines Fehlers im API Server von JTL Wawi (vllt. gefixt in zukünftigen Versionen)
- **Category/Supplier search mode**: Bei vielen Kategorien, kann das starten der App sehr lange dauern, da es erstmal alle finden muss.
- **Manche Artikel werden nicht gefunden**: Wenn manche Artikel nicht gefunden werden, hilft es manchmal nur nach teilen des Titels oder der Artikel zu suchen, bzw. nach dem vollständigen Name/Artikelnummer
- **Programm/API server hängt sich auf**: Falls das passiert, am besten WawiIC und den API server neu starten.
- **Vaterartikel überprüfen**: Nach dem zusammenfügen von Artikeln sollte die Artikelbeschreibung, die Bilder, die SEO Beschreibung überprüfen, da KI Fehler machen kann.
- **Mehrere Suchen**: Wenn man nach Artikeln sucht, einen oder mehrere auswählt und danach nochmal sucht, sind die vorherigen Artikel immer noch ausgewählt