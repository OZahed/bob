package envs

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	separators  = []string{" ", ",", ";", "\n"}
	timeLayouts = []string{time.RFC3339, time.RFC3339Nano, time.DateTime, time.Stamp, time.DateOnly,
		time.TimeOnly, time.ANSIC, time.RFC822, time.UnixDate,
	}
)

// Methods can not be generic so I have to wrap everything
type Getter struct {
	key func(name string) string
}

func (a *Getter) GetString(name, def string) string {
	return GetDefault(a.key(name), def)
}

func (a *Getter) GetStringSlice(name string) []string {
	return GetDefault(a.key(name), []string{})
}

func (a *Getter) GetInt(name string, def int) int {
	return GetDefault(a.key(name), def)
}

func (a *Getter) GetInt64(name string, def int64) int64 {
	return GetDefault(a.key(name), def)
}

func (a *Getter) GetInt32(name string, def int32) int32 {
	return GetDefault(a.key(name), def)
}

func (a *Getter) GetFloat64(name string, def float64) float64 {
	return GetDefault(a.key(name), def)
}

func (a *Getter) GetFloat32(name string, def float32) float32 {
	return GetDefault(a.key(name), def)
}

func (a *Getter) GetBool(name string) bool {
	return GetDefault(a.key(name), false)
}

func (a *Getter) GetTime(name string) time.Time {
	return GetDefault(a.key(name), time.Time{})
}

func (a *Getter) GetDuration(name string, def time.Duration) time.Duration {
	return GetDefault(a.key(name), def)
}

func GetDefault[T any](name string, def T) T {
	val := Get[T](name)

	if reflect.ValueOf(val).IsZero() {
		return def
	}

	return val
}

func Get[T any](name string) T {
	tp := reflect.TypeFor[T]()
	var res any

	val := os.Getenv(name)
	if val == "" {
		return reflect.New(tp).Elem().Interface().(T)
	}

	switch tp.Kind() {
	case reflect.String:
		res = val
	case reflect.Slice:
		if tp.Elem().Kind() == reflect.String {
			for _, sep := range separators {
				split := strings.Split(val, sep)
				if split[0] != val {
					res = split
					break
				}
			}
		}

		if tp.Elem().Kind() == reflect.Int {
			split := strings.Split(val, ",")
			arr := make([]int, 0)
			for _, str := range split {
				arr = append(arr, int(parseInt64(str)))
			}
			res = arr
		}
	case reflect.Int:
		res = int(parseInt64(val))
	case reflect.Int32:
		res = int32(parseInt64(val))
	case reflect.Int64:
		res = parseInt64(val)
	case reflect.Float64:
		res, _ = strconv.ParseFloat(val, 64)
	case reflect.Float32:
		res, _ = strconv.ParseFloat(val, 32)
	case reflect.Bool:
		res, _ = strconv.ParseBool(val)
	}

	if tp == reflect.TypeOf(time.Duration(0)) {
		res, _ = time.ParseDuration(val)
	}

	if tp == reflect.TypeOf(time.Time{}) {
		for _, layout := range timeLayouts {
			t, err := time.Parse(layout, val)
			if err == nil && !t.IsZero() {
				res = t
				break
			}
		}
	}

	if res == nil {
		fmt.Println("nil")
	}

	if reflect.TypeOf(res) != reflect.TypeFor[T]() {
		return reflect.New(tp).Elem().Interface().(T)
	}

	return reflect.ValueOf(res).Interface().(T)
}

func parseInt64(val string) int64 {
	n, _ := strconv.ParseInt(val, 10, 64)
	return n
}

func MakeKeyProviderPrefix(prefix string) func(name string) string {
	return func(name string) string {
		if prefix == "" {
			return name
		}

		return prefix + "_" + name
	}
}
