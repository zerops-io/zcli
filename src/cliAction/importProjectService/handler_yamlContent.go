package importProjectService

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/zerops-io/zcli/src/i18n"
)

func getImportYamlContent(config RunConfig) ([]byte, error) {
	workingDir, err := filepath.Abs(config.WorkingDir)
	if err != nil {
		return nil, err
	}

	fmt.Println("working dir is", workingDir)

	if config.ImportYamlPath == nil {
		return nil, errors.New("no path to yaml")
	}

	fmt.Println("yaml path", *config.ImportYamlPath)

	importYamlPath := path.Join(workingDir, *config.ImportYamlPath)
	fmt.Println("PATH ", importYamlPath)

	importYamlStat, err := os.Stat(importYamlPath)
	if err != nil {
		if os.IsNotExist(err) {
			if config.ImportYamlPath != nil {
				return nil, errors.New(i18n.ImportYamlNotFound)
			}
		}
		return nil, nil
	}

	fmt.Printf("%s: %s\n", i18n.ImportYamlFound, importYamlPath)

	if importYamlStat.Size() == 0 {
		return nil, errors.New(i18n.ImportYamlEmpty)
	}
	// TODO ask if the size is ok for this yaml (might be larger than zerops.yaml)
	if importYamlStat.Size() > 10*1024 {
		return nil, errors.New(i18n.ImportYamlTooLarge)
	}

	yamlContent, err := os.ReadFile(importYamlPath)
	if err != nil {
		return nil, err
	}

	return yamlContent, nil
}