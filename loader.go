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
	"strings"
)

const varTag = "env_var"
const parserTag = "env_parser"
const paramsTag = "env_params"
const separator = ";"
const equal = "="

type Parser func(t reflect.Type, val string, params []string, kwParams map[string]string) (interface{}, error)

var bitSizeMap = map[reflect.Kind]int{
	reflect.Int:        0,
	reflect.Int16:      16,
	reflect.Int32:      32,
	reflect.Int64:      64,
	reflect.Uint:       0,
	reflect.Uint16:     16,
	reflect.Uint32:     32,
	reflect.Uint64:     64,
	reflect.Float32:    32,
	reflect.Float64:    64,
	reflect.Complex64:  64,
	reflect.Complex128: 128,
}

var intParser Parser = func(t reflect.Type, val string, params []string, kwParams map[string]string) (interface{}, error) {
	i, err := strconv.ParseInt(val, 10, bitSizeMap[t.Kind()])
	if t.Kind() == reflect.Int64 {
		return i, err
	} else {
		return reflect.ValueOf(i).Convert(t).Interface(), err
	}
}

var uintParser Parser = func(t reflect.Type, val string, params []string, kwParams map[string]string) (interface{}, error) {
	i, err := strconv.ParseUint(val, 10, bitSizeMap[t.Kind()])
	if t.Kind() == reflect.Uint64 {
		return i, err
	} else {
		return reflect.ValueOf(i).Convert(t).Interface(), err
	}
}

var floatParser Parser = func(t reflect.Type, val string, params []string, kwParams map[string]string) (interface{}, error) {
	f, err := strconv.ParseFloat(val, bitSizeMap[t.Kind()])
	if t.Kind() == reflect.Float64 {
		return f, err
	} else {
		return reflect.ValueOf(f).Convert(t).Interface(), err
	}
}

var complexParser Parser = func(t reflect.Type, val string, params []string, kwParams map[string]string) (interface{}, error) {
	c, err := strconv.ParseComplex(val, bitSizeMap[t.Kind()])
	if t.Kind() == reflect.Complex128 {
		return c, err
	} else {
		return reflect.ValueOf(c).Convert(t).Interface(), err
	}
}

var jsonParser Parser = func(t reflect.Type, val string, params []string, kwParams map[string]string) (interface{}, error) {
	v := reflect.New(t)
	err := json.Unmarshal([]byte(val), v.Interface())
	return v.Interface(), err
}

var parsers = map[reflect.Kind]Parser{
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
	reflect.Bool: func(t reflect.Type, val string, params []string, kwParams map[string]string) (interface{}, error) {
		return strconv.ParseBool(val)
	},
	reflect.String: func(t reflect.Type, val string, params []string, kwParams map[string]string) (interface{}, error) {
		return val, nil
	},
	reflect.Slice:  jsonParser,
	reflect.Map:    jsonParser,
	reflect.Struct: jsonParser,
}

func getEnv(st reflect.StructField) (val string, parserKw string, params []string, kwParams map[string]string, ok bool) {
	if val, ok = st.Tag.Lookup(varTag); ok && val != "" {
		val, ok = os.LookupEnv(val)
		if psr, k := st.Tag.Lookup(parserTag); k && psr != "" {
			parserKw = psr
		}
		if prms, k := st.Tag.Lookup(paramsTag); k && prms != "" {
			parts := strings.Split(prms, separator)
			for _, v := range parts {
				if strings.Contains(v, equal) {
					if kwParams == nil {
						kwParams = make(map[string]string)
					}
					kp := strings.Split(v, equal)
					kwParams[kp[0]] = kp[1]
				} else {
					params = append(params, v)
				}
			}
		}
	}
	return
}

func getParser(kwParsers map[string]Parser, typeParsers map[reflect.Type]Parser, kindParsers map[reflect.Kind]Parser, parserKw string, fType reflect.Type) (parser Parser, ok bool) {
	if parserKw != "" && kwParsers != nil {
		if parser, ok = kwParsers[parserKw]; ok {
			return
		}
	}
	if typeParsers != nil {
		if parser, ok = typeParsers[fType]; ok {
			return
		}
	}
	if kindParsers != nil {
		if parser, ok = kindParsers[fType.Kind()]; ok {
			return
		}
	}
	if parser, ok = parsers[fType.Kind()]; ok {
		return
	}
	return
}

func loadEnv(v reflect.Value, kwParsers map[string]Parser, typeParsers map[reflect.Type]Parser, kindParsers map[reflect.Kind]Parser) error {
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
			if envVal, parserKw, params, kwParams, ok := getEnv(structField); ok {
				fieldType := fieldValue.Type()
				if isNilPtr {
					fieldType = fieldValue.Type().Elem()
					fieldValue.Set(reflect.New(fieldType))
					fieldValue = fieldValue.Elem()
				}
				if p, k := getParser(kwParsers, typeParsers, kindParsers, parserKw, fieldType); k {
					if itf, err := p(fieldType, envVal, params, kwParams); err != nil {
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
						if _, _, _, _, k := getEnv(st); k {
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
					if err := loadEnv(fieldValue, kwParsers, typeParsers, kindParsers); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func LoadEnvUserParser(itf interface{}, keywordParsers map[string]Parser, typeParsers map[reflect.Type]Parser, kindParsers map[reflect.Kind]Parser) error {
	if v := reflect.ValueOf(itf); v.Kind() == reflect.Ptr {
		if v = v.Elem(); v.Kind() == reflect.Struct {
			return loadEnv(v, keywordParsers, typeParsers, kindParsers)
		} else {
			panic(fmt.Sprintf("'%s' provided but '%s' required", v.Kind(), reflect.Struct))
		}
	} else {
		panic(fmt.Sprintf("'%s' provided but '%s' required", v.Kind(), reflect.Ptr))
	}
}

func LoadEnv(itf interface{}) error {
	return LoadEnvUserParser(itf, nil, nil, nil)
}
