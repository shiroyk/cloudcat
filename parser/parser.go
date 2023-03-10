// Package parser the schema parser
package parser

import (
	"github.com/shiroyk/cloudcat/internal/ext"
)

// Parser the content schema
type Parser interface {
	// GetString gets the string of the content with the given arguments.
	// e.g.:
	//
	// content := `<ul><li>1</li><li>2</li></ul>`
	// GetString(ctx, content, "ul li") returns "1\n2"
	//
	GetString(*Context, any, string) (string, error)
	// GetStrings gets the strings of the content with the given arguments.
	// e.g.:
	//
	// content := `<ul><li>1</li><li>2</li></ul>`
	// GetStrings(ctx, content, "ul li") returns []string{"1", "2"}
	//
	GetStrings(*Context, any, string) ([]string, error)
	// GetElement gets the element of the content with the given arguments.
	// e.g.:
	//
	// content := `<ul><li>1</li><li>2</li></ul>`
	// GetElement(ctx, content, "ul li") returns "<li>1</li>\n<li>2</li>"
	//
	GetElement(*Context, any, string) (string, error)
	// GetElements gets the elements of the content with the given arguments.
	// e.g.:
	//
	// content := `<ul><li>1</li><li>2</li></ul>`
	// GetElements(ctx, content, "ul li") returns []string{"<li>1</li>", "<li>2</li>"}
	//
	GetElements(*Context, any, string) ([]string, error)
}

// Register registers the Parser with the given key Parser
func Register(key string, parser Parser) {
	if key == "and" || key == "or" {
		panic("register key not supported")
	}
	ext.Register(key, ext.ParserExtension, parser)
}

// GetParser returns a Parser with the given key
func GetParser(key string) (Parser, bool) {
	parsers := ext.Get(ext.ParserExtension)
	if p, ok := parsers[key]; ok {
		return p.Module.(Parser), true
	}
	return nil, false
}
