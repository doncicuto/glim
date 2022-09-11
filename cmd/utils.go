/*
Copyright © 2022 Miguel Ángel Álvarez Cabrerizo <mcabrerizo@sologitops.com>

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

package cmd

import (
	"bufio"
	b64 "encoding/base64"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/viper"
)

func truncate(text string, length int) string {
	if len(text) > length {
		format := fmt.Sprintf("%%.%ds...", length-3)
		return fmt.Sprintf(format, text)
	}
	return text
}

func RestClient(token string) *resty.Client {
	// Rest API authentication
	client := resty.New()

	// Set bearer token
	if token != "" {
		client.SetAuthToken(token)
	}
	tlscacert := viper.GetString("tlscacert")
	client.SetRootCertificate(tlscacert)
	return client
}

func JPEGToBase64(path string) (*string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(f)
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	encoded := b64.StdEncoding.EncodeToString(content)
	return &encoded, nil
}
