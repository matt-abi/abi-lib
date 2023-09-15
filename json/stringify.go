package json

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/ability-sh/abi-lib/dynamic"
)

func stringifyObject(v reflect.Value, w *bytes.Buffer, idx int) (int, error) {

	var err error = nil
	ftype := v.Type()

	for i := 0; i < ftype.NumField(); i++ {
		fd := ftype.Field(i)
		fv := v.Field(i)

		var tags = strings.Split(fd.Tag.Get("json"), ",")

		if len(tags) > 0 && tags[0] != "" {

			if tags[0] == "-" {
				continue
			}

			if fv.CanInterface() {
				if len(tags) > 1 && tags[1] == "omitempty" && dynamic.IsNil(fv.Interface()) {
					continue
				}

				if idx != 0 {
					w.WriteString(",")
				}

				err = stringify(tags[0], w)

				if err != nil {
					return idx, err
				}

				w.WriteString(":")

				err = stringify(fv.Interface(), w)

				if err != nil {
					return idx, err
				}
				idx = idx + 1
			}

		} else if fv.Kind() == reflect.Struct {
			idx, err = stringifyObject(fv, w, idx)
			if err != nil {
				return idx, err
			}
		}

	}

	return idx, nil
}

type mapItem struct {
	key   string
	value reflect.Value
}

type mapItemList []*mapItem

func (m mapItemList) Len() int {
	return len(m)
}

func (m mapItemList) Less(i, j int) bool {
	return strings.Compare(m[i].key, m[j].key) < 0
}

func (m mapItemList) Swap(i, j int) {
	t := m[i]
	m[i] = m[j]
	m[j] = t
}

func stringify(object interface{}, w *bytes.Buffer) error {

	if object == nil {
		w.WriteString("null")
		return nil
	}

	v := reflect.ValueOf(object)

	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			w.WriteString("null")
			return nil
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Map:
		w.WriteString("{")
		var items mapItemList
		for _, key := range v.MapKeys() {
			vv := v.MapIndex(key)
			if key.CanInterface() && vv.CanInterface() {
				items = append(items, &mapItem{key: dynamic.StringValue(key.Interface(), ""), value: vv})
			}
		}
		if len(items) > 0 {
			sort.Sort(items)
			for i, item := range items {
				if i != 0 {
					w.WriteString(",")
				}
				err := stringify(item.key, w)
				if err != nil {
					return err
				}
				w.WriteString(":")
				err = stringify(item.value.Interface(), w)
				if err != nil {
					return err
				}
			}
		}
		w.WriteString("}")
	case reflect.Slice:
		w.WriteString("[")
		for i := 0; i < v.Len(); i++ {
			vv := v.Index(i)
			if vv.CanInterface() {
				if i != 0 {
					w.WriteString(",")
				}
				err := stringify(vv.Interface(), w)
				if err != nil {
					return err
				}
			}
		}
		w.WriteString("]")
	case reflect.Struct:
		w.WriteString("{")
		_, err := stringifyObject(v, w, 0)
		if err != nil {
			return err
		}
		w.WriteString("}")
	case reflect.Int64:
		{
			i64 := v.Int()
			i32 := int32(i64)
			if i64 == int64(i32) {
				b, err := json.Marshal(object)
				if err != nil {
					return err
				}
				w.Write(b)
			} else {
				b, err := json.Marshal(fmt.Sprintf("%d", i64))
				if err != nil {
					return err
				}
				w.Write(b)
			}
		}
	case reflect.Uint64:
		{
			i64 := v.Uint()
			i32 := uint32(i64)
			if i64 == uint64(i32) {
				b, err := json.Marshal(object)
				if err != nil {
					return err
				}
				w.Write(b)
			} else {
				b, err := json.Marshal(fmt.Sprintf("%d", i64))
				if err != nil {
					return err
				}
				w.Write(b)
			}
		}
	default:
		b, err := json.Marshal(object)
		if err != nil {
			return err
		}
		w.Write(b)
	}

	return nil
}

func Stringify(object interface{}) (string, error) {

	w := bytes.NewBuffer(nil)

	err := stringify(object, w)

	if err != nil {
		return "", err
	}

	return w.String(), nil
}
