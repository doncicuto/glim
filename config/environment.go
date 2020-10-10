/*
Copyright © 2020 Miguel Ángel Álvarez Cabrerizo <mcabrerizo@arrakis.ovh>

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

package config

import (
	"os"
	"strconv"
)

// CheckAPISecret - TODO comment
func CheckAPISecret() bool {
	return os.Getenv("API_SECRET") != ""
}

// AccessTokenExpiry - TODO comment
func AccessTokenExpiry() int {
	expiry, err := strconv.Atoi(os.Getenv("ACCESS_TOKEN_EXPIRY_TIME_SECONDS"))
	if err != nil {
		return 3600 // 1 hour default access token expiry time
	}
	return expiry
}

// RefreshTokenExpiry - TODO comment
func RefreshTokenExpiry() int {
	expiry, err := strconv.Atoi(os.Getenv("REFRESH_TOKEN_EXPIRY_TIME_SECONDS"))
	if err != nil {
		return 259200 // 3 days default refresh token expiry time
	}
	return expiry
}

// MaxDaysWoRelogin - TODO comment
func MaxDaysWoRelogin() int {
	maxDays, err := strconv.Atoi(os.Getenv("MAX_DAYS_WITHOUT_RELOGIN"))
	if err != nil {
		return 7 // Max 7 days without relogin
	}
	return maxDays
}
