package shouchan

import (
	"reflect"
	"strings"

	"github.com/hujun-open/extyaml"
	"gopkg.in/yaml.v3"
)

func reflectValue[X any](p *X) reflect.Value {
	return reflect.ValueOf(p).Elem()
}

// cloneInto copies src into dst. An existing destination pointer is kept and
// the value is written through it, so flag bindings captured by myflags stay
// attached to the same objects.
func cloneInto(dst, src reflect.Value) {
	if !src.IsValid() {
		return
	}
	if dst.Kind() == reflect.Pointer || src.Kind() == reflect.Pointer {
		clonePointer(dst, src)
		return
	}
	switch src.Kind() {
	case reflect.Struct:
		if structNeedsWalk(src.Type()) {
			for i := 0; i < src.NumField(); i++ {
				if !src.Type().Field(i).IsExported() {
					continue
				}
				cloneInto(dst.Field(i), src.Field(i))
			}
			return
		}
		if dst.CanSet() {
			dst.Set(src)
		}
	case reflect.Slice:
		cloneSlice(dst, src)
	case reflect.Map:
		cloneMap(dst, src)
	case reflect.Array:
		if typeNeedsWalk(src.Type().Elem()) {
			for i := 0; i < src.Len(); i++ {
				cloneInto(dst.Index(i), src.Index(i))
			}
			return
		}
		if dst.CanSet() {
			dst.Set(src)
		}
	default:
		if dst.CanSet() && src.Type().AssignableTo(dst.Type()) {
			dst.Set(src)
		}
	}
}

func clonePointer(dst, src reflect.Value) {
	if src.Kind() != reflect.Pointer {
		if dst.Kind() == reflect.Pointer {
			if dst.IsNil() {
				if !dst.CanSet() {
					return
				}
				dst.Set(reflect.New(dst.Type().Elem()))
			}
			cloneInto(dst.Elem(), src)
		}
		return
	}
	if src.IsNil() {
		if dst.Kind() == reflect.Pointer && dst.CanSet() {
			dst.Set(reflect.Zero(dst.Type()))
		}
		return
	}
	if dst.Kind() != reflect.Pointer {
		cloneInto(dst, src.Elem())
		return
	}
	if dst.IsNil() {
		if !dst.CanSet() {
			return
		}
		dst.Set(reflect.New(dst.Type().Elem()))
	}
	cloneInto(dst.Elem(), src.Elem())
}

func cloneSlice(dst, src reflect.Value) {
	if !dst.CanSet() {
		return
	}
	if src.IsNil() {
		dst.Set(reflect.Zero(dst.Type()))
		return
	}
	cloned := reflect.MakeSlice(src.Type(), src.Len(), src.Len())
	for i := 0; i < src.Len(); i++ {
		cloneInto(cloned.Index(i), src.Index(i))
	}
	dst.Set(cloned)
}

func cloneMap(dst, src reflect.Value) {
	if !dst.CanSet() {
		return
	}
	if src.IsNil() {
		dst.Set(reflect.Zero(dst.Type()))
		return
	}
	cloned := reflect.MakeMapWithSize(src.Type(), src.Len())
	iter := src.MapRange()
	for iter.Next() {
		k := reflect.New(src.Type().Key()).Elem()
		v := reflect.New(src.Type().Elem()).Elem()
		cloneInto(k, iter.Key())
		cloneInto(v, iter.Value())
		cloned.SetMapIndex(k, v)
	}
	dst.Set(cloned)
}

func structNeedsWalk(t reflect.Type) bool {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		if typeNeedsWalk(f.Type) {
			return true
		}
	}
	return false
}

func typeNeedsWalk(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Map:
		return true
	case reflect.Array:
		return typeNeedsWalk(t.Elem())
	case reflect.Struct:
		return structNeedsWalk(t)
	default:
		return false
	}
}

// mergeYAML copies fields present in buf from src onto dst.
// src must already be the extyaml decoding of buf, so slices and maps in src
// contain only the YAML values. Omitted fields of dst are left unchanged.
func mergeYAML(dst, src reflect.Value, buf []byte) error {
	var doc yaml.Node
	if err := yaml.Unmarshal(buf, &doc); err != nil {
		return err
	}
	node := unwrapYAML(&doc)
	if node == nil || node.Kind == 0 {
		return nil
	}
	d := followAlloc(dst)
	s := follow(src)
	if node.Kind == yaml.MappingNode && d.IsValid() && d.Kind() == reflect.Struct && s.IsValid() && s.Kind() == reflect.Struct {
		return mergeMapping(d, s, node)
	}
	cloneInto(dst, src)
	return nil
}

func mergeMapping(dst, src reflect.Value, node *yaml.Node) error {
	for i := 0; i+1 < len(node.Content); i += 2 {
		key := unwrapYAML(node.Content[i])
		if key == nil {
			continue
		}
		df, sf, ok := findYAMLField(dst, src, key.Value)
		if !ok {
			continue
		}
		if err := mergeField(df, sf, node.Content[i+1]); err != nil {
			return err
		}
	}
	return nil
}

func mergeField(dst, src reflect.Value, node *yaml.Node) error {
	node = unwrapYAML(node)
	if node == nil || node.Kind == 0 {
		cloneInto(dst, src)
		return nil
	}
	d := followAlloc(dst)
	s := follow(src)
	if node.Kind == yaml.MappingNode && d.IsValid() && d.Kind() == reflect.Struct && s.IsValid() && s.Kind() == reflect.Struct {
		return mergeMapping(d, s, node)
	}
	cloneInto(dst, src)
	return nil
}

func follow(v reflect.Value) reflect.Value {
	for v.IsValid() && v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return reflect.Value{}
		}
		v = v.Elem()
	}
	return v
}

func followAlloc(v reflect.Value) reflect.Value {
	for v.IsValid() && v.Kind() == reflect.Pointer {
		if v.IsNil() {
			if !v.CanSet() {
				return reflect.Value{}
			}
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}
	return v
}

func findYAMLField(dst, src reflect.Value, key string) (reflect.Value, reflect.Value, bool) {
	if !dst.IsValid() || dst.Kind() != reflect.Struct || !src.IsValid() || src.Kind() != reflect.Struct {
		return reflect.Value{}, reflect.Value{}, false
	}
	t := dst.Type()
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if !sf.IsExported() {
			continue
		}
		name, inline, skip := yamlFieldInfo(sf)
		if skip {
			continue
		}
		if inline {
			dv, sv, ok := findYAMLField(followAlloc(dst.Field(i)), follow(src.Field(i)), key)
			if ok {
				return dv, sv, true
			}
			continue
		}
		if strings.EqualFold(name, key) {
			return dst.Field(i), src.Field(i), true
		}
	}
	return reflect.Value{}, reflect.Value{}, false
}

func yamlFieldInfo(sf reflect.StructField) (name string, inline, skip bool) {
	if _, ok := sf.Tag.Lookup(extyaml.SkipTag); ok {
		return "", false, true
	}
	tag := sf.Tag.Get("yaml")
	if tag == "-" {
		return "", false, true
	}
	parts := strings.Split(tag, ",")
	if parts[0] != "" {
		name = parts[0]
	}
	for _, p := range parts[1:] {
		if strings.TrimSpace(p) == "inline" {
			inline = true
		}
	}
	if sf.Anonymous && name == "" {
		inline = true
	}
	if inline {
		return name, true, false
	}
	if name == "" {
		name = strings.ToLower(sf.Name)
	}
	return name, false, false
}

func unwrapYAML(n *yaml.Node) *yaml.Node {
	if n == nil {
		return nil
	}
	if n.Kind == yaml.AliasNode && n.Alias != nil {
		return unwrapYAML(n.Alias)
	}
	if n.Kind == yaml.DocumentNode {
		if len(n.Content) == 0 {
			return nil
		}
		return unwrapYAML(n.Content[0])
	}
	return n
}
