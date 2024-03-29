package js

import (
	"context"
	"errors"
	"fmt"

	"github.com/dop251/goja"
	"github.com/shiroyk/cloudcat/plugin/jsmodule"
	"github.com/spf13/cast"
)

// vmContextKey the VM current context
var vmContextKey = goja.NewSymbol("__ctx__")

// Throw js exception
func Throw(vm *goja.Runtime, err error) {
	if e, ok := err.(*goja.Exception); ok { //nolint:errorlint
		panic(e)
	}
	panic(vm.ToValue(err))
}

// ToBytes tries to return a byte slice from compatible types.
func ToBytes(data any) ([]byte, error) {
	switch dt := data.(type) {
	case []byte:
		return dt, nil
	case string:
		return []byte(dt), nil
	case goja.ArrayBuffer:
		return dt.Bytes(), nil
	default:
		return nil, fmt.Errorf("invalid type %T, expected string, []byte or ArrayBuffer", data)
	}
}

// ToStrings tries to return a string slice or string from compatible types.
func ToStrings(data any) (s any, err error) {
	switch dt := data.(type) {
	case string:
		return dt, nil
	case []string:
		return dt, nil
	case []byte:
		return string(dt), nil
	case []any:
		return cast.ToStringSliceE(dt)
	case goja.ArrayBuffer:
		return string(dt.Bytes()), nil
	default:
		return nil, fmt.Errorf("invalid type %T, expected string, string array or ArrayBuffer", data)
	}
}

// Unwrap the goja.Value to the raw value
func Unwrap(value goja.Value) (any, error) {
	if value == nil {
		return nil, nil
	}
	switch v := value.Export().(type) {
	default:
		return v, nil
	case goja.ArrayBuffer:
		return v.Bytes(), nil
	case *goja.Promise:
		switch v.State() {
		case goja.PromiseStateRejected:
			return nil, errors.New(v.Result().String())
		case goja.PromiseStateFulfilled:
			return v.Result().Export(), nil
		default:
			return nil, errors.New("unexpected promise pending state")
		}
	}
}

// VMContext returns the current context of the goja.Runtime
func VMContext(runtime *goja.Runtime) context.Context {
	ctx := context.Background()
	if v := runtime.GlobalObject().GetSymbol(vmContextKey); v != nil {
		if c, ok := v.Export().(context.Context); ok {
			ctx = c
		}
	}
	return ctx
}

// InitGlobalModule init all global modules
func InitGlobalModule(runtime *goja.Runtime) {
	// Init global modules
	for _, extension := range jsmodule.AllModules() {
		if mod, ok := extension.Module.(jsmodule.Global); ok {
			_ = runtime.Set(extension.Name, mod.Exports())
		}
	}
}
