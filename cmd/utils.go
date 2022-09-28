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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/doncicuto/glim/models"
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

type JSONSuccessOutput struct {
	Message string `json:"message"`
}

type JSONErrorOutput struct {
	ErrorMessage string `json:"error"`
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

func printError(errorMessage string, jsonOutput bool) {
	if jsonOutput {
		output := JSONErrorOutput{}
		enc := json.NewEncoder(os.Stderr)

		if errorMessage != "" {
			output.ErrorMessage = errorMessage
		}

		enc.Encode(output)
	} else {
		fmt.Println(errorMessage)
	}
}

func printMessage(message string, jsonOutput bool) {
	if jsonOutput {
		output := JSONSuccessOutput{}
		enc := json.NewEncoder(os.Stdout)

		if message != "" {
			output.Message = message
		}

		enc.Encode(output)
	} else {
		fmt.Println(message)
	}
}

func printCSVMessages(messages []string, jsonOutput bool) {
	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.Encode(messages)
	} else {
		for _, message := range messages {
			fmt.Print(message)
		}
	}
}

func encodeUserToJson(user *models.UserInfo) {
	enc := json.NewEncoder(os.Stdout)
	enc.Encode(user)
}

func encodeUsersToJson(users *[]models.UserInfo) {
	enc := json.NewEncoder(os.Stdout)
	enc.Encode(users)
}

func encodeGroupToJson(group *models.GroupInfo) {
	enc := json.NewEncoder(os.Stdout)
	enc.Encode(group)
}

func encodeGroupsToJson(groups *[]models.GroupInfo) {
	enc := json.NewEncoder(os.Stdout)
	enc.Encode(groups)
}
