### 1. N/A

1. route definition

- Url: /api/v1/login
- Method: POST
- Request: `LoginReq`
- Response: `LoginResp`

2. request definition



```golang
type LoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
```


3. response definition



```golang
type LoginResp struct {
	AccessToken string `json:"access_token"`
	SessionId string `json:"session_id"`
	ExpiresIn int64 `json:"expires_in"`
	TokenType string `json:"token_type"`
}
```

### 2. N/A

1. route definition

- Url: /api/v1/logout
- Method: POST
- Request: `-`
- Response: `LogoutResp`

2. request definition



3. response definition



```golang
type LogoutResp struct {
	Ok bool `json:"ok"`
	Message string `json:"message"`
}
```

### 3. N/A

1. route definition

- Url: /api/v1/logout-all
- Method: POST
- Request: `-`
- Response: `LogoutResp`

2. request definition



3. response definition



```golang
type LogoutResp struct {
	Ok bool `json:"ok"`
	Message string `json:"message"`
}
```

### 4. N/A

1. route definition

- Url: /api/v1/me
- Method: GET
- Request: `-`
- Response: `MeResp`

2. request definition



3. response definition



```golang
type MeResp struct {
	Uid string `json:"uid"`
	Jti string `json:"jti"`
	Iat int64 `json:"iat"`
}
```

### 5. N/A

1. route definition

- Url: /api/v1/ping
- Method: GET
- Request: `-`
- Response: `EmptyResp`

2. request definition



3. response definition



```golang
type EmptyResp struct {
}
```

### 6. N/A

1. route definition

- Url: /api/v1/refresh
- Method: POST
- Request: `-`
- Response: `LoginResp`

2. request definition



3. response definition



```golang
type LoginResp struct {
	AccessToken string `json:"access_token"`
	SessionId string `json:"session_id"`
	ExpiresIn int64 `json:"expires_in"`
	TokenType string `json:"token_type"`
}
```

### 7. N/A

1. route definition

- Url: /api/v1/user/info
- Method: GET
- Request: `-`
- Response: `UserInfoResp`

2. request definition



3. response definition



```golang
type UserInfoResp struct {
	UserId string `json:"user_id"`
	Username string `json:"username"`
	Email string `json:"email"`
	DisplayName string `json:"display_name"`
	AvatarUrl string `json:"avatar_url"`
}
```

