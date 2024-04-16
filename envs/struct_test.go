package envs_test

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/OZahed/bob/envs"
)

type TestParsVal struct {
	Name string `env:"NAME"`
}

func (t *TestParsVal) ParseEnv(prefix string) error {
	key := fmt.Sprintf("%s_%s", prefix, "NAME")
	t.Name = os.Getenv(strings.ReplaceAll(key, ".", "_"))

	if t.Name == "" {
		t.Name = "From ParsEnv"
	}

	return nil
}

func TestMarshaler_LoadStruct_defaults(t *testing.T) {
	date, _ := time.Parse(time.DateOnly, "2024-04-16")
	type dest struct {
		Str struct {
			At      time.Time          `env:"AT,default=2024-04-16"`
			Map     map[string]float64 `env:"MAP_VALUE,default=name:1.2,key2:5.8"`
			Timeout time.Duration      `env:"TIMEOUT,default=1s"`
		} `env:"TIMES"`
		Name    string      `env:"NAME,default=Omid"`
		ParsVal TestParsVal `env:"testValue"`
		Strings []int       `env:"STRING,default=1;2;3;4;5"`
	}

	wantDefaults := dest{
		Name:    "Omid",
		Strings: []int{1, 2, 3, 4, 5},
		ParsVal: TestParsVal{
			Name: "From ParsEnv",
		},
	}

	wantDefaults.Str.At = date
	wantDefaults.Str.Map = map[string]float64{
		"name": 1.2,
		"key2": 5.8,
	}
	wantDefaults.Str.Timeout = time.Second

	t.Run("want defaults", func(t *testing.T) {
		destination := dest{}
		if err := envs.NewParser(nil, nil).ParseStruct(&destination, "TEST"); (err != nil) != false {
			t.Errorf("Marshaler.Marshal() error = %v, wantErr %v", err, nil)
		}

		if !reflect.DeepEqual(destination, wantDefaults) {
			t.Errorf("%v  wantErr %v", destination, wantDefaults)
		}
	})

}

func TestMarshaler_LoadStruct_osSetEnv(t *testing.T) {
	type Config struct {
		Date    time.Time      `env:"DATE"`
		TestMap map[int]string `env:"MAP,default=1:Hello world"`
		TestVal TestParsVal    `env:"PARSE_VAL"`
		Strings []string       `env:"STRINGS,default=item1;item2"`
		Ints    []int          `env:"INTS,default=1, 2, 3, 4"`
		Server  struct {
			Host    string        `env:"HOST,default=127.0.0.1"`
			Port    int           `env:"PORT,default=8080"`
			TimeOut time.Duration `env:"TIMEOUT,default=10s"`
			TLS     bool          `env:"TLS"`
		} `env:"SERVER"`
		BadKey int `env:"BAD_KEY,default=100"`
	}

	const (
		urlString = "https://test.com"
		port      = 3000
		intVal    = 123456
		appName   = "TEST"
		badKey    = "BAD_KEY"
		stringVal = "abcdefg"
		timeStr   = "2024-04-16 13:32:27"
		parseVal  = "parse val name"
	)

	strings := []string{"item1", "item2", "item3"}

	testEnvs := map[string]string{
		"APP_STRINGS":        "item1 item2 item3",
		"APP_INTS":           "1,2,3,4,5,6",
		"APP_DATE":           timeStr,
		"APP_PARSE_VAL_NAME": parseVal,
		"APP_MAP":            "1:hello world; 2:second val; 3:abcd",
		"APP_SERVER_PORT":    strconv.Itoa(port),
		"APP_SERVER_HOST":    "localhost",
		"APP_SERVER_TIMEOUT": "2s",
		"APP_SERVER_TLS":     "t",
	}

	for k, v := range testEnvs {
		_ = os.Setenv(k, v)
	}

	date, _ := time.Parse(time.DateTime, timeStr)

	want := Config{
		Strings: strings,
		Ints:    []int{1, 2, 3, 4, 5, 6},
		Date:    date,
		BadKey:  100,
		TestMap: map[int]string{
			1: "hello world",
			2: "second val",
			3: "abcd",
		},
		TestVal: TestParsVal{
			Name: parseVal,
		},
	}

	want.Server.Host = "localhost"
	want.Server.Port = port
	want.Server.TimeOut = 2 * time.Second
	want.Server.TLS = true

	t.Run("want defaults", func(t *testing.T) {
		cfg := Config{}

		if err := envs.NewParser(envs.DefaultKeyFunc, envs.DefaultEnvGetter).
			ParseStruct(&cfg, "APP"); (err != nil) != false {
			t.Errorf("Marshaler.Marshal() error = %v, wantErr %v", err, nil)
		}

		if !reflect.DeepEqual(cfg, want) {
			t.Errorf("got: %v  want: %v", cfg, want)
		}
	})
}

func TestMarshaler_ParseStruct_WithoutTags(t *testing.T) {
	type Config struct {
		Date     time.Time
		TestMap  map[int]string
		ParseVal TestParsVal
		Strings  []string
		Ints     []int
		BadKey   int
		Server   struct {
			Host    string
			Port    int
			Timeout time.Duration
			TLS     bool
		}
	}

	const (
		urlString = "https://test.com"
		port      = 3000
		intVal    = 123456
		appName   = "TEST"
		badKey    = "BAD_KEY"
		stringVal = "abcdefg"
		timeStr   = "2024-04-16 13:32:27"
		parseVal  = "parse val name"
	)

	strings := []string{"item1", "item2", "item3"}

	testEnvs := map[string]string{
		"APP_STRINGS":        "item1 item2 item3",
		"APP_INTS":           "1,2,3,4,5,6",
		"APP_DATE":           timeStr,
		"APP_PARSE_VAL_NAME": parseVal,
		"APP_TEST_MAP":       "1:hello world; 2:second val; 3:abcd",
		"APP_SERVER_PORT":    strconv.Itoa(port),
		"APP_SERVER_HOST":    "localhost",
		"APP_SERVER_TIMEOUT": "2s",
		"APP_SERVER_TLS":     "t",
	}

	for k, v := range testEnvs {
		_ = os.Setenv(k, v)
	}

	date, _ := time.Parse(time.DateTime, timeStr)

	want := Config{
		Strings: strings,
		Ints:    []int{1, 2, 3, 4, 5, 6},
		Date:    date,
		BadKey:  0,
		TestMap: map[int]string{
			1: "hello world",
			2: "second val",
			3: "abcd",
		},
		ParseVal: TestParsVal{
			Name: parseVal,
		},
	}

	want.Server.Host = "localhost"
	want.Server.Port = port
	want.Server.Timeout = 2 * time.Second
	want.Server.TLS = true

	t.Run("want defaults", func(t *testing.T) {
		cfg := Config{}

		if err := envs.NewParser(envs.DefaultKeyFunc, envs.DefaultEnvGetter).
			ParseStruct(&cfg, "APP"); (err != nil) != false {
			t.Errorf("Marshaler.Marshal() error = %v, wantErr %v", err, nil)
		}

		if !reflect.DeepEqual(cfg, want) {
			t.Errorf("got: %v  want: %v", cfg, want)
		}
	})
}
