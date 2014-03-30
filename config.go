package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
)

const (
	DEFAULT_API_KEY     = "68d11739028e2d581b8c2fce470a972b"
	DEFAULT_API_ROOT    = "https://api.trello.com/1/"
	DEFAULT_CONFIG_FILE = "~/.clitrello"
)

type Config struct {
	ApiKey  string `json:"key"`
	ApiRoot string `json:"root"`
	File    string `json:"-"`
	Token   string `json:"token"`
}

func GetConfiguration() *Config {
	// Read a whole configuration from the command line. It will later override
	// whatever we have in our persisted file (if any).
	lineConfig := commandLineConfig()

	// Load the configuration file (whose path may have itself been overrided
	// through a command line argument) and override with command line.
	config := NewConfig()
	config.Load(lineConfig.File)

	// As far as I know, Golang doesn't have partial function application so we
	// rely on a closure.
	flag.Visit(func(flag *flag.Flag) {
		overrideConfigurationFields(config, flag)
	})

	return config
}

/**
 * Basic configuration manipulation.
 */

func NewConfig() *Config {
	return &Config{
		ApiKey:  DEFAULT_API_KEY,
		ApiRoot: DEFAULT_API_ROOT,
		File:    DEFAULT_CONFIG_FILE,
	}
}

func (config *Config) Load(file string) {
	// The absence of configuration file doesn't constitute and error: we just
	// leave the receiver unchanged.
	if f, err := os.Open(file); err == nil {
		jsonDecoder := json.NewDecoder(f)
		jsonDecoder.Decode(config)
	}
}

func (config *Config) Save(file string) {
	// Impossibility to save the configuration to file _is_ an error.
	if f, err := os.Create(file); err != nil {
		log.Fatal(err)
	} else {
		jsonEncoder := json.NewEncoder(f)
		jsonEncoder.Encode(&config)
	}
}

/**
 * Higher level functions to deal with the ability to override and file defined
 * configuration values with a command-line switch.
 */

func commandLineConfig() *Config {
	config := NewConfig()
	flag.StringVar(&config.ApiKey, "api_key", config.ApiKey, "Trello API key")
	flag.StringVar(&config.ApiRoot, "api_root", config.ApiRoot, "Trello API root URL")
	flag.StringVar(&config.File, "config", config.File, "Clitrello configuration file")
	flag.StringVar(&config.Token, "token", config.Token, "Trello user token")
	flag.Parse()
	return config
}

func overrideConfigurationFields(config *Config, flag *flag.Flag) {
	// We associate each overridable field of the Config structure with its
	// command-line equivalent.
	mapping := map[string]*string{
		"api_key":  &config.ApiKey,
		"api_root": &config.ApiRoot,
		"config":   &config.File,
		"token":    &config.Token,
	}

	if configField, ok := mapping[flag.Name]; ok {
		*configField = flag.Value.String()
	}
}
