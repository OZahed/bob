# Configs

configs is a light weight library that helps with reading environment variables it helps you with loading env values
directly into a struct

## How to Use?

let's say you have some environment variables being set, you can just call the ParseStruct function wih a prefix value
of WnvLoader Object

```go
envKeysPrefix := "APPLICATION_NAME" // or just empty string

cfg := Config{}
err := configs.EnvGetter(configs.DefaultKeyBuilder, configs.DefaultEnvGetter).ParseStruct(&cfg, envKeysPrefix )
if err != nil {
    panic(err)
}

// Do something with the config values

```
