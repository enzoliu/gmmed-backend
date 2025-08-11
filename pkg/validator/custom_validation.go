package validator

import (
	"cmp"
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-playground/validator/v10"
	"github.com/guregu/null/v5"
)

var (
	timeDurationType = reflect.TypeOf(time.Duration(0))
)

func CustomUUIDValidation() *ValidationRegister {
	return &ValidationRegister{
		Tag:  "custom_uuid",
		Func: NewRegexValidationFunc(customUUIDRegex),
	}
}

// drop-in replacement for build-in `uuid`
func NullableUUIDValidation() *ValidationRegister {
	return &ValidationRegister{
		Tag:  "uuid",
		Func: NewRegexValidationFunc(uuidRegex),
	}
}

// drop-in replacement for build-in `max`
func NullableMaxValidation() *ValidationRegister {
	compareToBool := func(i int) bool {
		return i <= 0
	}

	return &ValidationRegister{
		Tag:  "max",
		Func: newOrderedValidationFunc(compareToBool),
	}
}

// drop-in replacement for build-in `min`
func NullableMinValidation() *ValidationRegister {
	compareToBool := func(i int) bool {
		return i >= 0
	}

	return &ValidationRegister{
		Tag:  "min",
		Func: newOrderedValidationFunc(compareToBool),
	}
}

// drop-in replacement for build-in `lte`
func NullableLteValidation() *ValidationRegister {
	compareToBool := func(i int) bool {
		return i <= 0
	}

	return &ValidationRegister{
		Tag:  "lte",
		Func: newOrderedValidationFunc(compareToBool),
	}
}

// drop-in replacement for build-in `gte`
func NullableGteValidation() *ValidationRegister {
	compareToBool := func(i int) bool {
		return i >= 0
	}

	return &ValidationRegister{
		Tag:  "gte",
		Func: newOrderedValidationFunc(compareToBool),
	}
}

// drop-in replacement for build-in `http_url`
func NullableHttpUrlValidation() *ValidationRegister {
	checkUrl := func(input string) bool {
		input = strings.ToLower(input)

		url, err := url.Parse(input)
		if err != nil || url.Host == "" {
			return false
		}

		return url.Scheme == "http" || url.Scheme == "https"
	}

	f := func(fl validator.FieldLevel) bool {
		switch val := fl.Field().Interface().(type) {
		case string:
			return checkUrl(val)

		case null.String:
			// assume omitempty option is used
			return !val.Valid || val.String == "" || checkUrl(val.String)

		default:
			panic(fmt.Sprintf("unsupported type '%T'", val))
		}
	}

	return &ValidationRegister{
		Tag:  "http_url",
		Func: f,
	}
}

func NewRegexValidationFunc(regex *regexp.Regexp) func(validator.FieldLevel) bool {
	return func(fl validator.FieldLevel) bool {
		switch val := fl.Field().Interface().(type) {
		case string:
			return regex.MatchString(val)

		case null.String:
			// assume omitempty option is used
			return !val.Valid || val.String == "" || regex.MatchString(val.String)

		default:
			panic(fmt.Sprintf("unsupported type '%T'", val))
		}
	}
}

func NewEnumValidationFunc[T fmt.Stringer](values ...T) func(validator.FieldLevel) bool {
	enumMap := make(map[string]struct{}, len(values))
	for _, v := range values {
		enumMap[v.String()] = struct{}{}
	}

	return func(fl validator.FieldLevel) bool {
		switch val := fl.Field().Interface().(type) {
		case T:
			_, ok := enumMap[val.String()]
			return ok

		case string:
			_, ok := enumMap[val]
			return ok

		case null.Value[T]:
			// assume omitempty option is used
			if !val.Valid {
				return true
			}
			_, ok := enumMap[val.V.String()]
			return ok

		case null.String:
			// assume omitempty option is used
			if !val.Valid {
				return true
			}
			_, ok := enumMap[val.String]
			return ok

		default:
			panic(fmt.Sprintf("unsupported type '%T'", val))
		}
	}
}

// Copy from https://github.com/go-playground/validator/blob/v10.20.0/baked_in.go#L2333
// With some modifications to support null types
func newOrderedValidationFunc(compareToBool func(int) bool) func(validator.FieldLevel) bool {
	return func(fl validator.FieldLevel) bool {
		field := fl.Field()
		param := fl.Param()

		switch field.Kind() {
		case reflect.String:
			p := asInt(param)
			compare := cmp.Compare(int64(utf8.RuneCountInString(field.String())), p)
			return compareToBool(compare)

		case reflect.Slice, reflect.Map, reflect.Array:
			p := asInt(param)
			compare := cmp.Compare(int64(field.Len()), p)
			return compareToBool(compare)

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			p := asIntFromType(field.Type(), param)
			compare := cmp.Compare(field.Int(), p)
			return compareToBool(compare)

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			p := asUint(param)
			compare := cmp.Compare(field.Uint(), p)
			return compareToBool(compare)

		case reflect.Float32:
			p := asFloat32(param)
			compare := cmp.Compare(float64(field.Float()), p)
			return compareToBool(compare)

		case reflect.Float64:
			p := asFloat64(param)
			compare := cmp.Compare(field.Float(), p)
			return compareToBool(compare)

		case reflect.Struct:
			switch val := field.Interface().(type) {
			case time.Time:
				now := time.Now().Unix()
				compare := cmp.Compare(val.Unix(), now)
				return compareToBool(compare)

			case null.String:
				// assume omitempty option is used
				if !val.Valid || val.String == "" {
					return true
				}
				p := asInt(param)
				compare := cmp.Compare(int64(utf8.RuneCountInString(field.String())), p)
				return compareToBool(compare)

			case null.Int:
				// assume omitempty option is used
				if !val.Valid || val.Int64 == 0 {
					return true
				}
				p := asInt(param)
				compare := cmp.Compare(val.Int64, p)
				return compareToBool(compare)
			}
		}

		panic(fmt.Sprintf("unsupported type '%T'", field.Interface()))
	}
}

func asInt(param string) int64 {
	i, err := strconv.ParseInt(param, 0, 64)
	panicIf(err)
	return i
}

func asIntFromTimeDuration(param string) int64 {
	d, err := time.ParseDuration(param)
	if err != nil {
		// attempt parsing as an integer assuming nanosecond precision
		return asInt(param)
	}
	return int64(d)
}

func asIntFromType(t reflect.Type, param string) int64 {
	switch t {
	case timeDurationType:
		return asIntFromTimeDuration(param)
	default:
		return asInt(param)
	}
}

func asUint(param string) uint64 {
	i, err := strconv.ParseUint(param, 0, 64)
	panicIf(err)
	return i
}

func asFloat64(param string) float64 {
	i, err := strconv.ParseFloat(param, 64)
	panicIf(err)
	return i
}

func asFloat32(param string) float64 {
	i, err := strconv.ParseFloat(param, 32)
	panicIf(err)
	return i
}

func panicIf(err error) {
	if err != nil {
		panic(err.Error())
	}
}
