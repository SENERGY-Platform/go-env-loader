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
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
)

const tag = "env_var"

func setField(field reflect.Value, value string) (err error) {
	if field.IsValid() {
		var v interface{}
		var ui uint64
		var i int64
		var f float64
		var c complex128
		switch field.Kind() {
		case reflect.Uint:
			ui, err = strconv.ParseUint(value, 10, 0)
			v = uint(ui)
		case reflect.Uint8:
			ui, err = strconv.ParseUint(value, 10, 8)
			v = uint8(ui)
		case reflect.Uint16:
			ui, err = strconv.ParseUint(value, 10, 16)
			v = uint16(ui)
		case reflect.Uint32:
			ui, err = strconv.ParseUint(value, 10, 32)
			v = uint32(ui)
		case reflect.Uint64:
			v, err = strconv.ParseUint(value, 10, 64)
		case reflect.Int:
			i, err = strconv.ParseInt(value, 10, 0)
			v = int(i)
		case reflect.Int8:
			i, err = strconv.ParseInt(value, 10, 8)
			v = int8(i)
		case reflect.Int16:
			i, err = strconv.ParseInt(value, 10, 16)
			v = int16(i)
		case reflect.Int32:
			i, err = strconv.ParseInt(value, 10, 32)
			v = int32(i)
		case reflect.Int64:
			v, err = strconv.ParseInt(value, 10, 64)
		case reflect.Float32:
			f, err = strconv.ParseFloat(value, 32)
			v = float32(f)
		case reflect.Float64:
			v, err = strconv.ParseFloat(value, 64)
		case reflect.Complex64:
			c, err = strconv.ParseComplex(value, 64)
			v = complex64(c)
		case reflect.Complex128:
			v, err = strconv.ParseComplex(value, 128)
		case reflect.Bool:
			v, err = strconv.ParseBool(value)
		case reflect.Slice, reflect.Map, reflect.Struct:
			x := reflect.New(field.Type())
			if err = json.Unmarshal([]byte(value), x.Interface()); err != nil {
				return
			}
			field.Set(x.Elem())
			return
		case reflect.String:
			v = value
		default:
			err = errors.New(fmt.Sprintf("'%s' not supported", field.Kind()))
		}
		if err != nil {
			return
		}
		field.Set(reflect.ValueOf(v))
	}
	return
}

func getEnv(t reflect.Type, i int) (val string, ok bool) {
	if val, ok = t.Field(i).Tag.Lookup(tag); ok {
		val, ok = os.LookupEnv(val)
	}
	return
}

func loadEnv(t reflect.Type, v reflect.Value) (err error) {
	for i := 0; i < t.NumField(); i++ {
		if val, ok := getEnv(t, i); ok {
			if err = setField(v.Field(i), val); err != nil {
				return
			}
		} else if v.Field(i).Kind() == reflect.Struct {
			if err = loadEnv(t.Field(i).Type, v.Field(i)); err != nil {
				return
			}
		}
	}
	return
}

func LoadEnv(itf interface{}) error {
	if v := reflect.ValueOf(itf); v.Kind() == reflect.Ptr {
		if v = v.Elem(); v.Kind() == reflect.Struct {
			t := reflect.TypeOf(itf).Elem()
			return loadEnv(t, v)
		} else {
			panic(fmt.Sprintf("'%s' provided but '%s' required", v.Kind(), reflect.Struct))
		}
	} else {
		panic(fmt.Sprintf("'%s' provided but '%s' required", v.Kind(), reflect.Ptr))
	}
}
