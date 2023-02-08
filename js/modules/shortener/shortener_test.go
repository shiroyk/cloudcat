package shortener

import (
	"context"
	"testing"

	"github.com/shiroyk/cloudcat/cache"
	"github.com/shiroyk/cloudcat/cache/memory"
	"github.com/shiroyk/cloudcat/di"
	"github.com/shiroyk/cloudcat/js/modulestest"
)

func TestShortener(t *testing.T) {
	t.Parallel()
	di.Override[cache.Shortener](memory.NewShortener())
	ctx := context.Background()
	vm := modulestest.New()

	tpl := `POST http://localhost
Content-Type: application/json

{\"key\":\"foo\"}`
	_, err := vm.RunString(ctx,
		"const s = require('cloudcat/shortener');"+
			"id = s.set(`"+tpl+"`);"+
			"assert.equal(s.get(id), `"+tpl+"`);",
	)
	if err != nil {
		t.Fatal(err)
	}
}