// SPDX-License-Identifier: Apache-2.0 OR MIT
// Copyright (c) 2026 Sebastien Rousseau

package archtest

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os/exec"
	"slices"
	"strings"
	"testing"
)

const module = "github.com/sebastienrousseau/dot-top"

// uiRoot is the tree that must do no I/O: Update and View stay pure.
const uiRoot = module + "/internal/ui"

// forbidden lists packages ui code must not import, directly or through
// another package of this module. Subpackages count (os/exec, net/http).
var forbidden = []string{"os", "net", "syscall"}

type pkg struct {
	ImportPath string
	Imports    []string
	Deps       []string
}

func isForbidden(path string) bool {
	for _, f := range forbidden {
		if path == f || strings.HasPrefix(path, f+"/") {
			return true
		}
	}
	return false
}

func inUI(path string) bool { return path == uiRoot || strings.HasPrefix(path, uiRoot+"/") }

// violations reports "<ui package> -> <module package> imports <forbidden>"
// for every ui package whose own imports, or whose in-module dependencies'
// imports, include a forbidden package. Standard-library internals (fmt
// importing os) are not counted.
func violations(pkgs []pkg) []string {
	byPath := make(map[string]pkg, len(pkgs))
	for _, p := range pkgs {
		byPath[p.ImportPath] = p
	}
	var out []string
	for _, p := range pkgs {
		if !inUI(p.ImportPath) {
			continue
		}
		for _, src := range append([]string{p.ImportPath}, p.Deps...) {
			dep, ok := byPath[src]
			if !ok {
				continue // standard library or third party
			}
			for _, imp := range forbiddenImports(dep) {
				out = append(out, p.ImportPath+" -> "+src+" imports "+imp)
			}
		}
	}
	slices.Sort(out)
	return out
}

func forbiddenImports(p pkg) []string {
	var out []string
	for _, imp := range p.Imports {
		if isForbidden(imp) {
			out = append(out, imp)
		}
	}
	return out
}

func decode(r io.Reader) ([]pkg, error) {
	var pkgs []pkg
	dec := json.NewDecoder(r)
	for {
		var p pkg
		err := dec.Decode(&p)
		if errors.Is(err, io.EOF) {
			return pkgs, nil
		}
		if err != nil {
			return nil, err
		}
		pkgs = append(pkgs, p)
	}
}

func TestUIDoesNoIO(t *testing.T) {
	cmd := exec.Command("go", "list", "-json=ImportPath,Imports,Deps", module+"/...")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	raw, err := cmd.Output()
	if err != nil {
		t.Fatalf("go list: %v\n%s", err, stderr.String())
	}
	pkgs, err := decode(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	if !slices.ContainsFunc(pkgs, func(p pkg) bool { return inUI(p.ImportPath) }) {
		t.Fatal("no ui packages found; the rule would pass vacuously")
	}
	if v := violations(pkgs); len(v) > 0 {
		t.Fatalf("internal/ui must not do I/O:\n%s", strings.Join(v, "\n"))
	}
}

func TestViolationsDetectsDirectAndIndirect(t *testing.T) {
	pkgs := []pkg{
		{ImportPath: uiRoot + "/a", Imports: []string{"fmt", "os/exec"}},
		{ImportPath: uiRoot + "/b", Imports: []string{module + "/internal/io"}, Deps: []string{module + "/internal/io", "fmt", "os"}},
		{ImportPath: module + "/internal/io", Imports: []string{"net"}},
		{ImportPath: uiRoot + "/clean", Imports: []string{"fmt", "unicode/utf8"}, Deps: []string{"fmt", "os"}},
		{ImportPath: module + "/internal/uix", Imports: []string{"syscall"}}, // not under ui/
	}
	want := []string{
		uiRoot + "/a -> " + uiRoot + "/a imports os/exec",
		uiRoot + "/b -> " + module + "/internal/io imports net",
	}
	if got := violations(pkgs); !slices.Equal(got, want) {
		t.Fatalf("violations =\n%q\nwant\n%q", got, want)
	}
}
