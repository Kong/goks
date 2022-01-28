package main

import (
	"flag"
	"io/ioutil"
	"log"

	"github.com/kong/goks/plugin"
)

var (
	schemaFile = flag.String("schema-file", "schema.lua", "schema to use")
	pluginFile = flag.String("plugin-file", "plugin.json", "plugin to verify")
)

func main() {
	flag.Parse()
	validator, err := plugin.NewValidator()
	if err != nil {
		log.Fatalln("failed to create a VM:", err)
	}
	schema, err := ioutil.ReadFile(*schemaFile)
	if err != nil {
		log.Fatalln("failed to read schema file:", err)
	}
	_, err = validator.LoadSchema(string(schema))
	if err != nil {
		log.Fatalln("failed to load schema:", err)
	}

	p, err := ioutil.ReadFile(*pluginFile)
	if err != nil {
		log.Fatalln("failed to read plugin file:", err)
	}
	plugin := string(p)

	plugin, err = validator.ProcessAutoFields(plugin)
	if err != nil {
		log.Fatalln("failed to populate defaults:", err)
	}

	err = validator.Validate(plugin)
	if err != nil {
		log.Fatalln("failed to validate plugin:", err)
	}
}
