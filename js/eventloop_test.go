package js

import (
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBasicEventLoop(t *testing.T) {
	t.Parallel()
	loop := NewEventLoop(NewTestVM(t).Runtime())
	var ran int
	f := func() error { //nolint:unparam
		ran++
		return nil
	}
	assert.NoError(t, loop.Start(f))
	assert.Equal(t, 1, ran)
	assert.NoError(t, loop.Start(f))
	assert.Equal(t, 2, ran)
	assert.Error(t, loop.Start(func() error {
		_ = f()
		loop.RegisterCallback()(f)
		return errors.New("something")
	}))
	assert.Equal(t, 3, ran)
}

func TestEventLoopRegistered(t *testing.T) {
	t.Parallel()
	loop := NewEventLoop(NewTestVM(t).Runtime())
	var ran int
	f := func() error {
		ran++
		r := loop.RegisterCallback()
		go func() {
			time.Sleep(time.Second)
			r(func() error {
				ran++
				return nil
			})
		}()
		return nil
	}
	start := time.Now()
	assert.NoError(t, loop.Start(f))
	took := time.Since(start)
	assert.Equal(t, 2, ran)
	assert.Less(t, time.Second, took)
	assert.Greater(t, time.Second+time.Millisecond*100, took)
}

func TestEventLoopWaitOnRegistered(t *testing.T) {
	t.Parallel()
	var ran int
	loop := NewEventLoop(NewTestVM(t).Runtime())
	f := func() error {
		ran++
		r := loop.RegisterCallback()
		go func() {
			time.Sleep(time.Second)
			r(func() error {
				ran++
				return nil
			})
		}()
		return fmt.Errorf("expected")
	}
	start := time.Now()
	assert.Error(t, loop.Start(f))
	took := time.Since(start)
	loop.WaitOnRegistered()
	took2 := time.Since(start)
	assert.Equal(t, 2, ran)
	assert.Greater(t, time.Millisecond*50, took)
	assert.Less(t, time.Second, took2)
	assert.Greater(t, time.Second+time.Millisecond*100, took2)
}

func TestEventLoopAllCallbacksGetCalled(t *testing.T) {
	t.Parallel()
	sleepTime := time.Millisecond * 500
	loop := NewEventLoop(NewTestVM(t).Runtime())
	var called int64
	f := func() error {
		for i := 0; i < 100; i++ {
			bad := i == 99
			r := loop.RegisterCallback()

			go func() {
				if !bad {
					time.Sleep(sleepTime)
				}
				r(func() error {
					if bad {
						return errors.New("something")
					}
					atomic.AddInt64(&called, 1)
					return nil
				})
			}()
		}
		return fmt.Errorf("expected")
	}
	for i := 0; i < 3; i++ {
		called = 0
		start := time.Now()
		assert.Error(t, loop.Start(f))
		took := time.Since(start)
		loop.WaitOnRegistered()
		took2 := time.Since(start)
		assert.Greater(t, time.Millisecond*50, took)
		assert.Less(t, sleepTime, took2)
		assert.Greater(t, sleepTime+time.Millisecond*100, took2)
		assert.EqualValues(t, called, 99)
	}
}

func TestEventLoopPanicOnDoubleCallback(t *testing.T) {
	t.Parallel()
	loop := NewEventLoop(NewTestVM(t).Runtime())
	var ran int
	f := func() error {
		ran++
		r := loop.RegisterCallback()
		go func() {
			time.Sleep(time.Second)
			r(func() error {
				ran++
				return nil
			})

			assert.Panics(t, func() { r(func() error { return nil }) })
		}()
		return nil
	}
	start := time.Now()
	assert.NoError(t, loop.Start(f))
	took := time.Since(start)
	assert.Equal(t, 2, ran)
	assert.Less(t, time.Second, took)
	assert.Greater(t, time.Second+time.Millisecond*100, took)
}

func TestEventLoopRejectUndefined(t *testing.T) {
	t.Parallel()
	vm := NewTestVM(t)
	loop := NewEventLoop(vm.Runtime())
	err := loop.Start(func() error {
		_, err := vm.Runtime().RunString("Promise.reject()")
		return err
	})
	loop.WaitOnRegistered()
	assert.EqualError(t, err, "Uncaught (in promise) undefined")
}

func TestEventLoopRejectString(t *testing.T) {
	t.Parallel()
	vm := NewTestVM(t)
	loop := NewEventLoop(vm.Runtime())
	err := loop.Start(func() error {
		_, err := vm.Runtime().RunString("Promise.reject('some string')")
		return err
	})
	loop.WaitOnRegistered()
	assert.EqualError(t, err, "Uncaught (in promise) some string")
}

func TestEventLoopRejectSyntaxError(t *testing.T) {
	t.Parallel()
	vm := NewTestVM(t)
	loop := NewEventLoop(vm.Runtime())
	err := loop.Start(func() error {
		_, err := vm.Runtime().RunString("Promise.resolve().then(()=> {some.syntax.error})")
		return err
	})
	loop.WaitOnRegistered()
	assert.EqualError(t, err, "Uncaught (in promise) ReferenceError: some is not defined\n\tat <eval>:1:30(1)\n")
}

func TestEventLoopRejectGoError(t *testing.T) {
	t.Parallel()
	vm := NewTestVM(t)
	loop := NewEventLoop(vm.Runtime())
	rt := vm.Runtime()
	assert.NoError(t, rt.Set("f", rt.ToValue(func() error {
		return errors.New("some error")
	})))
	err := loop.Start(func() error {
		_, err := vm.Runtime().RunString("Promise.resolve().then(()=> {f()})")
		return err
	})
	loop.WaitOnRegistered()
	assert.EqualError(t, err, "Uncaught (in promise) GoError: some error\n\tat github.com/shiroyk/cloudcat/js.TestEventLoopRejectGoError.func1 (native)\n\tat <eval>:1:31(2)\n")
}

func TestEventLoopRejectThrow(t *testing.T) {
	t.Parallel()
	vm := NewTestVM(t)
	loop := NewEventLoop(vm.Runtime())
	rt := vm.Runtime()
	assert.NoError(t, rt.Set("f", rt.ToValue(func() error {
		Throw(rt, errors.New("throw error"))
		return nil
	})))
	err := loop.Start(func() error {
		_, err := vm.Runtime().RunString("Promise.resolve().then(()=> {f()})")
		return err
	})
	loop.WaitOnRegistered()
	assert.EqualError(t, err, "Uncaught (in promise) throw error")
}

func TestEventLoopAsyncAwait(t *testing.T) {
	t.Parallel()
	vm := NewTestVM(t)
	loop := NewEventLoop(vm.Runtime())
	err := loop.Start(func() error {
		_, err := vm.Runtime().RunString(`
        async function a() {
            some.error.here
        }
        Promise.resolve().then(async () => {
            await a();
        })
        `)
		return err
	})
	loop.WaitOnRegistered()
	assert.EqualError(t, err, "Uncaught (in promise) ReferenceError: some is not defined\n\tat a (<eval>:3:13(1))\n\tat <eval>:6:20(2)\n")
}
