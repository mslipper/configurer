[![configurer](https://circleci.com/gh/mslipper/configurer.svg?style=svg)](https://github.com/mslipper/configurer)


# configurer

A Golang library for configuration management.

## Usage

`configurer` can be integrated with your application in three steps.

**Step 1:** Define a `Config` struct in your application:

```go
type Config struct {
	DatabaseURL string `toml:"database_url" config:"required"`
	ListenPort  int    `toml:"listen_port"  config:"default=8080"`
}
```

The `config` struct tags in the example above are used to define validation and default values for each config field. They're completely optional, but useful.

**Step 2:** Create a config file in JSON, YAML TOML:

```toml
database_url = "postgres://localhost:5432/bigdb"
listen_port = 8080
```

**Step 3:** Load that config in your application:

```go
var cfg Config
if err := configurer.LoadURL("file:///my-config.toml", &cfg); err != nil {
	log.Fatalf("error loading config: %v", err)
}

// do stuff with your config
```

If the config file fails validation, `LoadURL` will return an error. Otherwise, it will unmarshal the config file into the `Config` struct while honoring any defaults you defined.

That's it! `configurer` also supports some advanced configuration options that extend the library to support additional config file formats and source URLs.

## Customizing Behavior

`configurer` uses a `config` struct tag to control how configuration files are unmarshalled.

### Default Values

A field's default value can be set by tagging it with `default=<value>`, where `<value>` is the value you want the field to have if it doesn't exist in the config. Slice, array, or struct fields cannot be tagged with `default`.

### Environment Overrides

A field's default value can also be pulled from an environment variable by tagging it with `env=<variable-name>`. Defaults defined by environment variables override those defined by the `default` tag above. Slice, array, or struct fields cannot be tagged with `env`.

### Required Fields

You can mark a field as required with the  `required` tag. `required` fields must be explicitly set in the config file or have a default value. If the field is a slice, array, or string, then its length must also be non-zero.

## Acknowledgements

`configurer` is the spiritual successor to [configor](https://github.com/jinzhu/configor), which appears to be unmaintained.