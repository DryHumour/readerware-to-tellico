# Template Customization Guide

In `readerware-to-tellico`, templates are not just an internal implementation mechanism; they are a first-class customization point. The entire pipeline uses Go's standard `text/template` engine to let you audit, clean, restructure, and map your library data during conversion without modifying any Go source code.

This guide explains the template architecture, how to write your own custom overrides, the available configuration functions, and how to utilize the built-in helper functions.

---

## Viewing and Exporting Default Templates

Because all default templates are compiled directly into the tool's binary, the CLI provides dedicated sub-commands to inspect and bootstrap your custom template environments.

### Listing All Default Templates
To see the names of all built-in templates embedded inside the application, use the `list` command:

```bash
readerware-to-tellico templates list
```

### Inspecting a Specific Template
If you want to view the source code of a specific built-in template (e.g., to see how the book author formatting is structured), you can print it directly to `stdout`:

```bash
readerware-to-tellico templates export books.author.gotmpl
```

### Bootstrapping Your Custom Overrides
To export **all** default templates to a local folder so you have a complete workspace of editable starter templates, specify an output directory with the `-o` (or `--output-dir`) flag:

```bash
readerware-to-tellico templates export --output-dir ./my-custom-templates
```

---

## Template Directory Structure

When you provide a custom template directory using the `--template-dirs` flag, the conversion engine loads your custom templates on top of the built-in defaults. If any template name matches, your custom template will **override** the default.

The template loader looks for files ending in `.gotmpl` and organizes them into three main layers:

### 1. Structure Templates
These templates define the core skeleton of the resulting Tellico XML collection:
* `books.header.gotmpl` / `music.header.gotmpl` / `video.header.gotmpl`: Renders the XML headers and collection-level definitions.
* `books.footer.gotmpl` / `music.footer.gotmpl` / `video.footer.gotmpl`: Renders the XML footers and embeds referenced cover images.
* `books.entry.gotmpl` / `music.entry.gotmpl` / `video.entry.gotmpl`: The main orchestrator template executed once for each CSV row to construct the `<entry>` element.

### 2. Configuration Templates
Executed exactly once before CSV processing begins, these files configure the name cleaning, role mapping, category blocklists, and target CSV columns:
* `books.config.gotmpl`
* `music.config.gotmpl`
* `video.config.gotmpl`

### 3. Field-Specific & Cleaning Templates
These templates handle the parsing, cleaning, and formatting of specific XML nodes or CSV columns:
* `<collection_kind>.<field_name>.gotmpl`: Renders individual Tellico XML fields (e.g., `books.title.gotmpl` or `music.label.gotmpl`).
* `clean.<COLUMN_NAME>.gotmpl`: Per-column custom scrubbing templates (e.g., `clean.PUBLISHER.gotmpl`).
* `clean.default.gotmpl`: The fallback cleaning template applied to any CSV columns that do not have a dedicated `clean.*` template.

---

## Configuration Templates (`*.config.gotmpl`)

Configuration templates allow you to define rule sets for name-scrubbing, category pruning, and mapping. They are executed at the start of a conversion run and populate configuration objects via helper methods on the template context.

Below are the key configuration functions you can invoke within your `.config.gotmpl` files:

### Name Normalization Rules
These functions influence the behavior of the name-cleaning and flipping parser (which turns flipped "Last, First" strings back into natural "First Last" order).

* **`.Names.Keep (list <strings>)`**: Keep these specific name strings exactly as-is (case-sensitive). Suitable for "Various Artists" or literal band names with commas.
* **`.Names.Discard (list <strings>)`**: Entirely discard these names (case-insensitive). Useful for noise values like "N/A" or "unknown".
* **`.Names.Scrub (list <strings>)`**: Strip specific noise phrases from within name strings (case-insensitive).
* **`.Names.Audit (dict <trigger> <reason>)`**: Define phrases that automatically trigger a low-confidence audit warning (`rw_requires_audit`) requiring human review.
* **`.Names.Role (dict <phrase> <canonical_role>)`**: Map embedded role annotations (e.g., `"(ed.)"` or `"[translator]"`) to canonical roles so they are extracted out of name strings.
* **`.Names.Marker (dict <phrase> <canonical_marker>)`**: Map name annotations (e.g., `"(signed)"`) to canonical database markers.
* **`.Names.Suffixes (list <strings>)` / `.Names.Honorifics (list <strings>)`**: Inform the name parser of suffixes (e.g., `Jr.`, `Ph.D.`) and honorifics (e.g., `Sir`, `Dr.`) so they are not parsed as surnames.
* **`.Names.Corporate (list <strings>)`**: Identify collective/corporate names (e.g., `Symphony`, `Orchestra`, `Press`, `Band`) that should never be flipped.
* **`.Names.Collaboration (list <strings>)`**: Specify collaborative separators (e.g., `&`, `/`, `featuring`) used to split multi-artist rows.

### Column Mapping Rules
Configure which CSV columns are sent to which parsing engines:

* **`.Columns.Names (dict <role_name> <list_of_columns>)`**: Map CSV columns containing person/group names to specific Tellico roles:
  ```go
  {{- .Columns.Names (dict
      "Authors"      (list "AUTHOR" "AUTHOR2" "AUTHOR3")
      "Editors"      (list "EDITOR")
      "Translators"  (list "TRANSLATOR")
  ) -}}
  ```
* **`.Columns.Markers (list <strings>)`**: Identify columns that may contain special copy markers (such as `TITLE` or `EDITION` containing signs like `[signed]`).
* **`.Columns.Categories (list <strings>)`**: Specify which CSV columns contain classification categories or category paths (e.g., `CATEGORY1`, `CATEGORY2`).

### Category Blocklists
* **`.Blocklist (list <strings>)`**: Prune unwanted category nodes (such as Amazon browse nodes, e.g., `"Books"`, `"Fiction & Literature"`, `"Formats"`) from importing into Tellico's categories.

---

## Data Properties inside Templates

When a structure or field template is executed, it receives a data context object representing the current collection entry.

### 1. Raw Column Values
To pull the cleaned raw string value of any CSV column, use the `.V` method with the column name:
```go
<title>{{ .V "TITLE" | xml }}</title>
```
*Note: `.V` automatically trims leading and trailing whitespaces from the raw column value.*

### 2. Pre-Computed / Normalized Fields
For complex structures (like authors, categories, or ratings) which are pre-scrubbed and aggregated by the Go engine, you can access clean structures via dedicated lazy-loaded methods:

* **`.Authors()`**: Returns a list of clean, normalized author strings.
* **`.Editors()`**: Returns a list of clean, normalized editors.
* **`.Credits()`**: Returns a list of custom parsed credit/role pairings (e.g., illustrators, translators).
* **`.Categories()`**: Returns a list of custom-categorized paths.
* **`.Genres()`**: Returns a list of parsed genre categories.

---

## Template Helper Functions

The engine includes the complete [Sprig](https://masterminds.github.io/sprig/) template library (over 100 functions for string manipulation, math, lists, etc.) alongside several project-specific helper functions:

### XML and HTML Sanitization
* **`xml <string>`**: Escapes characters (`<`, `>`, `&`, `"`, `'`) for safe insertion into Tellico XML.
* **`htmlToText <string>`**: Strips HTML tags and decodes character entities into normal, readable plain text.
* **`squeeze <string>`**: Condenses multiple sequential whitespace characters down to a single space.
* **`paragraphs <string>`**: Wraps lines of text separated by double-newlines into clean block structures.

### Formatting & Normalization
* **`isbn <string>`**: Hyphenates ISBN-10 or ISBN-13 strings using standard regional publisher hyphenation rules.
* **`checkbox <value>`**: Normalizes variations of boolean markers (e.g. `1`, `Y`, `yes`, etc.) into `"true"` or `"false"`.
* **`price <string>`**: Normalizes prices and sanitizes currencies.
* **`rating <string>`**: Normalizes numeric rating scales down to standard Tellico 1-to-5 star values.
* **`dimensions <string>`**: Standardizes physical size/dimension values.

### Template Composition
* **`required <error_msg> <value>`**: Stops the conversion process with a user-friendly error message if the passed value is empty or nil.
* **`include <template_name> <context>`**: Renders another template inside the current template, allowing you to pipe the output:
  ```go
  {{ include "books.title" . | indent 4 }}
  ```
* **`tpl <template_string> <context>`**: Compiles and executes an inline template string dynamically.

---

## Example Override

To customize how book publishers are imported into Tellico, you can create a custom `books.publisher.gotmpl` file inside your custom template directory:

```go
{{- define "books.publisher" -}}
  {{- with .V "PUBLISHER" -}}
    <publisher>{{ . | htmlToText | squeeze | xml }}</publisher>
  {{- end -}}
{{- end -}}
```

When you run:
```bash
readerware-to-tellico books --template-dirs /path/to/my/templates
```
The engine will transparently use your custom publisher formatting while keeping the rest of the default templates intact!
