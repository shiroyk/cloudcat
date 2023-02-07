package json

import (
	"flag"
	"os"
	"reflect"
	"testing"

	c "github.com/shiroyk/cloudcat/parser"
)

var (
	json    Parser //nolint:gochecknoglobals
	ctx     *c.Context
	content = `
{
    "store": {
        "book": [
            {
                "category": "reference",
                "author": "Nigel Rees",
                "title": "Sayings of the Century",
                "price": 8.95
            },
            {
                "category": "fiction",
                "author": "Evelyn Waugh",
                "title": "Sword of Honour",
                "price": 12.99
            },
            {
                "category": "fiction",
                "author": "Herman Melville",
                "title": "Moby Dick",
                "isbn": "0-553-21311-3",
                "price": 8.99
            },
            {
                "category": "fiction",
                "author": "J. R. R. Tolkien",
                "title": "The Lord of the Rings",
                "isbn": "0-395-19395-8",
                "price": 22.99
            }
        ],
        "bicycle": {
            "color": "red",
            "price": 19.95
        }
    },
    "expensive": 10
}`
)

func TestMain(m *testing.M) {
	flag.Parse()
	ctx = c.NewContext(c.Options{Config: c.Config{Separator: ", "}})
	code := m.Run()
	os.Exit(code)
}

func assertString(t *testing.T, arg string, assert func(string) bool) {
	str, err := json.GetString(ctx, content, arg)
	if err != nil {
		t.Fatal(err)
	}

	if !assert(str) {
		t.Fatalf("unexpected result %s", str)
	}
}

func assertStrings(t *testing.T, arg string, assert func([]string) bool) {
	str, err := json.GetStrings(ctx, content, arg)
	if err != nil {
		t.Fatal(err)
	}

	if !assert(str) {
		t.Fatalf("unexpected result %s", str)
	}
}

func assertGetElement(t *testing.T, arg string, assert func(string) bool) {
	obj, err := json.GetElement(ctx, content, arg)
	if err != nil {
		t.Fatal(err)
	}

	if !assert(obj) {
		t.Fatalf("Unexpected object %s", obj)
	}
}

func assertGetElements(t *testing.T, arg string, assert func([]string) bool) {
	objs, err := json.GetElements(ctx, content, arg)
	if err != nil {
		t.Fatal(err)
	}

	if !assert(objs) {
		t.Fatalf("Unexpected objects %s", objs)
	}
}

func TestParser(t *testing.T) {
	if _, ok := c.GetParser(key); !ok {
		t.Fatal("parser not registered")
	}

	contents := []any{114514, `}{`}
	for _, ct := range contents {
		if _, err := json.GetString(ctx, ct, ``); err == nil {
			t.Fatal("Unexpected type")
		}
	}

	if _, err := json.GetString(ctx, &contents[len(contents)-1], ""); err == nil {
		t.Fatal("Unexpected type")
	}
}

func TestGetString(t *testing.T) {
	t.Parallel()
	assertString(t, `$.store.book[*].author`, func(str string) bool {
		return str == `Nigel Rees, Evelyn Waugh, Herman Melville, J. R. R. Tolkien`
	})
}

func TestGetStrings(t *testing.T) {
	t.Parallel()
	assertStrings(t, `$...book[0].price`, func(str []string) bool {
		return reflect.DeepEqual(str, []string{"8.95"})
	})

	assertStrings(t, `$...book[-1].price`, func(str []string) bool {
		return reflect.DeepEqual(str, []string{"22.99"})
	})
}

func TestGetElement(t *testing.T) {
	t.Parallel()
	if _, err := json.GetElement(ctx, content, `$$$`); err == nil {
		t.Fatal("Unexpected path")
	}

	assertGetElement(t, `$.store.book[-1]`, func(obj string) bool {
		return obj != ""
	})

	str1, err := json.GetElement(ctx, content, `$.store.book[?(@.price > 20)]`)
	if err != nil {
		t.Fatal(err)
	}

	str2, err := json.GetElement(ctx, str1, `$.title`)
	if err != nil {
		t.Fatal(err)
	}
	if str2 != `The Lord of the Rings` {
		t.Fatalf("Unexpected string %s", str2)
	}
}

func TestGetElements(t *testing.T) {
	t.Parallel()
	assertGetElements(t, `$.store.book[?(@.price < 10)].isbn`, func(obj []string) bool {
		return obj[0] == `0-553-21311-3`
	})

	str1, err := json.GetElements(ctx, content, `$.store.book[3]`)
	if err != nil {
		t.Fatal(err)
	}

	str2, err := json.GetElement(ctx, str1[0], `$.category`)
	if err != nil {
		t.Fatal(err)
	}
	if str2 != `fiction` {
		t.Fatalf("Unexpected string %s", str2)
	}
}
