# Envs

configs is a light weight library that helps with reading environment variables it helps with loading env values
directly into a struct or Just simply provides some generic GetEnv interface

## How to Use?

the package has a very simple use case, struct fields can have a `env:"ENVNAME,default=default value"` or `env:"ENVNAME,default value"`

> if struct fields did not have an `env` struct tag, the field name as UPPERCASE_SNAKE_CASE would be considered as the `env:name`

## How it works

### Supported data types

- all `int`s and `uint`s
- all `float` types
- `time.Duration` and `time.Time`
- `string`
- all kinds of arrays ( preferably do not uses interface as array type )
- all kings of maps (preferably do not uses interface as key or value types )
- `anonymous struct`
- `struct`s
- `*url.Url`

inner struct keys will be concatenated with their parent keys for example in below scenario

```go
type Config struct{
	Server struct {
		Port int
	}
}
```

to parse Port value, the Parser will check for `SERVER.PORT` value

> NOTE: `DefaultKeyFunc` function will replace `.` chars with `_` but parser will send `SERVER.PORT` as key it is
> `KeyFunc`'s responsibility to change `PARENT.CHILD` string into required string for `GetFunc`

> NOTE: `GetFun` is responsible for reading ENVs or whatever other arbitrary config source

> NOTE: if a struct pointer did implement `EnvParser` parser would only call the interface and ignores the default process

\*\* envs package also provides a Generic `Get` and `GetDefault` function

## Basic Usage with`EnvParser` implementation Example

```go
type TestParsVal struct {
	When 	 time.Time
	Name 	 string
	Duration time.Duration
}

// EnvParser interface implementation
// Optional, if a struct implements EnvParser, envs pacakge
// will only call the ParseEnv for that struct and won't use reflection
// otherwise, it would recursively call reflection on struct fields
func (t *TestParsVal) ParseEnv(prefix string) error {
	whenKey := prefix + ".HAPPENED_AT"
	nameKey := prefix + ".SPECIAL_KEY"
	timeoutKey := prefix + ".TIMEOUT"

	when := envs.DefaultGetFunc(whenKey, "")
	t,err := time.Parse("your time format", when)
	if err != nil {
		return err
	}

	t.Time = t


	t.Name = envs.DefaultGetFunc(nameKey, "test name value")

	timeout := envs.DefaultGetFunc(timeoutKey, "10s")
	d,err := time.ParseDuration(timeout)
	if err != nil {
		return err
	}

	t.Duration = d

	return nil
}

type Config struct {
	// env Tags are not mandatory but required for default values
	Date     time.Time `env:"DATE,default=2024-03-20"`
	// you can change the default names just like json tags
	IntMaps  map[int]string `env:"MAP1,default=200:OK, 304:redirect, 404:not found"` // Will search for prefix.MAP1 or prefix_MAP1
	FloatMap map[string]float64 // will search for prefix.FLOAT_MAP or prefix_FLOAT_MAP
	ParseVal TestParsVal
	Strings  []string
	IntArr   []int
  // private fields will get ignored
	privateField int
	Server       struct {
		Host         string
		Port         int
		Timeout      time.Duration
		TLS          bool
	}
}

func main() {
	cfg := Config{}
	if err := envs.NewParser(envs.DefaultKeyFunc, envs.DefaultEnvGetter).ParseStruct(&cfg, "APP"); err != nil {
		log.Fatal(err)
	}

	// cfg is loaded and can be used

	// if you needed a value that is not inside your config struct
	anotherDuration :=  envs.Get[time.Duration]("APP_DURATION_KEY")
	durationWithDefault := envs.GetDefault("APP_DURATION_KEY2", time.Second * 2)
}

```

---

## to find out how to use the env parser check `struct_test.go` out
