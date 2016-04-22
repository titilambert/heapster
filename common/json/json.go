// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package json

import (
	"fmt"
	"net/url"
//	"strconv"
    "os"
    "bufio"

//	"k8s.io/heapster/version"
    "github.com/golang/glog"
)

type JsonClient interface {
	WriteString(string) (int, error)
    Flush() (error)
}

type JsonConfig struct {
	Path     string
}

func NewClient(c JsonConfig) (JsonClient, error) {

    f, err := os.Create(c.Path)
    glog.Infof("Created file %f", c.Path)

	if err != nil {
		return nil, err
	}
    writer := bufio.NewWriter(f)

	return writer, nil
}

func BuildConfig(uri *url.URL) (*JsonConfig, error) {
	config := JsonConfig{
        Path:     "",
	}

	if uri.Scheme != "" && uri.Scheme != "file" {
		return nil, fmt.Errorf("Bad file scheme: %s", uri.Scheme)
	}

	if len(uri.Path) > 0 {
		config.Path = uri.Path
	} else {
		return nil, fmt.Errorf("Bad file path")
    }

	return &config, nil
}
