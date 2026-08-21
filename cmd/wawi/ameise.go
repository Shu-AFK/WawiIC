package wawi

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	"golang.org/x/text/encoding/charmap"
)

// idHeaders are the column names a JTL-Ameise item export uses for the internal
// item id, in the order they are preferred.
var idHeaders = []string{
	"internerschlüssel", "internerschluessel", "internerschlussel",
	"kartikel", "interneid", "artikelid", "itemid", "id",
}

// ParseAmeiseCSV pulls the "Interner Schlüssel" column out of a JTL-Ameise export. Encoding
// and delimiter are configurable in Ameise, so both are detected from the file
// rather than assumed.
func ParseAmeiseCSV(data []byte) ([]int, error) {
	text, err := decodeText(data)
	if err != nil {
		return nil, err
	}

	reader := csv.NewReader(strings.NewReader(text))
	reader.Comma = detectDelimiter(text)
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("could not read the header row: %w", err)
	}

	column, err := findIDColumn(header)
	if err != nil {
		return nil, err
	}

	ids := make([]int, 0)
	seen := make(map[int]struct{})
	row := 1

	for {
		record, err := reader.Read()
		row++
		if err != nil {
			// A malformed row should not cost the rest of the export.
			if err.Error() == "EOF" || len(record) == 0 {
				break
			}
		}
		if column >= len(record) {
			continue
		}

		field := strings.TrimSpace(record[column])
		if field == "" {
			continue
		}

		id, err := strconv.Atoi(field)
		if err != nil {
			return nil, fmt.Errorf("row %d: %q in column %q ist kein Interner Schlüssel", row, field, header[column])
		}
		if id <= 0 {
			return nil, fmt.Errorf("row %d: %d ist kein gültiger Interner Schlüssel", row, id)
		}

		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}

	if len(ids) == 0 {
		return nil, fmt.Errorf("keine Internen Schlüssel in Spalte %q gefunden", header[column])
	}

	return ids, nil
}

func findIDColumn(header []string) (int, error) {
	normalised := make([]string, len(header))
	for i, name := range header {
		normalised[i] = normaliseHeader(name)
	}

	for _, want := range idHeaders {
		for i, name := range normalised {
			if name == want {
				return i, nil
			}
		}
	}

	return 0, fmt.Errorf(
		"keine Spalte mit dem Internen Schlüssel gefunden, gesucht wurde nach %s, die Datei hat: %s",
		strings.Join(idHeaders, ", "),
		strings.Join(header, ", "),
	)
}

func normaliseHeader(name string) string {
	name = strings.TrimPrefix(name, "\ufeff")
	name = strings.ToLower(strings.TrimSpace(name))

	var b strings.Builder
	for _, r := range name {
		switch r {
		case ' ', '.', '-', '_', '/':
			continue
		}
		b.WriteRune(r)
	}

	return b.String()
}

// detectDelimiter picks the separator that occurs most often in the header line.
func detectDelimiter(text string) rune {
	line := text
	if cut := strings.IndexAny(text, "\r\n"); cut >= 0 {
		line = text[:cut]
	}

	best := ';'
	bestCount := strings.Count(line, ";")
	for _, candidate := range []struct {
		sep   rune
		count int
	}{
		{',', strings.Count(line, ",")},
		{'\t', strings.Count(line, "\t")},
		{'|', strings.Count(line, "|")},
	} {
		if candidate.count > bestCount {
			best, bestCount = candidate.sep, candidate.count
		}
	}

	return best
}

// decodeText turns the raw file into UTF-8. Ameise can be configured to write
// UTF-16, UTF-8 or ANSI, and reading one as another yields silent nonsense
// rather than an error, so the byte order mark decides.
func decodeText(data []byte) (string, error) {
	switch {
	case bytes.HasPrefix(data, []byte{0xFF, 0xFE}):
		return decodeUTF16(data[2:], false), nil
	case bytes.HasPrefix(data, []byte{0xFE, 0xFF}):
		return decodeUTF16(data[2:], true), nil
	case bytes.HasPrefix(data, []byte{0xEF, 0xBB, 0xBF}):
		return string(data[3:]), nil
	}

	// UTF-16 without a mark still shows up as every other byte being zero.
	if looksLikeUTF16(data) {
		return decodeUTF16(data, data[0] == 0), nil
	}

	if utf8.Valid(data) {
		return string(data), nil
	}

	decoded, err := charmap.Windows1252.NewDecoder().Bytes(data)
	if err != nil {
		return "", fmt.Errorf("could not decode the file as UTF-8, UTF-16 or Windows-1252: %w", err)
	}

	return string(decoded), nil
}

func looksLikeUTF16(data []byte) bool {
	limit := min(len(data), 512)
	if limit < 4 {
		return false
	}

	zeros := 0
	for _, b := range data[:limit] {
		if b == 0 {
			zeros++
		}
	}

	return zeros*5 > limit
}

func decodeUTF16(data []byte, bigEndian bool) string {
	units := make([]uint16, 0, len(data)/2)
	for i := 0; i+1 < len(data); i += 2 {
		if bigEndian {
			units = append(units, uint16(data[i])<<8|uint16(data[i+1]))
		} else {
			units = append(units, uint16(data[i+1])<<8|uint16(data[i]))
		}
	}

	return string(utf16.Decode(units))
}
