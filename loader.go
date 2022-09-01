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
	"os"
	"reflect"
	"strconv"
)

const tag = "env_var"

func setField(field reflect.Value, value string) (err error) {
	if field.IsValid() {
		var v interface{}
		switch field.Kind() {
		case reflect.Int64:
			v, err = strconv.ParseInt(value, 10, 64)
		case reflect.Float64:
			v, err = strconv.ParseFloat(value, 64)
		case reflect.Bool:
			v, err = strconv.ParseBool(value)
		case reflect.Slice, reflect.Map, reflect.Struct:
			x := reflect.New(field.Type())
			if err = json.Unmarshal([]byte(value), x.Interface()); err != nil {
				return
			}
			field.Set(x.Elem())
			return
		default:
			v = value
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

func LoadEnv(c interface{}) error {
	if v := reflect.ValueOf(c); v.Kind() == reflect.Ptr {
		if v = v.Elem(); v.Kind() == reflect.Struct {
			t := reflect.TypeOf(c).Elem()
			return loadEnv(t, v)
		} else {
			panic("must be struct")
		}
	} else {
		panic("must be pointer")
	}
}
