package gq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildInFuncGet(t *testing.T) {
	t.Parallel()
	if _, err := gq.GetString(ctx, content, `-> get`); err == nil {
		t.Error("Unexpected function error")
	}

	if _, err := gq.GetString(ctx, content, `.body #a1 -> set(key111)`); err != nil {
		t.Error(err)
	}

	assertGetString(t, `-> get(key111) -> child`, "Google")
}

func TestBuildInFuncSet(t *testing.T) {
	t.Parallel()
	if _, err := gq.GetString(ctx, content, `-> set`); err == nil {
		t.Fatal("Unexpected function error")
	}

	if _, err := gq.GetString(ctx, content, `-> set(v1, '<i>v1</i>')`); err != nil {
		t.Error(err)
	}

	if _, err := gq.GetString(ctx, content, `.body #a1 -> text -> set(key222)`); err != nil {
		t.Error(err)
	}
}

func TestBuildInFuncText(t *testing.T) {
	t.Parallel()

	assertGetString(t, `#main #n1 -> text`, "1")

	assertGetString(t, `#main #n1`, "1")
}

func TestBuildInFuncAttr(t *testing.T) {
	t.Parallel()
	if _, err := gq.GetString(ctx, content, `#main #n1 -> text -> attr`); err == nil {
		t.Fatal("Unexpected function error")
	}

	if _, err := gq.GetString(ctx, content, `-> attr()`); err == nil {
		t.Fatal("Unexpected null argument")
	}

	assertGetString(t, `#main #n1 -> attr(class)`, "one even row")

	assertGetString(t, `#main #n1 -> attr(empty, default)`, "default")
}

func TestBuildInFuncJoin(t *testing.T) {
	t.Parallel()
	assertGetString(t, `#main div -> join(' < ')`, "1 < 2 < 3 < 4 < 5 < 6")

	assertGetString(t, `#main div -> join("")`, "123456")

	assertGetString(t, `#main div -> join('')`, "123456")
}

func TestBuildInFuncHref(t *testing.T) {
	t.Parallel()
	if _, err := gq.GetString(ctx, content, `.body ul #a4 -> text -> href`); err == nil {
		t.Fatal("Unexpected function error")
	}

	assertGetString(t, `.body ul #a4 a -> href`, "https://localhost/home")

	assertGetString(t, `.body ul #a4 a -> href(path)`, "https://localhost/path/home")

	assertGetString(t, `.body ul #a4 a -> href(path/)`, "https://localhost/path/home")

	assertGetString(t, `.body ul #a4 a -> href(/path/)`, "https://localhost/path/home")

	_, err := gq.GetString(ctx, content, `#main #n1 -> href`)
	assert.Error(t, err)
}

func TestBuildInFuncHtml(t *testing.T) {
	t.Parallel()
	if _, err := gq.GetString(ctx, content, `-> html(test)`); err == nil {
		t.Fatal("Unexpected function error")
	}

	assertGetString(t, `.body ul a -> html`, "Google\nGithub\nGolang\nHome")

	assertGetString(t, `.body ul a -> slice(0) -> html(true)`,
		`<a href="https://google.com" title="Google page">Google</a>`)
}

func TestBuildInFuncPrev(t *testing.T) {
	t.Parallel()
	if _, err := gq.GetString(ctx, content, `#foot #nf3 -> text -> prev`); err == nil {
		t.Fatal("Unexpected function error")
	}

	assertGetString(t, `#foot #nf3 -> prev`, "f2")

	assertGetString(t, `#foot #nf3 -> prev(#nf1)`, "f2")
}

func TestBuildInFuncNext(t *testing.T) {
	t.Parallel()
	if _, err := gq.GetString(ctx, content, `#foot #nf2 -> text -> next`); err == nil {
		t.Fatal("Unexpected function error")
	}

	assertGetString(t, `#foot #nf2 -> next`, "f3")

	assertGetString(t, `#foot #nf2 -> next(#nf4)`, "f3")
}

func TestBuildInFuncSlice(t *testing.T) {
	t.Parallel()
	if _, err := gq.GetString(ctx, content, `-> slice`); err == nil {
		t.Fatal("Unexpected function error")
	}

	if _, err := gq.GetString(ctx, content, `#main div -> text -> slice(0)`); err == nil {
		t.Fatal("Unexpected function error")
	}

	assertGetString(t, `#main div -> slice(0)`, "1")

	assertGetString(t, `#main div -> slice(-1)`, "6")

	assertGetString(t, `#main div -> slice(0, 3)`, "1\n2\n3")

	assertGetString(t, `#main div -> slice(0, -2)`, "1\n2\n3\n4")
}

func TestBuildInFuncChild(t *testing.T) {
	t.Parallel()
	if _, err := gq.GetString(ctx, content, `.body ul -> text -> child`); err == nil {
		t.Fatal("Unexpected function error")
	}

	assertGetString(t, `.body ul li -> child(a)`, "Google\nGithub\nGolang\nHome")

	assertGetString(t, `.body ul li -> child`, "Google\nGithub\nGolang\nHome")
}

func TestBuildInFuncParent(t *testing.T) {
	t.Parallel()
	if _, err := gq.GetString(ctx, content, `.body ul -> text -> parent`); err == nil {
		t.Fatal("Unexpected function error")
	}

	assertGetString(t, `.body ul a -> parent(#a1) -> attr(id)`, "a1")

	assertGetString(t, `.body ul a -> parent -> attr(id)`, "a1\na2\na3\na4")
}

func TestBuildInFuncParents(t *testing.T) {
	t.Parallel()
	if _, err := gq.GetString(ctx, content, `.body ul -> text -> parents`); err == nil {
		t.Fatal("Unexpected type")
	}

	if _, err := gq.GetString(ctx, content, `.body ul .selected -> parents(div, test)`); err == nil {
		t.Fatal("Unexpected argument")
	}

	assertGetString(t, `.body ul .selected -> parents(div, true) -> attr(id)`, "url")

	assertGetString(t, `.body ul .selected -> parents -> slice(0) -> attr(id)`, "url")
}

func TestBuildInFuncPrefix(t *testing.T) {
	t.Parallel()

	assertGetString(t, `#main #n1 -> text -> prefix(A)`, "A1")

	assertGetString(t, `#main #n1 -> prefix(B)`, "B1")

	assertGetStrings(t, `#main div -> slice(0, 2) -> text -> prefix(-)`, []string{"-1", "-2"})
}

func TestBuildInFuncSuffix(t *testing.T) {
	t.Parallel()

	assertGetString(t, `#main #n1 -> text -> suffix(A)`, "1A")

	assertGetString(t, `#main #n1 -> suffix(B)`, "1B")

	assertGetStrings(t, `.body a -> slice(0, 2) -> text -> suffix(.com)`, []string{"Google.com", "Github.com"})
}
