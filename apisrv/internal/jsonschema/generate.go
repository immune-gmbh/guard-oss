package jsonschema

import (
	"log"
	"sort"
	"strings"
	"unicode"

	"github.com/dave/jennifer/jen"
)

func formatId(s string) string {
	fields := strings.FieldsFunc(s, func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	})
	for i, v := range fields {
		fields[i] = strings.Title(v)
	}
	return strings.Join(fields, "")
}

func refName(ref string) string {
	prefix := "#/$defs/"
	if !strings.HasPrefix(ref, prefix) {
		return ""
	}
	return strings.TrimPrefix(ref, prefix)
}

func resolveRef(def *Schema, root *Schema) *Schema {
	if def.Ref == "" {
		return def
	}

	name := refName(def.Ref)
	if name == "" {
		log.Fatalf("unsupported $ref %q", def.Ref)
	}

	result, ok := root.Defs[name]
	if !ok {
		log.Fatalf("invalid $ref %q", def.Ref)
	}
	return &result
}

func schemaType(schema *Schema) Type {
	switch {
	case len(schema.Type) == 1:
		return schema.Type[0]
	case len(schema.Type) > 0:
		return ""
	}

	var v interface{}
	if schema.Const != nil {
		v = schema.Const
	} else if len(schema.Enum) > 0 {
		v = schema.Enum[0]
	}

	switch v.(type) {
	case bool:
		return TypeBoolean
	case map[string]interface{}:
		return TypeObject
	case []interface{}:
		return TypeArray
	case float64:
		return TypeNumber
	case string:
		return TypeString
	default:
		return ""
	}
}

func isRequired(schema *Schema, propName string) bool {
	for _, name := range schema.Required {
		if name == propName {
			return true
		}
	}
	return false
}

func generateStruct(schema *Schema, root *Schema) jen.Code {
	var fields []jen.Code
	var names []string
	var embedded map[string]bool = map[string]bool{}

	// embdedded structs
	for prop, sch := range schema.Properties {
		if prop == "" {
			ref := refName(sch.Ref)
			if ref != "" {
				for n := range resolveRef(&sch, root).Properties {
					embedded[n] = true
				}
			}
			fields = append(fields, generateSchemaType(&sch, root, true))
		}
	}

	for name := range schema.Properties {
		if name != "" && embedded[name] == false {
			names = append(names, name)
		}
	}
	sort.Strings(names)

	for _, name := range names {
		prop := schema.Properties[name]
		required := isRequired(schema, name)
		id := formatId(name)
		t := generateSchemaType(&prop, root, required)
		jsonTag := name
		if !required {
			jsonTag += ",omitempty"
		}
		tags := map[string]string{"json": jsonTag}
		fields = append(fields, jen.Id(id).Add(t).Tag(tags))
	}
	return jen.Struct(fields...)
}

func singlePatternProp(schema *Schema) *Schema {
	if len(schema.PatternProperties) != 1 {
		return nil
	}
	for _, prop := range schema.PatternProperties {
		return &prop
	}
	return nil
}

func noAdditionalProps(schema *Schema) bool {
	return schema.AdditionalProperties != nil && schema.AdditionalProperties.IsFalse()
}

func generateSchemaConstants(schema *Schema, root *Schema, stem string, symbolicNames map[string]bool) []jen.Code {
	if schema == nil {
		return nil
	}
	if refName(schema.Ref) != "" {
		schema = resolveRef(schema, root)
	}
	ty := schemaType(schema)
	if ty == TypeObject {
		var ret []jen.Code
		for prop, sch := range schema.Properties {
			ret = append(ret, generateSchemaConstants(&sch, root, stem+formatId(prop), symbolicNames)...)
		}
		return ret
	}
	if ty != TypeString && ty != TypeNumber && ty != TypeInteger && ty != TypeBoolean {
		return nil
	}

	// generate symbolic constants
	consts := make(map[string]interface{})
	if schema.Const != nil {
		consts[stem] = schema.Const
	} else if len(schema.Enum) > 1 {
		for _, e := range schema.Enum {
			if str, ok := e.(string); ok {
				name := formatId(str)
				consts[name] = e
			}
		}
	} else if len(schema.Enum) == 1 {
		consts[stem] = schema.Enum[0]
	}

	var ret []jen.Code
	for n, k := range consts {
		if _, ok := symbolicNames[n]; ok {
			continue
		}
		symbolicNames[n] = true

		switch ty {
		case TypeBoolean:
			ret = append(ret, jen.Const().Id(n).Bool().Op("=").Lit(k).Line())
		case TypeNumber:
			ret = append(ret, jen.Const().Id(n).Float64().Op("=").Lit(k).Line())
		case TypeInteger:
			ret = append(ret, jen.Const().Id(n).Int64().Op("=").Lit(k).Line())
		case TypeString:
			ret = append(ret, jen.Const().Id(n).String().Op("=").Lit(k).Line())
		}
	}
	return ret
}

func generateSchemaType(schema *Schema, root *Schema, required bool) jen.Code {
	if schema == nil {
		return jen.Interface()
	}

	refName := refName(schema.Ref)
	if refName != "" {
		schema = resolveRef(schema, root)
		t := jen.Id(formatId(refName))
		if !required && schemaType(schema) == TypeObject && noAdditionalProps(schema) && len(schema.PatternProperties) == 0 {
			t = jen.Op("*").Add(t)
		}
		return t
	}

	switch schemaType(schema) {
	case TypeNull:
		return jen.Struct()
	case TypeBoolean:
		return jen.Bool()
	case TypeArray:
		return jen.Index().Add(generateSchemaType(schema.Items, root, required))
	case TypeNumber:
		return jen.Float32()
	case TypeString:
		return jen.String()
	case TypeInteger:
		return jen.Int64()
	case TypeObject:
		noAdditionalProps := noAdditionalProps(schema)
		if noAdditionalProps && len(schema.PatternProperties) == 0 {
			t := generateStruct(schema, root)
			if !required {
				t = jen.Op("*").Add(t)
			}
			return t
		} else if patternProp := singlePatternProp(schema); noAdditionalProps && patternProp != nil {
			return jen.Map(jen.String()).Add(generateSchemaType(patternProp, root, true))
		} else {
			return jen.Map(jen.String()).Add(generateSchemaType(schema.AdditionalProperties, root, true))
		}
	default:
		return jen.Qual("encoding/json", "RawMessage")
	}
}

// transform
//   {
//     $id: blub
//     allOf: [ { $ref: A }, { $ref: B } ]
//   }
// to
//   type Blub struct {
//     A
//     B
//   }
// iff
//   A and B are structs
//
//return true if applied
func applyObjCompositionTransform(schema *Schema, root *Schema, f *jen.File, symbolicNames map[string]bool, id string) bool {
	objComposition :=
		len(schema.AllOf) > 0 &&
			len(schema.AnyOf) == 0 &&
			len(schema.OneOf) == 0 &&
			len(schema.DependentSchemas) == 0 &&
			schema.Then == nil &&
			schema.Else == nil
	if !objComposition {
		return false
	}

	for _, child := range schema.AllOf {
		if child.Ref != "" {
			child = *resolveRef(&child, root)
		}
		if schemaType(&child) != TypeObject {
			return false
		}
	}

	obj := Schema{
		Type:                 TypeSet{TypeObject},
		Properties:           make(map[string]Schema),
		AdditionalProperties: &Schema{Not: []Schema{{}}},
	}
	seenProps := make(map[string]bool)
	for _, super := range schema.AllOf {
		if refName(super.Ref) != "" {
			// HACK: signal struct embedding by using the empty string for
			// the field name. See generateSchemaType()
			obj.Properties[""] = super
			obj.Required = append(obj.Required, "")
			super = *resolveRef(&super, root)
		} else {
			// merge properties
			for prop, sch := range super.Properties {
				obj.Required = append(obj.Required, prop)

				if _, ok := seenProps[prop]; !ok {
					obj.Properties[prop] = sch
				} else {
					// merge symbolic constants
					objp := obj.Properties[prop]
					objp.Enum = append(objp.Enum, sch.Enum...)
					if sch.Const != nil {
						objp.Enum = append(objp.Enum, sch.Const)
					}
					obj.Properties[prop] = objp
				}
			}
		}
		if noAdditionalProps(&obj) {
			obj.AdditionalProperties = super.AdditionalProperties
		}
		for prop := range super.Properties {
			seenProps[prop] = true
		}
	}

	f.Type().Id(id).Add(generateSchemaType(&obj, root, true)).Line()
	f.Add(generateSchemaConstants(&obj, root, id, symbolicNames)...)
	return true
}

func GenerateDef(schema *Schema, root *Schema, f *jen.File, symbolicNames map[string]bool, name string) {
	id := formatId(name)

	if schema.Ref == "" && schemaType(schema) == "" {
		if applyObjCompositionTransform(schema, root, f, symbolicNames, id) {
			return
		}

		f.Type().Id(id).Struct(
			jen.Qual("encoding/json", "RawMessage"),
		).Line()

		var children []Schema
		for _, child := range schema.AllOf {
			children = append(children, child)
		}
		for _, child := range schema.AnyOf {
			children = append(children, child)
		}
		for _, child := range schema.OneOf {
			children = append(children, child)
		}
		if schema.Then != nil {
			children = append(children, *schema.Then)
		}
		if schema.Else != nil {
			children = append(children, *schema.Else)
		}
		for _, child := range schema.DependentSchemas {
			children = append(children, child)
		}

		for _, child := range children {
			refName := refName(child.Ref)
			if refName == "" {
				continue
			}

			t := generateSchemaType(&child, root, false)

			f.Func().Params(
				jen.Id("v").Id(id),
			).Id(formatId(refName)).Params().Params(
				t,
				jen.Id("error"),
			).Block(
				jen.Var().Id("out").Add(t),
				jen.Id("err").Op(":=").Qual("encoding/json", "Unmarshal").Params(
					jen.Id("v").Op(".").Id("RawMessage"),
					jen.Op("&").Id("out"),
				),
				jen.Return(
					jen.Id("out"),
					jen.Id("err"),
				),
			).Line()
		}
	} else {
		f.Type().Id(id).Add(generateSchemaType(schema, root, true)).Line()
	}
}

func GenerateInterface(def *Schema, f *jen.File, nam string) error {
	// func (Issue) Id() string
	f.Add(jen.Func().Params(
		jen.Id("i").Op("*").Id(formatId(nam)),
	).Id("Id").Params().Params(
		jen.String(),
	).Block(
		jen.Return(jen.Id("i").Op(".").Id("Common").Op(".").Id("Id")),
	).Line())

	// func (Issue) Incident() bool
	f.Add(jen.Func().Params(
		jen.Id("i").Op("*").Id(formatId(nam)),
	).Id("Incident").Params().Params(
		jen.Bool(),
	).Block(
		jen.Return(jen.Id("i").Op(".").Id("Common").Op(".").Id("Incident")),
	).Line())

	// func (Issue) Aspect() bool
	f.Add(jen.Func().Params(
		jen.Id("i").Op("*").Id(formatId(nam)),
	).Id("Aspect").Params().Params(
		jen.String(),
	).Block(
		jen.Return(jen.Id("i").Op(".").Id("Common").Op(".").Id("Aspect")),
	).Line())

	return nil
}
