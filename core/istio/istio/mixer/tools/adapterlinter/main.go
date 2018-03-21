// Copyright 2018 Istio Authors
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

package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"sort"
	"strings"
)

var exitCode int

var invalidImportPaths = map[string]string{
	"log": "\"log\" import is not recommended; Adapters must instead use env.Logger for logging during execution. " +
		"This logger understands which adapter is running and routes the data to the place where the operator " +
		"wants to see it.",
	"github.com/golang/glog": "\"github.com/golang/glog\" import is not recommended; Adapters must instead use env.Logger for logging during execution. " +
		"This logger understands which adapter is running and routes the data to the place where the operator " +
		"wants to see it.",
}

func main() {
	flag.Parse()
	for _, r := range getReport(flag.Args()) {
		reportErr(r)
	}
	os.Exit(exitCode)
}

func getReport(args []string) []string {
	var reports []string
	if len(args) == 0 {
		reports = doAllDirs([]string{"."})
	} else {
		reports = doAllDirs(args)
	}
	return reports
}

func doAllDirs(args []string) []string {
	reports := make([]string, 0)
	for _, name := range args {
		// Is it a directory?
		if fi, err := os.Stat(name); err == nil && fi.IsDir() {
			for _, r := range doDir(name) {
				reports = append(reports, r.msg)
			}
		} else {
			reportErr(fmt.Sprintf("not a directory: %s", name))
		}
	}
	return reports
}

func doDir(name string) reports {
	notests := func(info os.FileInfo) bool {
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") &&
			!strings.HasSuffix(info.Name(), "_test.go") {
			return true
		}
		return false
	}
	fs := token.NewFileSet()
	pkgs, err := parser.ParseDir(fs, name, notests, parser.Mode(0))
	if err != nil {
		reportErr(fmt.Sprintf("%v", err))
		return nil
	}
	rpts := make(reports, 0)
	for _, pkg := range pkgs {
		rpts = append(rpts, doPackage(fs, pkg)...)
	}
	sort.Sort(rpts)
	return rpts
}

func doPackage(fs *token.FileSet, pkg *ast.Package) reports {
	v := newVisitor(fs)
	for _, file := range pkg.Files {
		ast.Walk(&v, file)
	}
	return v.reports
}

func newVisitor(fs *token.FileSet) visitor {
	return visitor{
		fs: fs,
	}
}

type visitor struct {
	reports reports
	fs      *token.FileSet
}

/*
Validates the following:
1. Disallow use of goroutines
2. Disallow use of invalid imports.
*/
func (v *visitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}

	switch d := node.(type) {
	case *ast.GoStmt:
		v.reports = append(v.reports, v.goroutineReport(d.Pos()))
	case *ast.ImportSpec:
		if d.Path != nil {
			p := strings.Trim(d.Path.Value, "\"")
			for badImp, reportMsg := range invalidImportPaths {
				if p == badImp {
					v.reports = append(v.reports, v.invalidImportReport(d.Path.Pos(), reportMsg))
				}
			}
		}
	}

	return v
}

func (v *visitor) invalidImportReport(pos token.Pos, msg string) report {
	return report{
		pos,
		fmt.Sprintf("%v:%v:%v:%s",
			v.fs.Position(pos).Filename,
			v.fs.Position(pos).Line,
			v.fs.Position(pos).Column,
			msg),
	}
}

func (v *visitor) goroutineReport(pos token.Pos) report {
	return report{
		pos,
		fmt.Sprintf("%v:%v:%v:Adapters must use env.ScheduleWork or env.ScheduleDaemon in order to "+
			"dispatch goroutines. This ensures all adapter goroutines are prevented from crashing Mixer as a "+
			"whole by catching any panics they produce.",
			v.fs.Position(pos).Filename, v.fs.Position(pos).Line, v.fs.Position(pos).Column),
	}
}

type report struct {
	pos token.Pos
	msg string
}

type reports []report

func (l reports) Len() int           { return len(l) }
func (l reports) Less(i, j int) bool { return l[i].pos < l[j].pos }
func (l reports) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }

func reportErr(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	exitCode = 2
}
