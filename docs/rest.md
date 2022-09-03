# REST API

The API documentation is accesible visiting `https://yourserver:1323/swagger/index.html`.

> Note: you may have to add your CA.pem file to your trusted certificates so your browser doesn't complain about security.

Here's the web page offering API documentation.

![swagger_auth_bearer_0.png](./images/swagger_auth_bearer_0.png)

Glim uses JWT tokens to authenticate users so you'll have to request a token from the API using the /v1/login endpoint. Using Glim's Swagger server you can do it following these steps:

1. Go to the login endpoint and click on it

   ![swagger_auth_bearer_login1.png](./images/swagger_auth_bearer_login1.png)

2. Click the "Try it out button". Replace "string" with your username and password to log in and click on "Execute".

   ![swagger_auth_bearer_login2.png](./images/swagger_auth_bearer_login2.png)

3. If your credentials are fine and Glim's working as expected you'll see the access token that you can use in your next requests

   ![swagger_auth_bearer_login3.png](./images/swagger_auth_bearer_login3.png)

4. Now you can copy that token and use it to authenticate your requests. In Swagger you can click on the padlock and a form will be offered to enter that token. Finally click on Authorize to set the authentication token.

   > ⚠️ WARNING: You'll have to put "Bearer " (that's Bearer followed by space) before the token. This is needed as Swagger 2.0 can't use JWT directly. Unfortunately the [swag](https://github.com/swaggo/swag) library used by Glim doesn't support OpenAPI 3.0, so this workaround must be used.

   ![swagger_auth_bearer_1.png](./images/swagger_auth_bearer_1.png)

5. Once you've entered the token, click on the "Close" button. Now you'll see that the padlock icon shows a closed state and your token will be sent with your requests.

   ![swagger_auth_bearer_2.png](./images/swagger_auth_bearer_2.png)
