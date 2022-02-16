package plugin

import (
	"embed"

	"github.com/kong/goks/internal/vm"
)

type Validator struct {
	vm *vm.VM
}

type ValidatorOpts struct {
	InjectFS *embed.FS
}

func NewValidator(opts ValidatorOpts) (*Validator, error) {
	vm, err := vm.New(vm.Opts{InjectFS: opts.InjectFS})
	if err != nil {
		return nil, err
	}
	return &Validator{vm: vm}, nil
}

func (v *Validator) LoadSchema(schema string) (string, error) {
	pluginName, err := v.vm.CallByParams("load_plugin_schema", schema)
	return pluginName, err
}

func (v *Validator) Validate(pluginInstance string) error {
	_, err := v.vm.CallByParams("validate", pluginInstance)
	return err
}

func (v *Validator) ProcessAutoFields(pluginInstance string) (string, error) {
	return v.vm.CallByParams("process_auto_fields", pluginInstance)
}

func (v *Validator) SchemaAsJSON(schemaName string) (string, error) {
	return v.vm.CallByParams("schema_as_json", schemaName)
}
