## High Priority
- [x] Fix non-deterministic output ordering caused by map iteration (breaks diffing/reproducibility of generated tellico.xml):
    - `rowData.Categories()` iterates `map[string]bool` (row_data.go)
    - `rowData.Genres()` returns `slices.Collect(maps.Keys(uniq))` (row_data.go)
    - `aggregation.Markers()` returns `slices.Collect(maps.Keys(...))` (aggregation.go)
    - `AggregateNames` iterates `columns.Names` map; cross-role credit order varies (aggregate.go)
- [x] Stop integration tests from hitting the real network: `convert_integration_test.go` uses `NewConverter` without `WithHTTPClient`, so `isbn.LoadHyphenator` runs against `http.DefaultClient` (no timeout). Inject a stub client or short-circuit.
- [ ] Write output atomically: `os.Create` truncates the `.tc` before input is even validated, leaving a corrupt partial zip on failure. Write to a temp file and rename on success (converter.go).
- [x] Fix env var mismatch for `java-path`: `bindFlags` binds nested commands under `extract.java-path` (env `RW2TC_EXTRACT_JAVA_PATH`), but the error message in cmd/images.go tells users to set `RW2TC_IMAGES_EXTRACT_JAVA_PATH`. Use full command ancestry for the namespace or fix the message.
- [x] Fix `buffered_file.go`: `Peek(3)` fails on files < 3 bytes, so an empty CSV reports "failed to peek file for BOM: EOF" instead of `ErrEmptyInputFile`.
- [x] Fix `csv get` silently truncating on parse errors: the read loop breaks on any error, not just `io.EOF` (cmd/csv.go).
- [x] Stop joining `io.EOF` into `NewFileEmptyError` — makes `errors.Is(err, io.EOF)` true for a semantic error (copier/errors.go).
- [ ] Fix misleading JRE help text in `images extract`: any non-`exec.ExitError` failure (e.g. db path validation) gets the "requires a JRE" blurb appended (cmd/images.go).
- [x] Report real CSV line numbers in row errors: `lineNumber` counts records, not file lines; multi-line fields drift. Use `csv.Reader.FieldPos(0)` (converter_convert.go).

## Medium Priority
- [ ] Add descriptions to all Readerware fields in Tellico
- [ ] deal with year-only Tellico dates
- [ ] Deduplicate cmd/books.go, cmd/music.go, cmd/video.go (~100 lines each of identical flag registration + run logic) behind a shared helper
- [ ] Deduplicate `ConfigureHeaders` across BooksPolicy/MusicPolicy/VideoPolicy (identical implementations; move to shared code)
- [ ] Reconsider `panic` in `extract()` for canonical-map misses (FIXME already notes this; normalize/common.go)
- [ ] Implement `MusicEntry.Labels()` properly (currently a pass-through placeholder with FIXME; music_entry.go)
- [ ] Document the implicit contract that rendering entry templates mutates the image index (`ManifestEntry.Use()`/`IsUsed` drives the footer manifest) — will break if rendering is ever parallelized
- [ ] Fix build-from-fresh-clone: `ImageDumper.class`/`hsqldb.jar` are gitignored `go:embed` targets produced by `go generate` (requires JDK). Document in README or commit generated artifacts.

## Low Priority
- [ ] Ensure that slog is handled in all test cases
- [ ] Add support for loans
- [ ] Remove commented-out debug `fmt.Printf` lines in collection_info.go
- [x] Remove `"extract"` from `extractCmd`'s own alias list (cmd/images.go)
- [ ] Add lint/staticcheck step to CI (currently only build + `go test -race`)
- [ ] Document that `--concurrency 0` selects the simple (sequential) copier
- [ ] Consider hoisting the validator+translator construction in `config.Validate()` out of the per-call path
- [ ] Unify the two BOM-strip implementations (cmd/csv.go vs convert/buffered_file.go)
