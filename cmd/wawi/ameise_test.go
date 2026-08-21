package wawi

import (
	"bytes"
	"testing"
	"unicode/utf16"
)

func toUTF16LE(s string) []byte {
	out := []byte{0xFF, 0xFE}
	for _, u := range utf16.Encode([]rune(s)) {
		out = append(out, byte(u), byte(u>>8))
	}
	return out
}

func toWindows1252(s string) []byte {
	// The characters used here map one to one onto Windows-1252.
	out := make([]byte, 0, len(s))
	for _, r := range s {
		out = append(out, byte(r))
	}
	return out
}

func TestParseAmeiseCSV(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want []int
	}{
		{
			name: "semicolon with utf-8 bom",
			data: append([]byte{0xEF, 0xBB, 0xBF},
				[]byte("kArtikel;Name;Preis\r\n1234;Stuhl;19,99\r\n5678;Tisch;29,99\r\n")...),
			want: []int{1234, 5678},
		},
		{
			name: "utf-16 le as ameise writes it",
			data: toUTF16LE("kArtikel;Name\r\n1234;Stuhl\r\n5678;Tisch\r\n"),
			want: []int{1234, 5678},
		},
		{
			name: "ansi with umlauts",
			data: toWindows1252("Interne ID;Name\r\n1234;Küche\r\n5678;Größe\r\n"),
			want: []int{1234, 5678},
		},
		{
			name: "interner schlüssel as ameise names it",
			data: []byte("Interner Schlüssel;Artikelnummer;Name\r\n1234;VK-1;Stuhl\r\n5678;VK-2;Tisch\r\n"),
			want: []int{1234, 5678},
		},
		{
			name: "interner schluessel written out",
			data: toWindows1252("Interner Schluessel;Name\r\n1234;Stuhl\r\n5678;Tisch\r\n"),
			want: []int{1234, 5678},
		},
		{
			name: "comma separated with quotes",
			data: []byte("\"kArtikel\",\"Name\"\n\"1234\",\"Stuhl, rot\"\n\"5678\",\"Tisch\"\n"),
			want: []int{1234, 5678},
		},
		{
			name: "id column is not the first",
			data: []byte("Artikelnummer;Name;Artikel-ID\r\nVK-1;Stuhl;1234\r\nVK-2;Tisch;5678\r\n"),
			want: []int{1234, 5678},
		},
		{
			name: "blank values and duplicates are dropped",
			data: []byte("kArtikel;Name\r\n1234;a\r\n;b\r\n1234;c\r\n5678;d\r\n"),
			want: []int{1234, 5678},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAmeiseCSV(tt.data)
			if err != nil {
				t.Fatalf("ParseAmeiseCSV() error = %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("ParseAmeiseCSV() = %v, want %v", got, tt.want)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("[%d] = %d, want %d", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestParseAmeiseCSVReportsUnknownHeader(t *testing.T) {
	_, err := ParseAmeiseCSV([]byte("Nummer;Name\r\n1234;Stuhl\r\n"))
	if err == nil {
		t.Fatal("ParseAmeiseCSV() error = nil, want an error")
	}
	// The message has to name the columns that are actually in the file.
	if !bytes.Contains([]byte(err.Error()), []byte("Nummer")) {
		t.Errorf("error %q does not list the columns found", err)
	}
}

// A file where the id column holds article numbers must fail loudly rather than
// silently produce nothing.
func TestParseAmeiseCSVRejectsNonNumericID(t *testing.T) {
	_, err := ParseAmeiseCSV([]byte("kArtikel;Name\r\n1234;Stuhl\r\nVK-2;Tisch\r\n"))
	if err == nil {
		t.Fatal("ParseAmeiseCSV() error = nil, want an error")
	}
	if !bytes.Contains([]byte(err.Error()), []byte("VK-2")) {
		t.Errorf("error %q does not name the offending value", err)
	}
}
