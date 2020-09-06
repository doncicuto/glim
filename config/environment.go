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
