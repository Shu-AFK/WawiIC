package wawi

import "testing"

func TestParseItemIDs(t *testing.T) {
	in := "\ufeff# Vaterartikel\n1234\n\n5678\n1234\n" +
		"9012,\"ABC-1\",Testartikel\n" +
		"3456;XYZ-2\n" +
		"  7890  \n"

	ids, err := ParseItemIDs(in)
	if err != nil {
		t.Fatalf("ParseItemIDs() error = %v", err)
	}

	want := []int{1234, 5678, 9012, 3456, 7890}
	if len(ids) != len(want) {
		t.Fatalf("ParseItemIDs() = %v, want %v", ids, want)
	}
	for i := range want {
		if ids[i] != want[i] {
			t.Errorf("ids[%d] = %d, want %d", i, ids[i], want[i])
		}
	}
}

func TestParseItemIDsRejectsBadInput(t *testing.T) {
	tests := map[string]string{
		"not a number": "1234\nABC-123\n",
		"zero":         "0\n",
		"negative":     "-5\n",
		"empty":        "# nur Kommentare\n\n",
	}

	for name, in := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseItemIDs(in); err == nil {
				t.Error("ParseItemIDs() error = nil, want an error")
			}
		})
	}
}
