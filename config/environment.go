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

// Package config implements utility routines for handling environment variables
// like GLIM_API_SECRET, ACCESS_TOKEN_EXPIRY_TIME_SECONDS, REFRESH_TOKEN_EXPIRY_TIME_SECONDS
// and MAX_DAYS_WITHOUT_RELOGIN and setting default values.
package config

import (
	"os"
	"strconv"
)

// CheckAPISecret returns a boolean value representing if the GLIM_API_SECRET environment
// variable has been set
func CheckAPISecret() bool {
	return os.Getenv("GLIM_API_SECRET") != ""
}

// AccessTokenExpiry returns the number of seconds for access token expiration if the
// ACCESS_TOKEN_EXPIRY_TIME_SECONDS environment variable has been set. If the environment
// variable hasn't been found, the function returns 3600 seconds.
func AccessTokenExpiry() int {
	expiry, err := strconv.Atoi(os.Getenv("ACCESS_TOKEN_EXPIRY_TIME_SECONDS"))
	if err != nil {
		return 3600 // 1 hour default access token expiry time
	}
	return expiry
}

// RefreshTokenExpiry returns the number of seconds for access token expiration if the
// REFRESH_TOKEN_EXPIRY_TIME_SECONDS environment variable has been set. If the environment
// variable hasn't been found, the function returns 259200 seconds (3 days).
func RefreshTokenExpiry() int {
	expiry, err := strconv.Atoi(os.Getenv("REFRESH_TOKEN_EXPIRY_TIME_SECONDS"))
	if err != nil {
		return 259200 // 3 days default refresh token expiry time
	}
	return expiry
}

// MaxDaysWoRelogin returns the number of days that we can use refresh tokens without log in
// again according to the MAX_DAYS_WITHOUT_RELOGIN environment variable. If the environment
// variable hasn't been found, the function returns 7 days.
func MaxDaysWoRelogin() int {
	maxDays, err := strconv.Atoi(os.Getenv("MAX_DAYS_WITHOUT_RELOGIN"))
	if err != nil {
		return 7 // Max 7 days without relogin
	}
	return maxDays
}
