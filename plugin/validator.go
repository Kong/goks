package plugin

import (
	"github.com/kong/goks/internal/vm"
)

type Validator struct {
	vm *vm.VM
}

func NewValidator() (*Validator, error) {
	vm, err := vm.New()
	if err != nil {
		return nil, err
	}
	return &Validator{vm: vm}, nil
}

func (v *Validator) LoadSchema(schema string) error {
	_, err := v.vm.CallByParams("load_plugin_schema", schema)
	return err
}

func (v *Validator) Validate(pluginInstance string) (string, error) {
	return v.vm.CallByParams("validate", pluginInstance)
}

func (v *Validator) ProcessAutoFields(pluginInstance string) (string, error) {
	return v.vm.CallByParams("process_auto_fields", pluginInstance)
}
