package handlers

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/doncicuto/glim/common"
)

func TestUserUpdate(t *testing.T) {
	// Setup
	h, e, settings := testSetup(t, false)
	defer testCleanUp()

	// Log in with admin, search and/or plain user and get tokens
	adminToken, _ := getUserTokens("admin", h, e, settings)
	searchToken, _ := getUserTokens("search", h, e, settings)
	plainUserToken, _ := getUserTokens("saul", h, e, settings)

	jpegPhoto := "\"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KKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigApD7daWg0Adv4N1ATWL2LnLwH5c91PpXVLXlWk6g2manDcgnYDhx/eU/5z+FepxlWQMpDAgYPrXzuYUfZ1eZbS1R7ODq89NR7D6KKK4TrPHaKKK+wPmwooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiij1AQjPGM16D4R1D7XpfkOwMtudmf7y/wn8uPwrz+tLQL/wDs3V45GYiJ/kf0wSOfwOK48bQdSlbqjpwtX2dS62Z6eOaUUxe/P5U8V82tdUe31sjx2iiivsD5sKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKRhnuRwRkdv8/4UtBoaGj0jwxqP9oaSpc/voj5cg75A4P5Y/WtoV5v4X1H7Dqqxu2Irj5GPof4T+fH416MvevmcZS9nVaWz1R7WFq+0pq+60Z4/RRRX0x4gUUUUAFFFFABRRRQAUUUUAFFFFABRRRQAUUUUAFFFFABRRRQAUUUUAFFFFABRRRQAUUUUAFFFFABRRRQAUUUUAFFFFABRRRQAUUUUAFFFFABRRRQAUUUUAFFFFABRRRQAUUUUDEOQdwJG3nP+f88V6hoWoDUtKimJBkA2yD0Ydf8AGvLyfp689q6LwfqJtdRe0kbCT8j034/rn9K4Mwoc9O63R1YOq4T9TnqKKK7zkCiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigApUdo5EkUkMrAg+hyKSg0NJ6NDXkFFFFAgooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACigdP696KHrsC7oKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigC9rOn/wBm6rNAFIjJ3x+m09B+HI/CqIruvGOnfadPW8QfPActjuvf8uK4QdP8/wCf/wBdcmCq+0pJvdaM6MTS9nUcRaKKK6znCiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigD//2Q==\""

	// Test cases
	testCases := []RestTestCase{
		{
			name:             "search user can't update accounts",
			expResCode:       http.StatusForbidden,
			reqURL:           "/v1/users/3",
			reqMethod:        http.MethodPut,
			secret:           searchToken,
			expectedBodyJSON: fmt.Sprintf(`{"message":"%s"}`, common.UserHasNoProperPermissionsMessage),
		},
		{
			name:             "plainuser can't update other's accounts",
			expResCode:       http.StatusForbidden,
			reqURL:           "/v1/users/4",
			reqMethod:        http.MethodPut,
			secret:           plainUserToken,
			expectedBodyJSON: fmt.Sprintf(`{"message":"%s"}`, common.UserHasNoProperPermissionsMessage),
		},
		{
			name:             "non-existent manager user can't update account info",
			expResCode:       http.StatusForbidden,
			reqURL:           "/v1/users/5",
			reqMethod:        http.MethodPut,
			secret:           "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJhcGkuZ2xpbS5zZXJ2ZXIiLCJleHAiOjE5NzcyNDUzOTksImlhdCI6MTY2MTYyNjA3MSwiaXNzIjoiYXBpLmdsaW0uc2VydmVyIiwianRpIjoiZTdiZmYzMjQtMzJmOC00MTNlLTgyNmYtNzc5Mzk5NDBjOTZkIiwibWFuYWdlciI6dHJ1ZSwicmVhZG9ubHkiOmZhbHNlLCJzdWIiOiJhcGkuZ2xpbS5jbGllbnQiLCJ1aWQiOjEwMDB9.amq5OV7gU7HUrn5YA8sbs2cXMRFeYHTmXm6bhXJ9PDg",
			expectedBodyJSON: `{"message":"wrong user attempting to update account"}`,
		},
		{
			name:             "uid must be an integer",
			expResCode:       http.StatusNotAcceptable,
			reqURL:           "/v1/users/wrong",
			reqMethod:        http.MethodPut,
			secret:           adminToken,
			expectedBodyJSON: `{"message":"uid param should be a valid integer"}`,
		},
		{
			name:             "non-existent accounts can't be updated",
			expResCode:       http.StatusNotFound,
			reqURL:           "/v1/users/3000",
			reqMethod:        http.MethodPut,
			secret:           adminToken,
			expectedBodyJSON: fmt.Sprintf(`{"message":"%s"}`, common.UserNotFoundMessage),
		},
		{
			name:             "only managers can update a username",
			expResCode:       http.StatusForbidden,
			reqURL:           "/v1/users/3",
			reqMethod:        http.MethodPut,
			secret:           plainUserToken,
			reqBodyJSON:      `{"username":"walter"}`,
			expectedBodyJSON: `{"message":"only managers can update the username"}`,
		},
		{
			name:             "only managers can update a username",
			expResCode:       http.StatusNotAcceptable,
			reqURL:           "/v1/users/3",
			reqMethod:        http.MethodPut,
			secret:           adminToken,
			reqBodyJSON:      `{"username":"kim"}`,
			expectedBodyJSON: `{"message":"username cannot be duplicated"}`,
		},
		{
			name:             "email must be valid",
			expResCode:       http.StatusNotAcceptable,
			reqURL:           "/v1/users/3",
			reqMethod:        http.MethodPut,
			secret:           plainUserToken,
			reqBodyJSON:      `{"email":"wrong"}`,
			expectedBodyJSON: fmt.Sprintf(`{"message":"%s"}`, common.InvalidEmail),
		},
		{
			name:             common.OnlyManagersUpdateManagerMessage,
			expResCode:       http.StatusForbidden,
			reqURL:           "/v1/users/3",
			reqMethod:        http.MethodPut,
			secret:           plainUserToken,
			reqBodyJSON:      `{"manager":true}`,
			expectedBodyJSON: fmt.Sprintf(`{"message":"%s"}`, common.OnlyManagersUpdateManagerMessage),
		},
		{
			name:             common.OnlyManagersUpdateReadonlyMessage,
			expResCode:       http.StatusForbidden,
			reqURL:           "/v1/users/3",
			reqMethod:        http.MethodPut,
			secret:           plainUserToken,
			reqBodyJSON:      `{"readonly":true}`,
			expectedBodyJSON: fmt.Sprintf(`{"message":"%s"}`, common.OnlyManagersUpdateReadonlyMessage),
		},
		{
			name:             common.OnlyManagersUpdateLockedMessage,
			expResCode:       http.StatusForbidden,
			reqURL:           "/v1/users/3",
			reqMethod:        http.MethodPut,
			secret:           plainUserToken,
			reqBodyJSON:      `{"locked":true}`,
			expectedBodyJSON: fmt.Sprintf(`{"message":"%s"}`, common.OnlyManagersUpdateLockedMessage),
		},
		{
			name:             "plainuser can update her acount",
			expResCode:       http.StatusOK,
			reqURL:           "/v1/users/3",
			reqMethod:        http.MethodPut,
			secret:           plainUserToken,
			reqBodyJSON:      fmt.Sprintf(`{"firstname":"saul","lastname":"goodman","email":"new@email.com","ssh_public_key":"key","jpeg_photo":%s}`, jpegPhoto),
			expectedBodyJSON: fmt.Sprintf(`{"uid":3,"username":"saul","name":"saul goodman","firstname":"saul","lastname":"goodman","email":"new@email.com","ssh_public_key":"key","jpeg_photo":%s,"manager":false,"readonly":false,"locked":false}`, jpegPhoto),
		},
	}

	for _, tc := range testCases {
		runTests(t, tc, e)
	}
}
