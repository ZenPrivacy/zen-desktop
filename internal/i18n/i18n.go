// Package i18n translates the text that Go renders for the user, such as the
// block page and the proxy error page.
//
// The catalogues in locales/ are separate from the frontend's because go:embed
// can't reach files outside this package's directory.
package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"sync/atomic"
)

const fallbackLocale = "en-US"

// bareLocales maps the bare languages that the frontend may store in
// Config.Locale to a catalogue. It must match the bare-language entries in
// the frontend's resources table in frontend/src/i18n/index.ts.
var bareLocales = map[string]string{
	"en": "en-US",
	"de": "de-DE",
	"fr": "fr-FR",
	"it": "it-IT",
	"ja": "ja-JP",
	"kk": "kk-KZ",
	"ru": "ru-RU",
	"tr": "tr-TR",
	"vi": "vi-VN",
	"zh": "zh-CN",
}

//go:embed locales/*.json
var localeFiles embed.FS

var catalogues = loadCatalogues()

var current atomic.Pointer[Catalogue]

func init() {
	current.Store(catalogues[fallbackLocale])
}

// Catalogue holds the translations for one locale.
type Catalogue struct {
	locale   string
	messages map[string]string
}

// T returns the translation for key. It falls back to en-US if the key is
// missing, and to the key itself if en-US doesn't have it either, so that a gap
// shows up on the page rather than as a blank.
func (c *Catalogue) T(key string) string {
	if s, ok := c.messages[key]; ok {
		return s
	}
	if s, ok := catalogues[fallbackLocale].messages[key]; ok {
		return s
	}
	return key
}

// Lang returns the catalogue's locale tag, for the lang attribute of a page.
func (c *Catalogue) Lang() string {
	return c.locale
}

// SetLocale sets the catalogue returned by Current to the one named locale, or
// the one bareLocales maps it to, or en-US.
func SetLocale(locale string) {
	current.Store(resolve(locale))
}

// Current returns the catalogue for the locale the user picked. Call it once
// per rendered page so that a locale change can't mix two languages on one page.
func Current() *Catalogue {
	return current.Load()
}

func resolve(locale string) *Catalogue {
	if full, ok := bareLocales[locale]; ok {
		locale = full
	}
	if c, ok := catalogues[locale]; ok {
		return c
	}
	return catalogues[fallbackLocale]
}

// loadCatalogues parses the embedded catalogues. It panics on a malformed file:
// the files are embedded at build time, and the tests catch a bad one before then.
func loadCatalogues() map[string]*Catalogue {
	entries, err := localeFiles.ReadDir("locales")
	if err != nil {
		panic(fmt.Sprintf("read locales: %v", err))
	}

	loaded := make(map[string]*Catalogue, len(entries))
	for _, entry := range entries {
		data, err := localeFiles.ReadFile(path.Join("locales", entry.Name()))
		if err != nil {
			panic(fmt.Sprintf("read %s: %v", entry.Name(), err))
		}
		var tree map[string]any
		if err := json.Unmarshal(data, &tree); err != nil {
			panic(fmt.Sprintf("parse %s: %v", entry.Name(), err))
		}

		locale := strings.TrimSuffix(entry.Name(), ".json")
		messages := make(map[string]string)
		if err := flatten("", tree, messages); err != nil {
			panic(fmt.Sprintf("parse %s: %v", entry.Name(), err))
		}
		loaded[locale] = &Catalogue{locale: locale, messages: messages}
	}

	if _, ok := loaded[fallbackLocale]; !ok {
		panic(fallbackLocale + " catalogue is missing")
	}
	return loaded
}

// flatten turns nested JSON objects into dotted keys, so that
// {"blockPage": {"title": "..."}} becomes "blockPage.title".
func flatten(prefix string, tree map[string]any, out map[string]string) error {
	for k, v := range tree {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}
		switch v := v.(type) {
		case string:
			out[key] = v
		case map[string]any:
			if err := flatten(key, v, out); err != nil {
				return err
			}
		default:
			return fmt.Errorf("%s: want a string or an object, got %T", key, v)
		}
	}
	return nil
}
