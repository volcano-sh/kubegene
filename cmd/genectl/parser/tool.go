/*
Copyright 2018 The Kubegene Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package parser

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/spf13/cobra"
	kubeyaml "k8s.io/apimachinery/pkg/util/yaml"
	"kubegene.io/kubegene/cmd/genectl/util"
)

func FetchTools(cmd *cobra.Command) (map[string]Tool, error) {
	toolRepo := util.GetFlagString(cmd, "tool-repo")

	// remote tool repo
	if strings.Index(toolRepo, "http://") == 0 || strings.Index(toolRepo, "https://") == 0 {
		bytes, err := DownloadToolFile(toolRepo)
		if err != nil {
			return nil, fmt.Errorf("fetch remote repo tools error: %v", err)
		}
		tools, err := ParseTool(bytes)
		if err != nil {
			return nil, fmt.Errorf("parse remote repo tools error: %v", err)
		}

		return TransTools2Map(tools), nil
	}

	// local tool repo
	toolFiles, err := GetAllToolFile(toolRepo)
	if err != nil {
		return nil, fmt.Errorf("fetch local repo tools error: %v", err)
	}

	totalTools := make([]Tool, 0)
	for _, toolFile := range toolFiles {
		body, err := ioutil.ReadFile(toolFile)
		if err != nil {
			return nil, fmt.Errorf("fetch local repo tools error: %v", err)
		}
		tools, err := ParseTool(body)
		if err != nil {
			return nil, fmt.Errorf("parse local repo tools error: %v", err)
		}
		totalTools = append(totalTools, tools...)
	}

	return TransTools2Map(totalTools), nil
}

func TransTools2Map(tools []Tool) map[string]Tool {
	toolsMap := make(map[string]Tool, len(tools))

	for _, tool := range tools {
		key := tool.Name + ":" + tool.Version
		toolsMap[key] = tool
	}

	return toolsMap
}

func ValidateToolAttr(tool Tool) error {
	if len(tool.Name) == 0 {
		return errors.New("tool Name is required")
	}
	if len(tool.Version) == 0 {
		return errors.New("tool version is required")
	}
	if len(tool.Image) == 0 {
		return errors.New("tool image is required")
	}
	return nil
}

func ParseTool(data []byte) ([]Tool, error) {
	var tools []Tool
	var empty = Tool{}
	reader := bytes.NewReader(data)

	// We store tools as a YAML stream; there may be more than one decoder.
	yamlDecoder := kubeyaml.NewYAMLOrJSONDecoder(reader, 512*1024)
	for {
		tool := Tool{}
		err := yamlDecoder.Decode(&tool)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("cannot parse tool: %v", err)
		}
		if reflect.DeepEqual(tool, empty) {
			continue
		}
		if err := ValidateToolAttr(tool); err != nil {
			return nil, fmt.Errorf("invalidate tool: %v", err)
		}
		tools = append(tools, tool)
	}

	return tools, nil
}

func GetAllToolFile(rootPath string) ([]string, error) {
	list := make([]string, 0)
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".yaml" {
			list = append(list, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk error [%v]\n", err)
	}
	return list, nil
}

func DownloadToolFile(url string) ([]byte, error) {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// ready body
	return ioutil.ReadAll(resp.Body)
}
