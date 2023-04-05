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
	once              sync.Once
	privicyOnce       sync.Once
	configData        map[string]interface{}
	privicyConfigData map[string]interface{}
	configError       error
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

// 草草加了一个privacy的方法，这个代码主要是gpt写的，懒得详细整，反正只有两个文件
func readPrivacyConfig() {
	data, err := os.ReadFile("user_config.yml")
	if err != nil {
		configError = fmt.Errorf("unable to read config file，你可能是没把user_config.yml.example的后缀去掉: %v", err)
		return
	}

	err = yaml.Unmarshal(data, &privicyConfigData)
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

func BindPrivicyConfig[T any](container *dig.Container, section string, configStruct T) error {
	privicyOnce.Do(readPrivacyConfig)
	if configError != nil {
		return configError
	}

	sectionData, ok := privicyConfigData[section]
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
