# Configs

configs is a light weight library that helps with reading environment variables it helps you with loading env values
directly into a struct or Just simply yu can use generic Get Function or Get For specific dataTypes

## How to Use?

let's say you have some environment variables being set, you can just call the ParseStruct function wih a prefix value
of WnvLoader Object

```go

type TestParsVal struct {
	Name string `env:"NAME"`
}


// ParseEnv is optional if a struct has ParseEnv function
// configs library would just call the ParseEnv otherwise it
// will run through struct fields one by one and checks for struct tags like:
//
//   type Config struct{
//      Date    time.Time      `env:"DATE,default=2024-04-16 13:32:27"`
//   }
func (t *TestParsVal) ParseEnv(prefix string) error {
	key := fmt.Sprintf("%s_%s", prefix, "NAME")
	t.Name = os.Getenv(strings.ReplaceAll(key, ".", "_"))

	if t.Name == "" {
		t.Name = "From ParsEnv"
	}

	return nil
}


type Config struct {
	Date    time.Time      `env:"DATE"`
    TestMap map[int]string `env:"MAP,default=1:Hello world"`
    TestVal TestParsVal    `env:"PARSE_VAL"`
    Strings []string       `env:"STRINGS,default=item1;item2"`
    Ints    []int          `env:"INTS,default=1, 2, 3, 4"`
   // even Anonymous could be used
   Server  struct {
    	Host    string        `env:"HOST,default=127.0.0.1"`
    	Port    int           `env:"PORT,default=8080"`
    	TimeOut time.Duration `env:"TIMEOUT,default=10s"`
    	TLS     bool          `env:"TLS"`
    } `env:"SERVER"`
    BadKey int `env:"BAD_KEY,default=100"`
}

envKeysPrefix := "APPLICATION_NAME" // or just empty string

cfg := Config{}
err := configs.EnvGetter(configs.DefaultKeyBuilder, configs.DefaultEnvGetter).ParseStruct(&cfg, envKeysPrefix )
if err != nil {
    panic(err)
}

// Do something with the config values

```

or if you just need to get values with default:

```go
// will return a value of default(second input) type
timeout := configs.GetDefault("exactEnvKey", time.Second * 10)

```

or without default value

```go

// will return a value of zero value of type time.Duration
timeout := configs.Get[time.Duration]("exactEnvKey");
```

or you may prefer explicit function names:

```go
str := configs.GetString("exactEnvKey", "default value") string
```
