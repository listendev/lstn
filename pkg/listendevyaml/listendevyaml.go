package listendevyaml

import (
	"errors"
	"os"
	"path"

	"gopkg.in/yaml.v3"
)

var (
	ErrNoConfigFileFound      = errors.New("no config file found")
	ErrCouldNotReadConfigFile = errors.New("could not read config file")
	ErrCouldNotParseConfig    = errors.New("could not parse config file")
)

type NPM struct {
	Lockfiles []string `yaml:"lockfiles"`
}

type Listendevyaml struct {
	NPM NPM `yaml:"npm"`
}

func Parse(data []byte) (*Listendevyaml, error) {
	var ldy Listendevyaml
	err := yaml.Unmarshal(data, &ldy)
	if err != nil {
		return nil, err
	}
	return &ldy, nil
}

func SearchAndLoadConfigFile(workingDirectory string) (*Listendevyaml, error) {
	var configFilePayload []byte
	possiblePaths := []string{".listendev/config.yaml", ".listendev/config.yml", ".listendev.yaml", ".listendev.yml"}

	for _, curpath := range possiblePaths {
		if _, err := os.Stat(curpath); err == nil {
			configFilePayload, err = os.ReadFile(path.Join(workingDirectory, curpath))
			if err != nil {
				return nil, errors.Join(err, ErrCouldNotReadConfigFile)
			}
			break
		}
	}

	if len(configFilePayload) == 0 {
		return nil, ErrNoConfigFileFound
	}

	configFile, err := Parse(configFilePayload)
	if err != nil {
		return nil, errors.Join(err, ErrCouldNotParseConfig)
	}

	return configFile, nil
}
