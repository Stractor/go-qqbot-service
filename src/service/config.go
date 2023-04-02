package service

import (
	"fmt"
	"os"
	"sync"

	"github.com/mitchellh/mapstructure"
	"go.uber.org/dig"
	"gopkg.in/yaml.v2"
)

var (
	once        sync.Once
	configData  map[string]interface{}
	configError error
)

func readConfig() {
	data, err := os.ReadFile("config.yml")
	if err != nil {
		configError = fmt.Errorf("unable to read config file: %v", err)
		return
	}

	err = yaml.Unmarshal(data, &configData)
	if err != nil {
		configError = fmt.Errorf("unable to unmarshal config data: %v", err)
		return
	}
}

func BindConfig[T any](container *dig.Container, section string, configStruct T) error {
	once.Do(readConfig)
	if configError != nil {
		return configError
	}

	sectionData, ok := configData[section]
	if !ok {
		return fmt.Errorf("section %s not found in config file", section)
	}
	err := mapstructure.Decode(sectionData, &configStruct)
	if err != nil {
		return fmt.Errorf("unable to unmarshal section data into config struct: %v", err)
	}

	err = container.Provide(func() T {
		return configStruct
	})
	if err != nil {
		return fmt.Errorf("unable to provide config struct to container: %v", err)
	}

	return nil
}
