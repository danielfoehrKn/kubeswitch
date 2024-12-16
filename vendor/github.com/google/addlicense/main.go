// Copyright 2018 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// This program ensures source code files have copyright license headers.
// See usage with "addlicense -h".
package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	doublestar "github.com/bmatcuk/doublestar/v4"
	"golang.org/x/sync/errgroup"
)

const helpText = `Usage: addlicense [flags] pattern [pattern ...]

The program ensures source code files have copyright license headers
by scanning directory patterns recursively.

It modifies all source files in place and avoids adding a license header
to any file that already has one.

The pattern argument can be provided multiple times, and may also refer
to single files.

Flags:

`

var (
	skipExtensionFlags stringSlice
	ignorePatterns     stringSlice
	spdx               spdxFlag

	holder    = flag.String("c", "Google LLC", "copyright holder")
	license   = flag.String("l", "apache", "license type: apache, bsd, mit, mpl")
	licensef  = flag.String("f", "", "license file")
	year      = flag.String("y", fmt.Sprint(time.Now().Year()), "copyright year(s)")
	verbose   = flag.Bool("v", false, "verbose mode: print the name of the files that are modified or were skipped")
	checkonly = flag.Bool("check", false, "check only mode: verify presence of license headers and exit with non-zero code if missing")
)

func init() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, helpText)
		flag.PrintDefaults()
	}
	flag.Var(&skipExtensionFlags, "skip", "[deprecated: see -ignore] file extensions to skip, for example: -skip rb -skip go")
	flag.Var(&ignorePatterns, "ignore", "file patterns to ignore, for example: -ignore **/*.go -ignore vendor/**")
	flag.Var(&spdx, "s", "Include SPDX identifier in license header. Set -s=only to only include SPDX identifier.")
}

// stringSlice stores the results of a repeated command line flag as a string slice.
type stringSlice []string

func (i *stringSlice) String() string {
	return fmt.Sprint(*i)
}

func (i *stringSlice) Set(value string) error {
	*i = append(*i, value)
	return nil
}

// spdxFlag defines the line flag behavior for specifying SPDX support.
type spdxFlag string

const (
	spdxOff  spdxFlag = ""
	spdxOn   spdxFlag = "true" // value set by flag package on bool flag
	spdxOnly spdxFlag = "only"
)

// IsBoolFlag causes a bare '-s' flag to be set as the string 'true'.  This
// allows the use of the bare '-s' or setting a string '-s=only'.
func (i *spdxFlag) IsBoolFlag() bool { return true }
func (i *spdxFlag) String() string   { return string(*i) }

func (i *spdxFlag) Set(value string) error {
	v := spdxFlag(value)
	if v != spdxOn && v != spdxOnly {
		return fmt.Errorf("error: flag 's' expects '%v' or '%v'", spdxOn, spdxOnly)
	}
	*i = v
	return nil
}

func main() {
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	// convert -skip flags to -ignore equivalents
	for _, s := range skipExtensionFlags {
		ignorePatterns = append(ignorePatterns, fmt.Sprintf("**/*.%s", s))
	}
	// verify that all ignorePatterns are valid
	for _, p := range ignorePatterns {
		if !doublestar.ValidatePattern(p) {
			log.Fatalf("-ignore pattern %q is not valid", p)
		}
	}

	// map legacy license values
	if t, ok := legacyLicenseTypes[*license]; ok {
		*license = t
	}

	data := licenseData{
		Year:   *year,
		Holder: *holder,
		SPDXID: *license,
	}

	tpl, err := fetchTemplate(*license, *licensef, spdx)
	if err != nil {
		log.Fatal(err)
	}
	t, err := template.New("").Parse(tpl)
	if err != nil {
		log.Fatal(err)
	}

	// process at most 1000 files in parallel
	ch := make(chan *file, 1000)
	done := make(chan struct{})
	go func() {
		var wg errgroup.Group
		for f := range ch {
			f := f // https://golang.org/doc/faq#closures_and_goroutines
			wg.Go(func() error {
				if *checkonly {
					// Check if file extension is known
					lic, err := licenseHeader(f.path, t, data)
					if err != nil {
						log.Printf("%s: %v", f.path, err)
						return err
					}
					if lic == nil { // Unknown fileExtension
						return nil
					}
					// Check if file has a license
					hasLicense, err := fileHasLicense(f.path)
					if err != nil {
						log.Printf("%s: %v", f.path, err)
						return err
					}
					if !hasLicense {
						fmt.Printf("%s\n", f.path)
						return errors.New("missing license header")
					}
				} else {
					modified, err := addLicense(f.path, f.mode, t, data)
					if err != nil {
						log.Printf("%s: %v", f.path, err)
						return err
					}
					if *verbose && modified {
						log.Printf("%s modified", f.path)
					}
				}
				return nil
			})
		}
		err := wg.Wait()
		close(done)
		if err != nil {
			os.Exit(1)
		}
	}()

	for _, d := range flag.Args() {
		if err := walk(ch, d); err != nil {
			log.Fatal(err)
		}
	}
	close(ch)
	<-done
}

type file struct {
	path string
	mode os.FileMode
}

func walk(ch chan<- *file, start string) error {
	return filepath.Walk(start, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			log.Printf("%s error: %v", path, err)
			return nil
		}
		if fi.IsDir() {
			return nil
		}
		if fileMatches(path, ignorePatterns) {
			if *verbose {
				log.Printf("skipping: %s", path)
			}
			return nil
		}
		ch <- &file{path, fi.Mode()}
		return nil
	})
}

// fileMatches determines if path matches one of the provided file patterns.
// Patterns are assumed to be valid.
func fileMatches(path string, patterns []string) bool {
	for _, p := range patterns {
		// ignore error, since we assume patterns are valid
		if match, _ := doublestar.Match(p, path); match {
			return true
		}
	}
	return false
}

// addLicense add a license to the file if missing.
//
// It returns true if the file was updated.
func addLicense(path string, fmode os.FileMode, tmpl *template.Template, data licenseData) (bool, error) {
	var lic []byte
	var err error
	lic, err = licenseHeader(path, tmpl, data)
	if err != nil || lic == nil {
		return false, err
	}

	b, err := ioutil.ReadFile(path)
	if err != nil {
		return false, err
	}
	if hasLicense(b) || isGenerated(b) {
		return false, err
	}

	line := hashBang(b)
	if len(line) > 0 {
		b = b[len(line):]
		if line[len(line)-1] != '\n' {
			line = append(line, '\n')
		}
		lic = append(line, lic...)
	}
	b = append(lic, b...)
	return true, ioutil.WriteFile(path, b, fmode)
}

// fileHasLicense reports whether the file at path contains a license header.
func fileHasLicense(path string) (bool, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return false, err
	}
	// If generated, we count it as if it has a license.
	return hasLicense(b) || isGenerated(b), nil
}

// licenseHeader populates the provided license template with data, and returns
// it with the proper prefix for the file type specified by path. The file does
// not need to actually exist, only its name is used to determine the prefix.
func licenseHeader(path string, tmpl *template.Template, data licenseData) ([]byte, error) {
	var lic []byte
	var err error
	base := strings.ToLower(filepath.Base(path))

	switch fileExtension(base) {
	case ".c", ".h", ".gv", ".java", ".scala", ".kt", ".kts":
		lic, err = executeTemplate(tmpl, data, "/*", " * ", " */")
	case ".js", ".mjs", ".cjs", ".jsx", ".tsx", ".css", ".scss", ".sass", ".ts":
		lic, err = executeTemplate(tmpl, data, "/**", " * ", " */")
	case ".cc", ".cpp", ".cs", ".go", ".hcl", ".hh", ".hpp", ".m", ".mm", ".proto", ".rs", ".swift", ".dart", ".groovy", ".v", ".sv":
		lic, err = executeTemplate(tmpl, data, "", "// ", "")
	case ".py", ".sh", ".yaml", ".yml", ".dockerfile", "dockerfile", ".rb", "gemfile", ".tcl", ".tf", ".bzl", ".pl", ".pp", "build":
		lic, err = executeTemplate(tmpl, data, "", "# ", "")
	case ".el", ".lisp":
		lic, err = executeTemplate(tmpl, data, "", ";; ", "")
	case ".erl":
		lic, err = executeTemplate(tmpl, data, "", "% ", "")
	case ".hs", ".sql", ".sdl":
		lic, err = executeTemplate(tmpl, data, "", "-- ", "")
	case ".html", ".xml", ".vue", ".wxi", ".wxl", ".wxs":
		lic, err = executeTemplate(tmpl, data, "<!--", " ", "-->")
	case ".php":
		lic, err = executeTemplate(tmpl, data, "", "// ", "")
	case ".ml", ".mli", ".mll", ".mly":
		lic, err = executeTemplate(tmpl, data, "(**", "   ", "*)")
	default:
		// handle various cmake files
		if base == "cmakelists.txt" || strings.HasSuffix(base, ".cmake.in") || strings.HasSuffix(base, ".cmake") {
			lic, err = executeTemplate(tmpl, data, "", "# ", "")
		}
	}
	return lic, err
}

// fileExtension returns the file extension of name, or the full name if there
// is no extension.
func fileExtension(name string) string {
	if v := filepath.Ext(name); v != "" {
		return v
	}
	return name
}

var head = []string{
	"#!",                       // shell script
	"<?xml",                    // XML declaratioon
	"<!doctype",                // HTML doctype
	"# encoding:",              // Ruby encoding
	"# frozen_string_literal:", // Ruby interpreter instruction
	"<?php",                    // PHP opening tag
	"# escape",                 // Dockerfile directive https://docs.docker.com/engine/reference/builder/#parser-directives
	"# syntax",                 // Dockerfile directive https://docs.docker.com/engine/reference/builder/#parser-directives
}

func hashBang(b []byte) []byte {
	var line []byte
	for _, c := range b {
		line = append(line, c)
		if c == '\n' {
			break
		}
	}
	first := strings.ToLower(string(line))
	for _, h := range head {
		if strings.HasPrefix(first, h) {
			return line
		}
	}
	return nil
}

// go generate: ^// Code generated .* DO NOT EDIT\.$
var goGenerated = regexp.MustCompile(`(?m)^.{1,2} Code generated .* DO NOT EDIT\.$`)

// cargo raze: ^DO NOT EDIT! Replaced on runs of cargo-raze$
var cargoRazeGenerated = regexp.MustCompile(`(?m)^DO NOT EDIT! Replaced on runs of cargo-raze$`)

// isGenerated returns true if it contains a string that implies the file was
// generated.
func isGenerated(b []byte) bool {
	return goGenerated.Match(b) || cargoRazeGenerated.Match(b)
}

func hasLicense(b []byte) bool {
	n := 1000
	if len(b) < 1000 {
		n = len(b)
	}
	return bytes.Contains(bytes.ToLower(b[:n]), []byte("copyright")) ||
		bytes.Contains(bytes.ToLower(b[:n]), []byte("mozilla public")) ||
		bytes.Contains(bytes.ToLower(b[:n]), []byte("spdx-license-identifier"))
}
