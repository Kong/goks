package vm

import (
	"bytes"
	"embed"
	"testing"

	"github.com/kong/goks"
	"github.com/stretchr/testify/assert"
	lua "github.com/yuin/gopher-lua"
)

//go:embed setup.go
var invalidLuaTree embed.FS

func TestNew(t *testing.T) {
	t.Run("instantiate standard VM", func(t *testing.T) {
		vm, err := New(Opts{})
		assert.NotNil(t, vm)
		assert.Nil(t, err)
	})

	t.Run("invalid VM instantiation with injected overrides", func(t *testing.T) {
		vm, err := New(Opts{InjectFS: &invalidLuaTree})
		assert.Nil(t, vm)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "file system must contain 'lua-tree/share/lua/5.1'")
	})

	t.Run("VM instantiation with injected overrides", func(t *testing.T) {
		vm, err := New(Opts{InjectFS: &goks.LuaTree})
		assert.NotNil(t, vm)
		assert.Nil(t, err)
	})
}

func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		New(Opts{})
	}
}

var testlua = `
_G["good_fn"] = function()
  return "foo", "bar"
end
_G["throw_err_fn"] = function()
  error("i like errors")
end
`

func TestVM_CallByParams(t *testing.T) {
	vm, err := New(Opts{})
	assert.NotNil(t, vm)
	assert.Nil(t, err)
	buf := bytes.NewBufferString(testlua)
	assert.Nil(t, vm.Execute(buf, "test"))
	t.Run("returns string params correctly", func(t *testing.T) {
		ret, err := vm.CallByParams("good_fn")
		assert.Equal(t, "foo", ret)
		assert.Equal(t, "bar", err.Error())
	})
	t.Run("returns error when nil function is invoked", func(t *testing.T) {
		ret, err := vm.CallByParams("does_not_exist_fn")
		assert.Equal(t, "", ret)
		assert.NotNil(t, err)
	})
	t.Run("returns error when error is returned by the function", func(t *testing.T) {
		ret, err := vm.CallByParams("throw_err_fn")
		assert.Equal(t, "", ret)
		assert.IsType(t, &lua.ApiError{}, err)
	})
	t.Run("returns error when function returns", func(t *testing.T) {
		ret, err := vm.CallByParams("throw_err_fn")
		assert.Equal(t, "", ret)
		assert.IsType(t, &lua.ApiError{}, err)
	})
}
