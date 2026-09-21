package palette_test

import (
	"testing"

	"testing/fstest"

	"github.com/averyfreeman/ansipath/internal/palette"
)

func TestLoadCatalogListsAndResolvesNamedPalettes(t *testing.T) {
	assets := fstest.MapFS{
		"palettes/dracula.json": {Data: []byte(`{
            "palette": ["RED", "BLUE"],
            "reset": "RESET",
            "yellow_bold": "YELLOW",
            "dim_red": "DIMRED"
        }`)},
		"palettes/ayu_dark.json": {Data: []byte(`{
            "palette": ["CYAN"],
            "reset": "RESET",
            "yellow_bold": "YELLOW",
            "dim_red": "DIMRED"
        }`)},
	}

	catalog, err := palette.Load(assets)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	wantNames := []string{"ayu_dark", "dracula"}
	gotNames := catalog.Names()
	if len(gotNames) != len(wantNames) {
		t.Fatalf("Names returned %v, want %v", gotNames, wantNames)
	}
	for i := range wantNames {
		if gotNames[i] != wantNames[i] {
			t.Errorf("Names[%d] = %q, want %q", i, gotNames[i], wantNames[i])
		}
	}

	got, err := catalog.Get("dracula")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if len(got.Colors) != 2 || got.Colors[0] != "RED" {
		t.Fatalf("Get returned %#v", got)
	}
}

func TestCatalogUsesCanonicalDefault(t *testing.T) {
	if palette.DefaultName != "catppuccin_mocha" {
		t.Fatalf("DefaultName = %q", palette.DefaultName)
	}
}
