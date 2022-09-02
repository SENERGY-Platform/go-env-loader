go-env-loader
=======

Load struct field values from environment variables defined via tags. Supports basic types, slices, maps and structs!

Quickstart
---

Declare environment variables:

```shell
export APP_ID='0c19d322-bc6f-43ea-8956-a853f4db9c06'
export DB_HOST='localhost'
export DB_PORT='5034'
export LOG_LEVEL='debug'
export KEY_MAP='{"success": 0, "error": 1}'
export INCLUDE='["/var/app", "/opt/mnt"]'
```

Declare struct types with `env_var` tags:

```go
package main

import (
	"fmt"
	"github.com/y-du/go-env-loader"
)

type DatabaseConfig struct {
	Host string `env_var:"DB_HOST"`
	Port int64  `env_var:"DB_PORT"`
}

type Config struct {
	AppId      string           `env_var:"APP_ID"`
	RetryDelay int64            `env_var:"RETRY_DELAY"`
	AllowRetry bool             `env_var:"ALLOW_RETRY"`
	LogLevel   string           `env_var:"LOG_LEVEL"`
	Database   DatabaseConfig   `env_var:"DB_CONFIG"`
	KeyMap     map[string]int64 `env_var:"KEY_MAP"`
	Include    []string         `env_var:"INCLUDE"`
}

func main() {
	// declare default values
	config := Config{
		RetryDelay: 5,
		AllowRetry: true,
		LogLevel:   "info",
	}

	// load values from environment
	if err := envldr.LoadEnv(&config); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("%#v\n", config)
	// prints: main.Config{AppId:"0c19d322-bc6f-43ea-8956-a853f4db9c06", RetryDelay:5, AllowRetry:true, LogLevel:"debug", Database:main.DatabaseConfig{Host:"localhost", Port:5034}, KeyMap:map[string]int64{"error":1, "success":0}, Include:[]string{"/var/app", "/opt/mnt"}}

}
```

Analogous to slices and maps, struct types can also be loaded from environment:

```shell
export DB_CONFIG='{"Host": "somedb", "Port": 4021}'
```

The tags `DB_HOST` and `DB_PORT` set in `DatabaseConfig` are now ignored:

```go
        fmt.Printf("%#v\n", config)
        // prints: main.Config{AppId:"0c19d322-bc6f-43ea-8956-a853f4db9c06", RetryDelay:5, AllowRetry:true, LogLevel:"debug", Database:main.DatabaseConfig{Host:"somedb", Port:4021}, KeyMap:map[string]int64{"error":1, "success":0}, Include:[]string{"/var/app", "/opt/mnt"}}
```
