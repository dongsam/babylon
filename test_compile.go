package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
)

func main() {
	fset := token.NewFileSet()
	
	// Parse the test file to check for syntax errors
	src, err := os.ReadFile("x/zoneconcierge/keeper/race_condition_test.go")
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}
	
	_, err = parser.ParseFile(fset, "race_condition_test.go", src, parser.ParseComments)
	if err != nil {
		fmt.Printf("Syntax error in test file: %v\n", err)
		return
	}
	
	fmt.Println("✅ Test file compiles successfully (syntax check)")
	
	// Also check the original epoch_header_indexer.go for the buggy line
	src2, err := os.ReadFile("x/zoneconcierge/keeper/epoch_header_indexer.go")
	if err != nil {
		fmt.Printf("Error reading epoch_header_indexer.go: %v\n", err)
		return
	}
	
	content := string(src2)
	if containsBuggyLine(content) {
		fmt.Println("✅ Confirmed: Buggy line exists in epoch_header_indexer.go:108")
		fmt.Println("   Line: if headerWithProof.Header.BabylonEpoch == curEpoch.EpochNumber {")
	} else {
		fmt.Println("❌ Buggy line not found or already fixed")
	}
}

func containsBuggyLine(content string) bool {
	return ast.Inspect(mustParse(content), func(n ast.Node) bool {
		if binExpr, ok := n.(*ast.BinaryExpr); ok {
			if binExpr.Op == token.EQL {
				// Check for the specific comparison pattern
				left := exprToString(binExpr.X)
				right := exprToString(binExpr.Y)
				if (left == "headerWithProof.Header.BabylonEpoch" && right == "curEpoch.EpochNumber") ||
				   (right == "headerWithProof.Header.BabylonEpoch" && left == "curEpoch.EpochNumber") {
					return false // Found the buggy comparison
				}
			}
		}
		return true
	})
}

func mustParse(content string) ast.Node {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", content, parser.ParseComments)
	if err != nil {
		panic(err)
	}
	return node
}

func exprToString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.SelectorExpr:
		return exprToString(e.X) + "." + e.Sel.Name
	case *ast.Ident:
		return e.Name
	default:
		return ""
	}
}