// Command source-reference builds an offline map of repository files and Go APIs.
package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var omitted = map[string]bool{".git": true, ".data": true, ".worktrees": true, "bin": true, "coverage": true, "secrets": true, ".secret": true, ".secrets": true, "credentials": true, ".ssh": true, ".aws": true, ".azure": true, "node_modules": true, "vendor": true, "__pycache__": true}

// privateNames are omitted even when they are not covered by an ignored directory.
var privateNames = map[string]bool{".pgpass": true, ".pg_service.conf": true, ".npmrc": true, ".netrc": true, ".pypirc": true, "credentials.json": true, "local-config.json": true, "application_default_credentials.json": true}

var jsSymbol = regexp.MustCompile(`(?m)(?:^|[;{}]\s*)(?:export\s+)?(?:async\s+)?function\s+([A-Za-z_$][\w$]*)\s*\(([^)]*)\)|(?:^|[;{}]\s*)(?:const|let|var)\s+([A-Za-z_$][\w$]*)\s*=\s*(?:async\s*)?(?:\([^)]*\)|[A-Za-z_$][\w$]*)\s*=>|\bclass\s+([A-Za-z_$][\w$]*)`)

// main regenerates the reference in the current checkout and reports failures.
func main() {
	if err := generate("."); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// clean renders source comments as safe, compact Markdown prose.
func clean(s string) string {
	return strings.Join(strings.Fields(strings.NewReplacer("|", "/", "`", "'", "[", "(", "]", ")").Replace(s)), " ")
}

// code formats a declaration while retaining Go syntax such as slice brackets.
func code(fset *token.FileSet, n ast.Node) string {
	var b bytes.Buffer
	_ = format.Node(&b, fset, n)
	return strings.Join(strings.Fields(strings.ReplaceAll(b.String(), "`", "'")), " ")
}

// purpose classifies a file by its role and uses document titles where available.
func purpose(path string, data []byte) string {
	ext := filepath.Ext(path)
	base := filepath.Base(path)
	if strings.HasSuffix(base, "_test.go") || strings.Contains(base, ".test.") || strings.HasSuffix(base, "_test.py") {
		return "Automated regression checks for " + strings.NewReplacer("_test.go", "", ".test.cjs", "", "_test.py", "").Replace(base) + "; read test names for individual contracts."
	}
	if ext == ".md" {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "# ") {
				return clean(strings.TrimPrefix(line, "# ")) + " — documentation."
			}
		}
	}
	if strings.HasPrefix(path, ".agents/") || strings.HasPrefix(path, ".claude/") {
		return "Developer-assistant skill or integration configuration; not a runtime application plugin."
	}
	switch ext {
	case ".svg", ".png", ".jpg", ".jpeg", ".ico":
		return "Presentation asset; consumed by the interface or documentation, with no executable methods."
	case ".go":
		return "Go implementation for " + strings.ReplaceAll(strings.TrimSuffix(base, ext), "_", " ") + "; declarations and calls are listed below."
	case ".js", ".cjs":
		return "JavaScript implementation for " + strings.ReplaceAll(strings.TrimSuffix(base, ext), "-", " ") + "; named functions are indexed below."
	case ".css":
		return "Visual styles for " + strings.TrimSuffix(base, ext) + "; selectors define presentation rather than application methods."
	case ".html":
		return "HTML entry point or static demonstration; loads the associated interface scripts and styles."
	case ".sql":
		return "Explicit database schema/migration SQL; apply only with the separate migration owner."
	case ".sh", ".ps1":
		return "Operator/developer automation; inspect commands and prerequisites before running."
	case ".json", ".yaml", ".yml", ".toml":
		return "Declarative configuration or dependency metadata; no class methods."
	}
	switch base {
	case "go.mod":
		return "Go module identity, minimum toolchain and dependency requirements."
	case "go.sum":
		return "Go dependency integrity checksums."
	case "Makefile":
		return "Build, verification, coverage and secret-check entry points."
	case "Dockerfile":
		return "Container image build instructions; part of the isolation boundary."
	case "LICENSE":
		return "Project redistribution and licensing terms."
	case ".env.example":
		return "Non-secret environment configuration examples; not automatically loaded by the controller."
	case ".gitignore":
		return "Version-control exclusions for private and generated local files."
	case ".dockerignore":
		return "Excludes private/unnecessary files from container build contexts."
	case "pre-commit":
		return "Staged-content secret scanner hook; does not create or sign commits."
	}
	return "Repository support file: " + clean(base) + "; follow the source link for its exact contents."
}

// describeGo indexes type declarations, functions, receiver methods and direct calls.
func describeGo(path string, data []byte) (string, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, data, parser.ParseComments)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Package: `%s`. Go uses types, structs, interfaces and receiver methods rather than classes.\n\n", f.Name.Name)
	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				if t, ok := spec.(*ast.TypeSpec); ok {
					doc := ""
					if t.Doc != nil {
						doc = t.Doc.Text()
					} else if d.Doc != nil {
						doc = d.Doc.Text()
					}
					if doc == "" {
						switch t.Type.(type) {
						case *ast.StructType:
							doc = "Structured data type; fields are shown in the declaration."
						case *ast.InterfaceType:
							doc = "Behavior contract; required method signatures are shown in the declaration."
						default:
							doc = "Named data type; underlying representation is shown in the declaration."
						}
					}
					fmt.Fprintf(&b, "- **Type `%s`**: %s\n", t.Name, clean(doc))
					fmt.Fprintf(&b, "  Declaration: `%s`.\n", code(fset, t.Type))
				}
			}
		case *ast.FuncDecl:
			name := d.Name.Name
			if d.Recv != nil {
				name = "(" + code(fset, d.Recv.List[0].Type) + ")." + name
			}
			sig := code(fset, d.Type)
			doc := ""
			if d.Doc != nil {
				doc = clean(d.Doc.Text())
			}
			if doc == "" {
				if strings.HasPrefix(d.Name.Name, "Test") {
					doc = "Regression contract: " + d.Name.Name + "."
				} else {
					doc = "Source has no prose contract; use the signature and direct calls below to trace behavior."
				}
			}
			calls := map[string]bool{}
			if d.Body != nil {
				ast.Inspect(d.Body, func(n ast.Node) bool {
					if c, ok := n.(*ast.CallExpr); ok {
						switch c.Fun.(type) {
						case *ast.Ident, *ast.SelectorExpr:
							calls[code(fset, c.Fun)] = true
						}
					}
					return true
				})
			}
			names := []string{}
			for s := range calls {
				names = append(names, s)
			}
			sort.Strings(names)
			fmt.Fprintf(&b, "- **`%s`** — %s\n  Signature: `%s`.\n", name, doc, sig)
			if len(names) > 0 {
				fmt.Fprintf(&b, "  Calls: `%s`.\n", strings.Join(names, "`, `"))
			}
		}
	}
	return b.String(), nil
}

// generate walks owned source files, omits private/generated state and writes reference pages.
func generate(root string) error {
	groups := map[string][]string{}
	dataByPath := map[string][]byte{}
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, e := filepath.Rel(root, p)
		if e != nil {
			return e
		}
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			if omitted[d.Name()] || rel == "docs/technical/reference" {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}
		base := d.Name()
		if privateNames[base] || strings.HasPrefix(base, "credentials-") || strings.HasPrefix(base, "service-account") || strings.HasPrefix(base, "service_account") || strings.HasPrefix(base, "id_rsa") || strings.HasPrefix(base, "id_ed25519") || base == ".DS_Store" || base == "settings.local.json" || (strings.HasPrefix(base, ".env") && base != ".env.example") || strings.HasSuffix(base, ".pem") || strings.HasSuffix(base, ".key") || strings.HasSuffix(base, ".log") {
			return nil
		}
		data, e := os.ReadFile(p)
		if e != nil {
			return e
		}
		group := strings.Split(rel, "/")[0]
		if !strings.Contains(rel, "/") {
			group = "root"
		}
		groups[group] = append(groups[group], rel)
		dataByPath[rel] = data
		return nil
	})
	if err != nil {
		return err
	}
	out := filepath.Join(root, "docs/technical/reference")
	if err = os.MkdirAll(out, 0755); err != nil {
		return err
	}
	keys := []string{}
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	index := "# Repository file and API reference\n\nGenerated from the local checkout with `go run ./tools/source-reference`. Regenerate after changing files.\n\nEvery included file has a source link and a purpose. Go entries include every top-level function, receiver method and type, including struct fields and interface methods. Existing comments supply contracts; direct calls expose implementation dependencies. Where a declaration has no prose contract, the reference says so rather than inventing behavior. JavaScript entries index named function declarations, arrow bindings and classes; anonymous callbacks and dynamically assigned properties must be read in the linked source. This is a source map, not a claim that every internal method has a reviewed narrative contract.\n\n"
	index += "## Modules\n\n"
	total := 0
	for _, group := range keys {
		files := groups[group]
		sort.Strings(files)
		filename := strings.TrimPrefix(group, ".") + ".md"
		total += len(files)
		var b strings.Builder
		fmt.Fprintf(&b, "# %s files\n\n[Reference index](README.md)\n\n", group)
		for _, path := range files {
			data := dataByPath[path]
			fmt.Fprintf(&b, "## %s\n\n[Open source](../../../%s)\n\n%s\n\n", path, path, purpose(path, data))
			if strings.HasSuffix(path, ".go") {
				s, e := describeGo(path, data)
				if e != nil {
					return e
				}
				detailName := strings.NewReplacer("/", "--", ".", "-").Replace(path) + ".md"
				detail := fmt.Sprintf("# %s\n\n[Module index](%s) · [Open source](../../../%s)\n\n%s\n\n%s", path, filename, path, purpose(path, data), s)
				if e := os.WriteFile(filepath.Join(out, detailName), []byte(strings.TrimRight(detail, "\n")+"\n"), 0644); e != nil {
					return e
				}
				fmt.Fprintf(&b, "[Types, methods and functions](%s)\n\n", detailName)
			} else if strings.HasSuffix(path, ".js") || strings.HasSuffix(path, ".cjs") {
				matches := jsSymbol.FindAllStringSubmatch(string(data), -1)
				for _, m := range matches {
					name := m[1]
					kind := "function"
					if name == "" {
						name = m[3]
						kind = "arrow function"
					}
					if name == "" {
						name = m[4]
						kind = "class"
					}
					fmt.Fprintf(&b, "- `%s`: named %s; inspect the source for parameters, callbacks and effects.\n", name, kind)
				}
				if len(matches) == 0 {
					b.WriteString("No named function/class declarations matched; this may be a script, test suite or callback-based module.\n")
				}
				b.WriteString("\n")
			}
		}
		if err = os.WriteFile(filepath.Join(out, filename), []byte(strings.TrimRight(b.String(), "\n")+"\n"), 0644); err != nil {
			return err
		}
		index += fmt.Sprintf("- [%s](%s): %d files.\n", group, filename, len(files))
	}
	index += fmt.Sprintf("\nTotal: %d files, plus this generated reference directory. Generated reference pages describe themselves here rather than recursively indexing their own contents.\n\n", total)
	index += "## Local files intentionally represented by category\n\n- `.git/`: Git objects, refs and repository metadata; no application behavior.\n- `.data/`, credential directories, private keys, `.env` and local assistant settings: private runtime state/configuration; never publish their contents.\n- `.worktrees/`: separate checkouts; document each checkout from its own root.\n- `bin/`, `coverage/`, dependency caches, logs and OS metadata: generated outputs; rebuild or regenerate rather than treating them as source.\n- Symlinks: not followed, to avoid escaping the checkout or exposing private data.\n\nThese categories deliberately do not enumerate private filenames or duplicate generated artifacts.\n"
	return os.WriteFile(filepath.Join(out, "README.md"), []byte(strings.TrimRight(index, "\n")+"\n"), 0644)
}
