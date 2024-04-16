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

## Basic Usage with`EnvParser` implementation Example

```go
type TestParsVal struct {
	Name string `env:"NAME"`
}

// ParseEnv implementation
func (t *TestParsVal) ParseEnv(prefix string) error {
	key := fmt.Sprintf("%s.%s", prefix, "NAME")
	// you can use your own EnvGetter and KeyFunc implementation
	t.Name = envs.DefaultEnvGetter(envs.DefaultKeyFunc(key, ".", "_"))

	if t.Name == "" {
		t.Name = "From ParsEnv"
	}

	return nil
}

type Config struct {
	Date     time.Time
	IntMaps  map[int]string
	FloarMap map[string]float64
	ParseVal TestParsVal
	Strings  []string
	Ints     []int
	private  int 	// this field won't be changed because it is not exported.
	Server   struct {
		Host    string
		Port    int
		Timeout time.Duration
		TLS     bool
	}
}

func main() {
	cfg := Config{}
	if err := envs.NewParser(envs.DefaultKeyFunc, envs.DefaultEnvGetter).ParseStruct(&cfg, "APP"); err != nil {
		log.Fatal(err)
	}

	// cfg is loaded and can be used
}

```

---

envs package also provides a Generic `Get` and `GetDefault` function which work like this:

```go
// will return a value of default(second input) type
timeout := envs.GetDefault("exactEnvKey", time.Second * 10)

// will return a value of zero value of type time.Duration
timeout := envs.Get[time.Duration]("exactEnvKey");

str := envs.GetDuration("exactEnvKey", time.Second*5)
```

---

## to find out how to use the env parser check `struct_test.go` out
