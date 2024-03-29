// Package gq the goquery parser
package gq

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/shiroyk/cloudcat/plugin"
	"github.com/shiroyk/cloudcat/plugin/parser"
	"github.com/spf13/cast"
)

// Key the gq parser register key.
const Key string = "gq"

// Parser the goquery parser
type Parser struct {
	parseFuncs FuncMap
}

// NewParser creates a new goquery Parser with the given FuncMap.
func NewParser(funcs FuncMap) parser.Parser {
	p := &Parser{parseFuncs: builtins()}
	for k, v := range funcs {
		p.parseFuncs[k] = v
	}
	return p
}

// init register the goquery Parser with default builtins funcs.
func init() {
	parser.Register(Key, &Parser{parseFuncs: builtins()})
}

// GetString gets the string of the content with the given arguments.
//
// content := `<ul><li>1</li><li>2</li></ul>`
// GetString(ctx, content, "ul li") returns "1\n2"
func (p *Parser) GetString(ctx *plugin.Context, content any, arg string) (ret string, err error) {
	nodes, err := getSelection(content)
	if err != nil {
		return
	}

	rule, funcs, err := parseRuleFunctions(p.parseFuncs, arg)
	if err != nil {
		return
	}

	var node any = nodes.Find(rule)

	for _, fun := range funcs {
		node, err = p.parseFuncs[fun.name](ctx, node, fun.args...)
		if err != nil || node == nil {
			return ret, err
		}
	}

	node, err = Join(ctx, node, "\n")
	if err != nil {
		return
	}

	return node.(string), nil
}

// GetStrings gets the strings of the content with the given arguments.
//
// content := `<ul><li>1</li><li>2</li></ul>`
// GetStrings(ctx, content, "ul li") returns []string{"1", "2"}
func (p *Parser) GetStrings(ctx *plugin.Context, content any, arg string) (ret []string, err error) {
	nodes, err := getSelection(content)
	if err != nil {
		return
	}

	rule, funcs, err := parseRuleFunctions(p.parseFuncs, arg)
	if err != nil {
		return
	}

	var node any = nodes.Find(rule)

	for _, fun := range funcs {
		node, err = p.parseFuncs[fun.name](ctx, node, fun.args...)
		if err != nil || node == nil {
			return nil, err
		}
	}

	if sel, ok := node.(*goquery.Selection); ok {
		str := make([]string, sel.Length())
		sel.EachWithBreak(func(i int, sel *goquery.Selection) bool {
			str[i] = strings.TrimSpace(sel.Text())
			return true
		})
		return str, nil
	}
	return cast.ToStringSliceE(node)
}

// GetElement gets the element of the content with the given arguments.
//
// content := `<ul><li>1</li><li>2</li></ul>`
// GetElement(ctx, content, "ul li") returns "<li>1</li>\n<li>2</li>"
func (p *Parser) GetElement(ctx *plugin.Context, content any, arg string) (ret string, err error) {
	nodes, err := getSelection(content)
	if err != nil {
		return
	}

	rule, funcs, err := parseRuleFunctions(p.parseFuncs, arg)
	if err != nil {
		return
	}

	var node any = nodes.Find(rule)

	for _, fun := range funcs {
		node, err = p.parseFuncs[fun.name](ctx, node, fun.args...)
		if err != nil || node == nil {
			return ret, err
		}
	}

	if sel, ok := node.(*goquery.Selection); ok {
		return goquery.OuterHtml(sel)
	}

	return cast.ToStringE(node)
}

// GetElements gets the elements of the content with the given arguments.
//
// content := `<ul><li>1</li><li>2</li></ul>`
// GetElements(ctx, content, "ul li") returns []string{"<li>1</li>", "<li>2</li>"}
func (p *Parser) GetElements(ctx *plugin.Context, content any, arg string) (ret []string, err error) {
	nodes, err := getSelection(content)
	if err != nil {
		return
	}

	rule, funcs, err := parseRuleFunctions(p.parseFuncs, arg)
	if err != nil {
		return
	}

	var node any = nodes.Find(rule)

	for _, fun := range funcs {
		node, err = p.parseFuncs[fun.name](ctx, node, fun.args...)
		if err != nil || node == nil {
			return nil, err
		}
	}

	if sel, ok := node.(*goquery.Selection); ok {
		objs := make([]string, sel.Length())
		sel.EachWithBreak(func(i int, sel *goquery.Selection) bool {
			if objs[i], err = goquery.OuterHtml(sel); err != nil {
				return false
			}
			return true
		})
		if err != nil {
			return
		}
		return objs, nil
	}
	return cast.ToStringSliceE(node)
}

// getSelection converts content to goquery.Selection
func getSelection(content any) (*goquery.Selection, error) {
	switch data := content.(type) {
	default:
		str, err := cast.ToStringE(content)
		if err != nil {
			return nil, err
		}
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(str))
		if err != nil {
			return nil, err
		}
		return doc.Selection, nil
	case nil:
		return new(goquery.Selection), nil
	case []string:
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(strings.Join(data, "\n")))
		if err != nil {
			return nil, err
		}
		return doc.Selection, nil
	case string:
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(data))
		if err != nil {
			return nil, err
		}
		return doc.Selection, nil
	}
}
