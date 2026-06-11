package convert

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"reflect"
	"regexp"
	"strings"
	"text/template"

	"github.com/DryHumour/readerware-to-tellico/internal/strutil"
	"github.com/DryHumour/readerware-to-tellico/internal/tellico/collection"
	"github.com/Masterminds/sprig/v3"
)

const (
	productInfoPattern = `` + // (for gofmt)
		(`(?:editorial\s+|starred\s+)?reviews?|` +
			`synopsis|` +
			`(?:product\s+|book\s+)?description|` +
			`about\s+the\s+(?:author|book)|` +
			`No\s+description\s+for\s+this\s+item\s+yet|` +
			`All\s+products\s+are\s+BRAND\s+NEW\s+and\s+factory\s+sealed|` +
			`Fast\s+shipping\s+and\s+100%\s+Satisfaction\s+Guaranteed|` +
			(`from\s+(?:` +
				`amazon\.(?:co\.)?[a-z]+|` +
				`library\s+journal|` +
				`our\s+editors|` +
				`publishers\s+weekly|` +
				`the\s+(?:publisher|back\s+cover|inside\s+flap)|` +
				`kirkus\s+reviews|` +
				`school\s+library\s+journal\s+(?:(?:grade\s+\d+|kindergarten)(?:\s+up)?|starred\s+review|adult)` +
				`)`))
)

var (
	// productInfoCleanRE is an anchored search used to strip matched prefixes.
	// It specifically allows "Amazon.*" prefixes before the core pattern.
	productInfoCleanRE = regexp.MustCompile(`^\s*(?i:(?:amazon\.(?:co\.)?[a-z]+|` + productInfoPattern + `)[:.]*(?:\s*$|\s+)?)+`)

	// productInfoAuditRE is an unanchored search used to flag records containing the pattern anywhere.
	productInfoAuditRE = regexp.MustCompile(`\b(?i:` + productInfoPattern + `)\b`)
)

// func init() { fmt.Println(productInfoCleanRE) }

var (
	// baseFuncMap contains the default template functions for our problem domain.
	baseFuncMap = template.FuncMap{
		"checkbox":     strutil.Checkbox,
		"dimensions":   strutil.Dimensions,
		"htmlToText":   strutil.HTMLToText,
		"isbn":         strutil.HyphenateISBN,
		"keywords":     strutil.Keywords,
		"paragraphs":   strutil.Paragraphs,
		"price":        strutil.Price,
		"product_info": productInfoFunc,
		"rating":       strutil.Rating1to5,
		"squeeze":      strutil.Squeeze,
		"xml":          strutil.XMLEscape,
	}

	//go:embed templates/*.gotmpl
	TemplatesFS embed.FS
)

// Base returns a new template with the base functions registered and the built-in templates parsed.
//
// [Sprig]: https://masterminds.github.io/sprig/
func Base(extraFuncs template.FuncMap) (*template.Template, error) {
	t := template.New("base")
	return t.
		Funcs(template.FuncMap{
			"required": requiredFunc,
			"include":  includeFunc(t),
			"tpl":      tplFunc(t),
		}).
		Funcs(sprig.TxtFuncMap()).
		Funcs(baseFuncMap).
		Funcs(extraFuncs).
		ParseFS(TemplatesFS, "templates/*.gotmpl")
}

// LoadAll loads all templates from the given filesystem.
// Template naming convention (books, music, video):
//   - books.<field>.gotmpl: Entry-level field templates (e.g., books.title.gotmpl)
//   - books.header.gotmpl: XML header with collection-level fields
//   - books.footer.gotmpl: XML footer with <images> section
//   - books.entry.gotmpl: Entry wrapper template
//   - books.config.gotmpl: Configuration for template functions e.g. normalize_name
//   - books.audit.gotmpl: Auditing of entries
//   - clean.<COLUMN>.gotmpl: Per-column cleaning templates
//   - clean.default.gotmpl: Fallback cleaning template
func LoadAll(t *template.Template, fsys fs.FS) (*template.Template, error) {
	return Load(t, fsys, "*.gotmpl")
}

// Load loads templates from the given filesystem with the given patterns.
func Load(t *template.Template, fsys fs.FS, pattern ...string) (*template.Template, error) {
	return t.ParseFS(fsys, pattern...)
}

// requiredFunc halts template execution and returns an error if the value is empty/nil.
func requiredFunc(warn string, val interface{}) (interface{}, error) {
	if val == nil {
		return nil, errors.New(warn)
	}

	// Fast-path for the most common empty string check
	if s, ok := val.(string); ok && s == "" {
		return nil, errors.New(warn)
	}

	// Catch-all for empty slices, maps, booleans (false), and zero-values
	v := reflect.ValueOf(val)
	if v.IsValid() && v.IsZero() {
		return nil, errors.New(warn)
	}

	return val, nil
}

// includeFunc renders a named template and returns it as a string, allowing it to be piped.
// e.g., {{ include "my-template" . | indent 4 }}
func includeFunc(t *template.Template) func(name string, data interface{}) (string, error) {
	return func(name string, data interface{}) (string, error) {
		var b strings.Builder
		// We execute the named template into our builder.
		// If the template doesn't exist, ExecuteTemplate returns a clean error.
		err := t.ExecuteTemplate(&b, name, data)
		if err != nil {
			return "", err
		}
		return b.String(), nil
	}
}

// tplFunc takes a raw string containing template syntax, compiles it on the fly,
// and executes it using the provided data context.
func tplFunc(t *template.Template) func(tplString string, data interface{}) (string, error) {
	return func(tplString string, data interface{}) (string, error) {
		// CRITICAL: We must Clone() the root template.
		// This ensures the inline template inherits all other parsed templates
		// and the FuncMap, but doesn't permanently pollute the root engine.
		cloned, err := t.Clone()
		if err != nil {
			return "", fmt.Errorf("failed to clone template engine: %w", err)
		}

		// Parse the inline string
		inline, err := cloned.New("inline-tpl").Parse(tplString)
		if err != nil {
			return "", fmt.Errorf("failed to parse inline template: %w", err)
		}

		var b strings.Builder
		if err := inline.Execute(&b, data); err != nil {
			return "", err
		}

		return b.String(), nil
	}
}

func productInfoFunc(auditor collection.Auditor, productInfo string) string {
	if productInfo == "" {
		return ""
	}
	clean := productInfoCleanRE.ReplaceAllLiteralString(productInfo, "")
	if locs := productInfoAuditRE.FindStringIndex(clean); locs != nil {
		return productInfo // too much risk of data loss
	}
	return clean
}
