package main

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"sort"

	"github.com/dave/jennifer/jen"
	"github.com/immune-gmbh/guard/apisrv/v2/internal/jsonschema"
	"gopkg.in/yaml.v2"
)

func convert(i interface{}) interface{} {
	switch x := i.(type) {
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[k.(string)] = convert(v)
		}
		return m2
	case []interface{}:
		for i, v := range x {
			x[i] = convert(v)
		}
	}
	return i
}

func loadSchema(filename string) *jsonschema.Schema {
	fd, err := os.Open(filename)
	if err != nil {
		log.Fatalf("failed to open schema file: %v", err)
	}
	defer fd.Close()

	var schema jsonschema.Schema
	var file interface{}
	var dec *json.Decoder
	if yaml.NewDecoder(fd).Decode(&file) == nil {
		buf, err := json.Marshal(convert(file))
		if err != nil {
			log.Fatal(err)
		}
		dec = json.NewDecoder(bytes.NewReader(buf))
	} else {
		dec = json.NewDecoder(fd)
	}

	if err := dec.Decode(&schema); err != nil {
		log.Fatalf("failed to load schema JSON: %v", err)
	}

	return &schema
}

func main() {
	if len(os.Args) < 3 {
		log.Fatalf("%s: SCHEMA GO-FILE\n", os.Args[0])
	}
	schema := loadSchema(os.Args[1])
	f := jen.NewFile("issuesv1")
	symbolicName := make(map[string]bool)

	if schema.Ref == "" {
		jsonschema.GenerateDef(schema, schema, f, symbolicName, "root")
	}

	var names []string
	for name := range schema.Defs {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		def := schema.Defs[name]
		jsonschema.GenerateDef(&def, schema, f, symbolicName, name)

		for _, base := range def.AllOf {
			if base.Ref == "#/$defs/common" {
				jsonschema.GenerateInterface(&def, f, name)
			}
		}
	}

	if err := f.Save(os.Args[2]); err != nil {
		log.Fatalf("failed to save output file: %v", err)
	}
}
