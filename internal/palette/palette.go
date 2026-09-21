package palette

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
)

const DefaultName = "catppuccin_mocha"

// Definition is the palette JSON format used by the repository assets.
type Definition struct {
	Colors     []string `json:"palette"`
	Reset      string   `json:"reset"`
	YellowBold string   `json:"yellow_bold"`
	DimRed     string   `json:"dim_red"`
}

// Catalog is the validated set of named palettes embedded in the binary.
type Catalog struct {
	definitions map[string]Definition
}

// Load reads and validates named palettes from an fs.FS.
func Load(assets fs.FS) (Catalog, error) {
	entries, err := fs.ReadDir(assets, "palettes")
	if err != nil {
		return Catalog{}, fmt.Errorf("read palettes: %w", err)
	}

	definitions := make(map[string]Definition)
	for _, entry := range entries {
		if entry.IsDir() || path.Ext(entry.Name()) != ".json" {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), path.Ext(entry.Name()))
		if reservedName(name) {
			continue
		}
		data, err := fs.ReadFile(assets, path.Join("palettes", entry.Name()))
		if err != nil {
			return Catalog{}, fmt.Errorf("read palette %q: %w", name, err)
		}
		var definition Definition
		if err := json.Unmarshal(data, &definition); err != nil {
			return Catalog{}, fmt.Errorf("parse palette %q: %w", name, err)
		}
		if len(definition.Colors) == 0 {
			return Catalog{}, fmt.Errorf("palette %q has no colors", name)
		}
		definitions[name] = definition
	}
	if len(definitions) == 0 {
		return Catalog{}, fmt.Errorf("no palettes found")
	}
	return Catalog{definitions: definitions}, nil
}

// Names returns palette names in deterministic order.
func (c Catalog) Names() []string {
	names := make([]string, 0, len(c.definitions))
	for name := range c.definitions {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Get resolves a palette by its stable name.
func (c Catalog) Get(name string) (Definition, error) {
	definition, ok := c.definitions[name]
	if !ok {
		return Definition{}, fmt.Errorf("unknown palette %q", name)
	}
	return definition, nil
}

func reservedName(name string) bool {
	switch name {
	case "palette", "cat_backup", "test_palette":
		return true
	default:
		return false
	}
}
