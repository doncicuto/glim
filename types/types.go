/*
Copyright © 2022 Miguel Ángel Álvarez Cabrerizo <mcabrerizo@arrakis.ovh>

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

package types

import (
	"github.com/doncicuto/glim/server/kv"
	"gorm.io/gorm"
)

// Tokens - TODO comment
type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Response - TODO comment
type Response struct {
	TokenType string  `json:"token_type"`
	ExpiresIn float64 `json:"expires_in"`
	ExpiresOn int64   `json:"expires_on"`
	Tokens
}

type LoginBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type APISettings struct {
	DB                 *gorm.DB
	KV                 kv.Store
	TLSCert            string
	TLSKey             string
	Address            string
	APISecret          string
	AccessTokenExpiry  uint
	RefreshTokenExpiry uint
	MaxDaysWoRelogin   int
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RefreshToken - TODO comment
type RefreshToken struct {
	Token string `json:"refresh_token"`
}

type APIError struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}
