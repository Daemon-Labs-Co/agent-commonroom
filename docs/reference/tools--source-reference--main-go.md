# tools/source-reference/main.go

[Module index](tools.md) · [Open source](../../tools/source-reference/main.go)

Go implementation for main; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`main`** — main regenerates the reference in the current checkout and reports failures.
  Signature: `func()`.
  Calls: `fmt.Fprintln`, `generate`, `os.Exit`.
- **`clean`** — clean renders source comments as safe, compact Markdown prose.
  Signature: `func(s string) string`.
  Calls: `strings.Fields`, `strings.Join`, `strings.NewReplacer`, `strings.NewReplacer("|", "/", "'", "'", "[", "(", "]", ")").Replace`.
- **`code`** — code formats a declaration while retaining Go syntax such as slice brackets.
  Signature: `func(fset *token.FileSet, n ast.Node) string`.
  Calls: `b.String`, `format.Node`, `strings.Fields`, `strings.Join`, `strings.ReplaceAll`.
- **`purpose`** — purpose classifies a file by its role and uses document titles where available.
  Signature: `func(path string, data []byte) string`.
  Calls: `clean`, `filepath.Base`, `filepath.Ext`, `string`, `strings.Contains`, `strings.HasPrefix`, `strings.HasSuffix`, `strings.NewReplacer`, `strings.NewReplacer("_test.go", "", ".test.cjs", "", "_test.py", "").Replace`, `strings.ReplaceAll`, `strings.Split`, `strings.TrimPrefix`, `strings.TrimSuffix`.
- **`describeGo`** — describeGo indexes type declarations, functions, receiver methods and direct calls.
  Signature: `func(path string, data []byte) (string, error)`.
  Calls: `append`, `ast.Inspect`, `b.String`, `clean`, `code`, `d.Doc.Text`, `fmt.Fprintf`, `len`, `parser.ParseFile`, `sort.Strings`, `strings.HasPrefix`, `strings.Join`, `t.Doc.Text`, `token.NewFileSet`.
- **`generate`** — generate walks owned source files, omits private/generated state and writes reference pages.
  Signature: `func(root string) error`.
  Calls: `append`, `b.String`, `b.WriteString`, `d.IsDir`, `d.Name`, `d.Type`, `describeGo`, `filepath.Join`, `filepath.Rel`, `filepath.ToSlash`, `filepath.WalkDir`, `fmt.Fprintf`, `fmt.Sprintf`, `jsSymbol.FindAllStringSubmatch`, `len`, `os.MkdirAll`, `os.ReadFile`, `os.WriteFile`, `purpose`, `sort.Strings`, `string`, `strings.Contains`, `strings.HasPrefix`, `strings.HasSuffix`, `strings.NewReplacer`, `strings.NewReplacer("/", "--", ".", "-").Replace`, `strings.Split`, `strings.TrimPrefix`.
