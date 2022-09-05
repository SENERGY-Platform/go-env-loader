/*
   Copyright 2022 Yann Dumont

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package envldr

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
)

const tag = "env_var"

type parser func(t reflect.Type, val string) (interface{}, error)

var intBitSizeMap = map[reflect.Kind]int{
	reflect.Int:    0,
	reflect.Int16:  16,
	reflect.Int32:  32,
	reflect.Int64:  64,
	reflect.Uint:   0,
	reflect.Uint16: 16,
	reflect.Uint32: 32,
	reflect.Uint64: 64,
}

var floatBitSizeMap = map[reflect.Kind]int{
	reflect.Float32: 32,
	reflect.Float64: 64,
}

var complexBitSizeMap = map[reflect.Kind]int{
	reflect.Complex64:  64,
	reflect.Complex128: 128,
}

var intParser parser = func(t reflect.Type, val string) (interface{}, error) {
	i, err := strconv.ParseInt(val, 10, intBitSizeMap[t.Kind()])
	if t.Kind() == reflect.Int64 {
		return i, err
	} else {
		return reflect.ValueOf(i).Convert(t).Interface(), err
	}
}

var uintParser parser = func(t reflect.Type, val string) (interface{}, error) {
	i, err := strconv.ParseUint(val, 10, intBitSizeMap[t.Kind()])
	if t.Kind() == reflect.Uint64 {
		return i, err
	} else {
		return reflect.ValueOf(i).Convert(t).Interface(), err
	}
}

var floatParser parser = func(t reflect.Type, val string) (interface{}, error) {
	f, err := strconv.ParseFloat(val, floatBitSizeMap[t.Kind()])
	if t.Kind() == reflect.Float64 {
		return f, err
	} else {
		return reflect.ValueOf(f).Convert(t).Interface(), err
	}
}

var complexParser parser = func(t reflect.Type, val string) (interface{}, error) {
	c, err := strconv.ParseComplex(val, complexBitSizeMap[t.Kind()])
	if t.Kind() == reflect.Complex128 {
		return c, err
	} else {
		return reflect.ValueOf(c).Convert(t).Interface(), err
	}
}

var jsonParser parser = func(t reflect.Type, val string) (interface{}, error) {
	v := reflect.New(t)
	err := json.Unmarshal([]byte(val), v.Interface())
	return v.Interface(), err
}

var parsers = map[reflect.Kind]parser{
	reflect.Uint:       uintParser,
	reflect.Uint8:      uintParser,
	reflect.Uint16:     uintParser,
	reflect.Uint32:     uintParser,
	reflect.Uint64:     uintParser,
	reflect.Int:        intParser,
	reflect.Int8:       intParser,
	reflect.Int16:      intParser,
	reflect.Int32:      intParser,
	reflect.Int64:      intParser,
	reflect.Float32:    floatParser,
	reflect.Float64:    floatParser,
	reflect.Complex64:  complexParser,
	reflect.Complex128: complexParser,
	reflect.Bool: func(t reflect.Type, val string) (interface{}, error) {
		return strconv.ParseBool(val)
	},
	reflect.String: func(t reflect.Type, val string) (interface{}, error) {
		return val, nil
	},
	reflect.Slice:  jsonParser,
	reflect.Map:    jsonParser,
	reflect.Struct: jsonParser,
}

func getEnv(st reflect.StructField) (val string, ok bool) {
	if val, ok = st.Tag.Lookup(tag); ok {
		val, ok = os.LookupEnv(val)
	}
	return
}

func loadEnv(v reflect.Value) error {
	for i := 0; i < v.Type().NumField(); i++ {
		structField := v.Type().Field(i)
		if structField.PkgPath == "" {
			fieldValue := v.Field(i)
			isNilPtr := false
			if fieldValue.Kind() == reflect.Ptr {
				if fieldValue.IsNil() {
					isNilPtr = true
				} else {
					fieldValue = fieldValue.Elem()
				}
			}
			if envVal, ok := getEnv(structField); ok {
				fieldType := fieldValue.Type()
				if isNilPtr {
					fieldType = fieldValue.Type().Elem()
					fieldValue.Set(reflect.New(fieldType))
					fieldValue = fieldValue.Elem()
				}
				if p, k := parsers[fieldType.Kind()]; k {
					if itf, err := p(fieldType, envVal); err != nil {
						return err
					} else {
						itfValue := reflect.Indirect(reflect.ValueOf(itf))
						fieldValue.Set(itfValue)
					}
				}

			} else {
				if isNilPtr && fieldValue.Type().Elem().Kind() == reflect.Struct {
					var hasEnvVal bool
					for x := 0; x < fieldValue.Type().Elem().NumField(); x++ {
						st := fieldValue.Type().Elem().Field(x)
						if _, k := getEnv(st); k {
							hasEnvVal = true
							break
						}
					}
					if hasEnvVal {
						fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
						fieldValue = fieldValue.Elem()
					}
				}
				if fieldValue.Kind() == reflect.Struct {
					if err := loadEnv(fieldValue); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func LoadEnv(itf interface{}) error {
	if v := reflect.ValueOf(itf); v.Kind() == reflect.Ptr {
		if v = v.Elem(); v.Kind() == reflect.Struct {
			return loadEnv(v)
		} else {
			panic(fmt.Sprintf("'%s' provided but '%s' required", v.Kind(), reflect.Struct))
		}
	} else {
		panic(fmt.Sprintf("'%s' provided but '%s' required", v.Kind(), reflect.Ptr))
	}
}
