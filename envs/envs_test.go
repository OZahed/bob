package envs_test

import (
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/OZahed/bob/envs"
)

// Since the function is a generic and depends on the input datatype common table tests do not work as intended
func TestGetEnv(t *testing.T) {
	const (
		unixDate  = "Sat Apr 13 17:42:36 +0330 2024"
		rfcDate   = "2024-04-13 14:12:56+00:00"
		port      = 3000
		intVal    = 123456
		appName   = "TEST"
		badKey    = "BAD_KEY"
		stringVal = "abcdefg"
		timeStr   = "2024-01-01"
	)

	testEnvs := map[string]string{
		"TEST_PORT":       strconv.Itoa(port),
		"TEST_DURATION":   "2s",
		"TEST_STRING_VAL": stringVal,
		"TEST_STRINGS1":   "item1,item2,item3",
		"TEST_DATE":       "2024-01-01",
		"TEST_BOOL":       "t",
	}

	for k, v := range testEnvs {
		_ = os.Setenv(k, v)
	}
	date, _ := time.Parse(time.DateOnly, timeStr)
	strings := []string{"item1", "item2", "item3"}
	keyProvider := envs.MakeKeyProviderPrefix(appName)

	t.Parallel()
	t.Run("Test Generic For Port", func(t *testing.T) {
		if got := envs.Get[int](keyProvider("PORT")); !reflect.DeepEqual(got, 3000) {
			t.Errorf("GetEnv() = %v, want %v", got, 3000)
		}
	})

	t.Run("Test Generic Default", func(t *testing.T) {
		if got := envs.GetDefault(keyProvider("PORT_BAD_KEY"), 8080); !reflect.DeepEqual(got, 8080) {
			t.Errorf("GetEnv() = %v, want %v", got, 8080)
		}
	})

	t.Run("Test Generic for duration", func(t *testing.T) {
		if got := envs.Get[time.Duration](keyProvider("DURATION")); !reflect.DeepEqual(got, time.Second*2) {
			t.Errorf("GetEnv() = %v, want %v", got, time.Second*2)
		}
	})

	t.Run("Test Generic With Default Value", func(t *testing.T) {
		if got := envs.GetDefault(keyProvider(badKey), time.Hour); !reflect.DeepEqual(got, time.Hour) {
			t.Errorf("GetEnv() = %v, want %v", got, time.Hour)
		}
	})

	t.Run("Test Generic for string", func(t *testing.T) {
		if got := envs.Get[string](keyProvider("STRING_VAL")); !reflect.DeepEqual(got, stringVal) {
			t.Errorf("GetEnv() = %v, want %v", got, stringVal)
		}
	})

	t.Run("Test Generic for string array", func(t *testing.T) {
		if got := envs.Get[[]string](keyProvider("STRINGS1")); !reflect.DeepEqual(got, strings) {
			t.Errorf("GetEnv() = %v, want %v", got, strings)
		}
	})

	t.Run("Test Generic for date", func(t *testing.T) {
		if got := envs.Get[time.Time](keyProvider("DATE")); !reflect.DeepEqual(got, date) {
			t.Errorf("GetEnv() = %v, want %v", got, date)
		}
	})

	t.Run("Test Generic date with default ", func(t *testing.T) {
		now := time.Now()
		if got := envs.GetDefault(keyProvider("DATE_adfasdf"), now); !reflect.DeepEqual(got, now) {
			t.Errorf("GetEnv() = %v, want %v", got, now)
		}
	})

	t.Run("Test Generic for bool", func(t *testing.T) {
		if got := envs.Get[bool](keyProvider("BOOL")); !reflect.DeepEqual(got, true) {
			t.Errorf("GetEnv() = %v, want %v", got, true)
		}
	})

	t.Run("Test Generic for bad key", func(t *testing.T) {
		if got := envs.Get[bool](keyProvider(badKey)); !reflect.DeepEqual(got, false) {
			t.Errorf("GetEnv() = %v, want %v", got, false)
		}

		if got := envs.Get[int](keyProvider(badKey)); !reflect.DeepEqual(got, 0) {
			t.Errorf("GetEnv() = %v, want %v", got, 0)
		}

		if got := envs.Get[time.Time](keyProvider(badKey)); !reflect.DeepEqual(got, time.Time{}) {
			t.Errorf("GetEnv() = %v, want %v", got, time.Time{})
		}
	})

	t.Run("Test Generic for wring value", func(t *testing.T) {
		const key = "test"

		_ = os.Setenv(key, "hello world")
		if got := envs.Get[bool](key); !reflect.DeepEqual(got, false) {
			t.Errorf("GetEnv() = %v, want %v", got, false)
		}

		_ = os.Setenv(key, "1.234")
		if got := envs.Get[int32](keyProvider(key)); !reflect.DeepEqual(got, int32(0)) {
			t.Errorf("GetEnv() = %v, want %v", got, 0)
		}

		_ = os.Setenv(key, "2024-04-38 25:12:28+03:30") // wrong time
		if got := envs.Get[time.Time](keyProvider(key)); !reflect.DeepEqual(got, time.Time{}) {
			t.Errorf("GetEnv() = %v, want %v", got, time.Time{})
		}
	})
}
