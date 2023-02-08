package http

import (
	"context"
	"fmt"
	"testing"

	"github.com/shiroyk/cloudcat/js/modulestest"
)

func TestURLSearchParams(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	vm := modulestest.New()

	testCase := []string{
		`form = new URLSearchParams();form.sort();`,
		`try {
			form = new URLSearchParams(0);
		 } catch (e) {
			assert(e.toString().includes('unsupported type'))
		 }`,
		`form = new URLSearchParams({'name': 'foo'});
		 form.forEach((v, k) => assert(v.length == 1))
		 assert.equal(form.get('name'), 'foo')`,
		`form.append('name', 'bar');
		 assert.equal(form.getAll('name').length, 2)`,
		`assert.equal(form.toString(), 'name=foo&name=bar')`,
		`form.append('value', 'zoo');
		 assert(compareArray(form.keys(), ['name', 'value']))`,
		`assert.equal(form.entries().length, 3)`,
		`form.delete('name');
		 assert.equal(form.getAll('name').length, 0)`,
		`assert(!form.has('name'))`,
		`form.set('name', 'foobar');
		 assert.equal(form.values().length, 2)`,
	} //nolint:gofumpt

	for i, s := range testCase {
		t.Run(fmt.Sprintf("Script%v", i), func(t *testing.T) {
			_, err := vm.RunString(ctx, s)
			if err != nil {
				t.Errorf("Script: %s , %s", s, err)
			}
		})
	}
}
