package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	errUnknownType = errors.New("unknown type")
	emptyField     reflect.StructField
	// ErrConvertMapStringSlice can not convert to map[string][]string
	ErrConvertMapStringSlice = errors.New("can not convert to map slices of strings")
	// ErrConvertToMapString can not convert to map[string]string
	ErrConvertToMapString = errors.New("can not convert to map of strings")
)

type DefFunc func(reflect.StructField) string

func (df DefFunc) hasDefault(field reflect.StructField) bool {
	// Use df as a function
	tagStr := df(field)
	return !(tagStr == "" || tagStr == "-")
}

func (df DefFunc) ignore(field reflect.StructField) bool {
	// Use df as a function
	tagStr := df(field)
	return tagStr == "-"
}

func GetTagDefFunc(tag string) DefFunc {
	return func(sf reflect.StructField) string {
		return sf.Tag.Get(tag)
	}
}

func GetSubTagDefFunc(tag, defKey string) DefFunc {
	return func(sf reflect.StructField) string {
		tagValue := sf.Tag.Get(tag)
		tagValue, opts := head(tagValue, ",")

		if tagValue == "" { // when field is "emptyField" variable
			return ""
		}

		var opt string
		for len(opts) > 0 {
			opt, opts = head(opts, ",")

			if k, v := head(opt, "="); k == defKey {
				return v
			}
		}
		return ""
	}
}

// Map `default` key in `form` tag to the Struct elements
//
//	Method	string	`form:"Method,default=POST"`
func MapFormDefault(ptr any) error {
	return mapFormByTag(ptr, GetSubTagDefFunc("form", "default"))
}

// Map `default` tag value to the Struct elements
//
//	Method	string	`form:"Method" default:"POST"`
func MapDefault(ptr any) error {
	return mapFormByTag(ptr, GetTagDefFunc("default"))
}

func mapFormByTag(ptr any, defFn DefFunc) error {
	if checkNonNilPointer(ptr) && findCoreType(reflect.TypeOf(ptr)) == reflect.Struct {
		return mappingByPtr(ptr, defFn)
	}
	return errUnknownType
}

func checkNonNilPointer(ptr any) bool {
	ptrType := reflect.TypeOf(ptr)
	if ptrType.Kind() == reflect.Ptr && !reflect.ValueOf(ptr).IsNil() {
		return true
	}
	return false
}

func findCoreType(_type reflect.Type) reflect.Kind {
	if _type.Kind() == reflect.Ptr {
		return findCoreType(_type.Elem())
	}
	return _type.Kind()
}

func mappingByPtr(ptr any, defFn DefFunc) error {
	_, err := mapping(reflect.ValueOf(ptr), emptyField, defFn)
	return err
}

func mapping(value reflect.Value, field reflect.StructField, defFn DefFunc) (bool, error) {
	if defFn.ignore(field) { // just ignoring this field
		return false, nil
	}

	vKind := value.Kind()

	if vKind == reflect.Ptr {
		var isNew bool
		vPtr := value
		if value.IsNil() {
			isNew = true
			vPtr = reflect.New(value.Type().Elem())
		}
		isSet, err := mapping(vPtr.Elem(), field, defFn)
		if err != nil {
			return false, err
		}
		if isNew && isSet {
			value.Set(vPtr)
		}
		return isSet, nil
	}

	if vKind != reflect.Struct || !field.Anonymous {
		ok, err := tryToSetValue(value, field, defFn)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}

	if vKind == reflect.Struct {
		tValue := value.Type()

		var isSet bool
		for i := 0; i < value.NumField(); i++ {
			sf := tValue.Field(i)
			if sf.PkgPath != "" && !sf.Anonymous { // unexported
				continue
			}
			ok, err := mapping(value.Field(i), sf, defFn)
			if err != nil {
				return false, err
			}
			isSet = isSet || ok
		}
		return isSet, nil
	}
	return false, nil
}

type setOptions struct {
	isDefaultExists bool
	defaultValue    string
}

func tryToSetValue(value reflect.Value, field reflect.StructField, defFn DefFunc) (bool, error) {
	var setOpt setOptions

	if defFn.hasDefault(field) {
		setOpt.isDefaultExists = true
		setOpt.defaultValue = defFn(field)
	}

	return setDefault(value, field, setOpt)
}

func setDefault(value reflect.Value, field reflect.StructField, opt setOptions) (isSet bool, err error) {
	if !opt.isDefaultExists {
		return false, nil
	}

	switch value.Kind() {
	case reflect.Slice:
		return true, setSlice([]string{opt.defaultValue}, value, field)
	case reflect.Array:
		vs := []string{opt.defaultValue}
		if len(vs) != value.Len() {
			return false, fmt.Errorf("%q is not valid value for %s", vs, value.Type().String())
		}
		return true, setArray(vs, value, field)
	default:
		return true, setWithProperType(opt.defaultValue, value, field)
	}
}

func setWithProperType(val string, value reflect.Value, field reflect.StructField) error {
	switch value.Kind() {
	case reflect.Int:
		return setIntField(val, 0, value)
	case reflect.Int8:
		return setIntField(val, 8, value)
	case reflect.Int16:
		return setIntField(val, 16, value)
	case reflect.Int32:
		return setIntField(val, 32, value)
	case reflect.Int64:
		switch value.Interface().(type) {
		case time.Duration:
			return setTimeDuration(val, value)
		}
		return setIntField(val, 64, value)
	case reflect.Uint:
		return setUintField(val, 0, value)
	case reflect.Uint8:
		return setUintField(val, 8, value)
	case reflect.Uint16:
		return setUintField(val, 16, value)
	case reflect.Uint32:
		return setUintField(val, 32, value)
	case reflect.Uint64:
		return setUintField(val, 64, value)
	case reflect.Bool:
		return setBoolField(val, value)
	case reflect.Float32:
		return setFloatField(val, 32, value)
	case reflect.Float64:
		return setFloatField(val, 64, value)
	case reflect.String:
		value.SetString(val)
	case reflect.Struct:
		switch value.Interface().(type) {
		case time.Time:
			return setTimeField(val, field, value)
		}
		return json.Unmarshal([]byte(val), value.Addr().Interface())
	case reflect.Map:
		return json.Unmarshal([]byte(val), value.Addr().Interface())
	default:
		return errUnknownType
	}
	return nil
}

func setIntField(val string, bitSize int, field reflect.Value) error {
	if val == "" {
		val = "0"
	}
	intVal, err := strconv.ParseInt(val, 10, bitSize)
	if err == nil {
		field.SetInt(intVal)
	}
	return err
}

func setUintField(val string, bitSize int, field reflect.Value) error {
	if val == "" {
		val = "0"
	}
	uintVal, err := strconv.ParseUint(val, 10, bitSize)
	if err == nil {
		field.SetUint(uintVal)
	}
	return err
}

func setBoolField(val string, field reflect.Value) error {
	if val == "" {
		val = "false"
	}
	boolVal, err := strconv.ParseBool(val)
	if err == nil {
		field.SetBool(boolVal)
	}
	return err
}

func setFloatField(val string, bitSize int, field reflect.Value) error {
	if val == "" {
		val = "0.0"
	}
	floatVal, err := strconv.ParseFloat(val, bitSize)
	if err == nil {
		field.SetFloat(floatVal)
	}
	return err
}

func setTimeField(val string, structField reflect.StructField, value reflect.Value) error {
	timeFormat := structField.Tag.Get("time_format")
	if timeFormat == "" {
		timeFormat = time.RFC3339
	}

	switch tf := strings.ToLower(timeFormat); tf {
	case "unix", "unixnano":
		tv, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}

		d := time.Duration(1)
		if tf == "unixnano" {
			d = time.Second
		}

		t := time.Unix(tv/int64(d), tv%int64(d))
		value.Set(reflect.ValueOf(t))
		return nil
	}

	if val == "" {
		value.Set(reflect.ValueOf(time.Time{}))
		return nil
	}

	l := time.Local
	if isUTC, _ := strconv.ParseBool(structField.Tag.Get("time_utc")); isUTC {
		l = time.UTC
	}

	if locTag := structField.Tag.Get("time_location"); locTag != "" {
		loc, err := time.LoadLocation(locTag)
		if err != nil {
			return err
		}
		l = loc
	}

	t, err := time.ParseInLocation(timeFormat, val, l)
	if err != nil {
		return err
	}

	value.Set(reflect.ValueOf(t))
	return nil
}

func setArray(vals []string, value reflect.Value, field reflect.StructField) error {
	for i, s := range vals {
		err := setWithProperType(s, value.Index(i), field)
		if err != nil {
			return err
		}
	}
	return nil
}

func setSlice(vals []string, value reflect.Value, field reflect.StructField) error {
	slice := reflect.MakeSlice(value.Type(), len(vals), len(vals))
	err := setArray(vals, slice, field)
	if err != nil {
		return err
	}
	value.Set(slice)
	return nil
}

func setTimeDuration(val string, value reflect.Value) error {
	d, err := time.ParseDuration(val)
	if err != nil {
		return err
	}
	value.Set(reflect.ValueOf(d))
	return nil
}

func head(str, sep string) (head string, tail string) {
	idx := strings.Index(str, sep)
	if idx < 0 {
		return str, ""
	}
	return str[:idx], str[idx+len(sep):]
}
