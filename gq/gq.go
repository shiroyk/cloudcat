// Package gq the goquery executor
package gq

import (
	"context"
	"fmt"
	"maps"
	"strings"
	"sync/atomic"

	"github.com/PuerkitoBio/goquery"
	"github.com/andybalholm/cascadia"
	"github.com/shiroyk/ski"
	"golang.org/x/net/html"
)

var buildInFuncs atomic.Value

func init() {
	buildInFuncs.Store(builtins())
	ski.Register("gq", new_value())
	ski.Register("gq.element", new_element())
	ski.Register("gq.elements", new_elements())
}

// SetFuncs set external FuncMap
func SetFuncs(m FuncMap) {
	funcs := maps.Clone(builtins())
	maps.Copy(funcs, m)
	buildInFuncs.Store(funcs)
}

func new_value() ski.NewExecutor {
	return ski.StringExecutor(func(str string) (ski.Executor, error) {
		ret, err := compile(str)
		if err != nil {
			return nil, err
		}
		ret.calls = append(ret.calls, call{fn: value})
		return ret, nil
	})
}

func new_element() ski.NewExecutor {
	return ski.StringExecutor(func(str string) (ski.Executor, error) {
		ret, err := compile(str)
		if err != nil {
			return nil, err
		}
		ret.calls = append(ret.calls, call{fn: element})
		return ret, nil
	})
}

func new_elements() ski.NewExecutor {
	return ski.StringExecutor(func(str string) (ski.Executor, error) {
		ret, err := compile(str)
		if err != nil {
			return nil, err
		}
		ret.calls = append(ret.calls, call{fn: elements})
		return ret, nil
	})
}

func compile(raw string) (ret matcher, err error) {
	funcs := strings.Split(raw, "->")
	if len(funcs) == 1 {
		ret.Matcher, err = cascadia.Compile(funcs[0])
		return
	}
	selector := strings.TrimSpace(funcs[0])
	if len(selector) == 0 {
		ret.Matcher = new(emptyMatcher)
	} else {
		ret.Matcher, err = cascadia.Compile(selector)
		if err != nil {
			return
		}
	}

	ret.calls = make([]call, 0, len(funcs)-1)

	for _, function := range funcs[1:] {
		function = strings.TrimSpace(function)
		if function == "" {
			continue
		}
		name, args, err := parseFuncArguments(function)
		if err != nil {
			return ret, err
		}
		fn, ok := buildInFuncs.Load().(FuncMap)[name]
		if !ok {
			return ret, fmt.Errorf("function %s not exists", name)
		}
		ret.calls = append(ret.calls, call{fn, args})
	}

	return
}

type call struct {
	fn   Func
	args []string
}

type matcher struct {
	goquery.Matcher
	calls []call
}

func (f matcher) Exec(ctx context.Context, arg any) (any, error) {
	nodes, err := selection(arg)
	if err != nil {
		return nil, err
	}

	var node any = nodes.FindMatcher(f)

	for _, c := range f.calls {
		node, err = c.fn(ctx, node, c.args...)
		if err != nil || node == nil {
			return nil, err
		}
	}

	return node, nil
}

func value(ctx context.Context, node any, _ ...string) (any, error) {
	if node == nil {
		return nil, nil
	}
	v, err := Text(ctx, node)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func element(_ context.Context, node any, _ ...string) (any, error) {
	switch t := node.(type) {
	default:
		return nil, fmt.Errorf("unexpected type %T", node)
	case string, []string, *html.Node, ski.Iterator, nil:
		return t, nil
	case *goquery.Selection:
		if len(t.Nodes) == 0 {
			return nil, nil
		}
		return t.Nodes[0], nil
	case []*html.Node:
		if len(t) == 0 {
			return nil, nil
		}
		return t[0], nil
	}
}

func elements(_ context.Context, node any, _ ...string) (any, error) {
	switch t := node.(type) {
	default:
		return nil, fmt.Errorf("unexpected type %T", node)
	case string, []string, *html.Node, ski.Iterator, nil:
		return t, nil
	case *goquery.Selection:
		return ski.NewIterator(t.Nodes), nil
	case []*html.Node:
		return ski.NewIterator(t), nil
	}
}

func cloneNode(n *html.Node) *html.Node {
	m := &html.Node{
		Type:       n.Type,
		DataAtom:   n.DataAtom,
		Data:       n.Data,
		Attr:       make([]html.Attribute, len(n.Attr)),
		FirstChild: n.FirstChild,
		LastChild:  n.LastChild,
	}
	copy(m.Attr, n.Attr)
	return m
}

// selection converts content to goquery.Selection
func selection(content any) (*goquery.Selection, error) {
	switch data := content.(type) {
	default:
		return nil, fmt.Errorf("unexpected type %T", content)
	case nil:
		return new(goquery.Selection), nil
	case *html.Node:
		root := &html.Node{Type: html.DocumentNode}
		root.AppendChild(cloneNode(data))
		return goquery.NewDocumentFromNode(root).Selection, nil
	case ski.Iterator:
		if data.Len() == 0 {
			return nil, nil
		}
		root := &html.Node{Type: html.DocumentNode}
		doc := goquery.NewDocumentFromNode(root)
		for i := 0; i < data.Len(); i++ {
			switch v := data.At(i).(type) {
			case *html.Node:
				root.AppendChild(cloneNode(v))
			case string:
				nodes, err := html.ParseFragment(strings.NewReader(v), &html.Node{Type: html.ElementNode})
				if err != nil {
					return nil, err
				}
				for _, node := range nodes {
					root.AppendChild(cloneNode(node))
				}
			default:
				return nil, fmt.Errorf("unexpected type %T", v)
			}
		}
		return doc.Selection, nil
	case []string:
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(strings.Join(data, "\n")))
		if err != nil {
			return nil, err
		}
		return doc.Selection, nil
	case fmt.Stringer:
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(data.String()))
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

type emptyMatcher struct{}

func (emptyMatcher) Match(*html.Node) bool { return true }

func (emptyMatcher) MatchAll(node *html.Node) []*html.Node { return []*html.Node{node} }

func (emptyMatcher) Filter(nodes []*html.Node) []*html.Node { return nodes }
