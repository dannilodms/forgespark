// Package controlflow checks the two rules of Control flow that no linter of
// golangci-lint checks whole: no `else`, and one level of nesting in a
// function.
//
// revive has a rule for nesting, `max-control-nesting`, and it does not
// count a `range` loop or a case of a `select`, so it lets the most common
// second level through: an `if` inside a `for ... range`.
package controlflow

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"

	"github.com/dannilodms/forgespark/apps/api/tools/guard/gosource"
	"github.com/dannilodms/forgespark/apps/api/tools/guard/report"
)

const (
	pointer        = " (Control flow in CODING_STANDARDS.md)"
	elseMessage    = "else: handle the exceptional case first and return" + pointer
	nestingMessage = "a second level of nesting: move the inner block into a function of its own" + pointer
)

// Check reads every Go file under root, test files included, and returns one
// finding for each `else`, an `else if` included, and one for each block
// that opens inside another block of the same function. The blocks are `if`,
// `for`, `switch` and `select`. A function literal starts again from zero.
// Generated code is not judged.
func Check(root string) ([]report.Finding, error) {
	files, err := gosource.Files(root)
	if err != nil {
		return nil, err
	}
	var findings []report.Finding
	var unread error
	for _, file := range files {
		found, err := checkFile(root, file)
		unread = errors.Join(unread, err)
		findings = append(findings, found...)
	}
	if unread != nil {
		return nil, unread
	}
	report.Sort(findings)
	return findings, nil
}

func checkFile(root, file string) ([]report.Finding, error) {
	positions := token.NewFileSet()
	parsed, err := parser.ParseFile(positions, filepath.Join(root, filepath.FromSlash(file)), nil, parser.ParseComments|parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}
	// Nobody writes generated code, so its style is not judged. A file is
	// generated when it carries the comment Go defines for it, before the
	// package clause: `// Code generated ... DO NOT EDIT.`
	if ast.IsGenerated(parsed) {
		return nil, nil
	}
	found := &findings{file: file, positions: positions, chained: map[ast.Node]bool{}}
	ast.Walk(walk{found: found}, parsed)
	return found.list, nil
}

// findings collects what one file breaks.
type findings struct {
	file      string
	positions *token.FileSet
	list      []report.Finding
	// chained holds the `if` of each `else if`. It continues the level of the
	// `if` before it, and its `else` is already a finding.
	chained map[ast.Node]bool
}

func (f *findings) add(position token.Pos, message string) {
	f.list = append(f.list, report.Finding{File: f.file, Line: f.positions.Position(position).Line, Message: message})
}

// walk visits the statements of a file, counting the blocks that are open
// around each one.
type walk struct {
	found *findings
	depth int
}

func (w walk) Visit(node ast.Node) ast.Visitor {
	if _, isLiteral := node.(*ast.FuncLit); isLiteral {
		return walk{found: w.found}
	}
	if !opensBlock(node) {
		return w
	}
	depth := w.depth
	if w.found.chained[node] {
		depth--
	}
	if depth > 0 {
		w.found.add(node.Pos(), nestingMessage)
	}
	w.found.addElse(node)
	return walk{found: w.found, depth: depth + 1}
}

// addElse reports the `else` of an `if`.
func (f *findings) addElse(node ast.Node) {
	statement, isIf := node.(*ast.IfStmt)
	if !isIf || statement.Else == nil {
		return
	}
	// The `else` keyword sits on the line where the block before it ends.
	f.add(statement.Body.End(), elseMessage)
	f.chained[statement.Else] = true
}

// opensBlock tells whether a statement opens a level of nesting.
func opensBlock(node ast.Node) bool {
	switch node.(type) {
	case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.SwitchStmt, *ast.TypeSwitchStmt, *ast.SelectStmt:
		return true
	}
	return false
}
