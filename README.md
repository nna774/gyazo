# Gyazo

[![PkgGoDev](https://pkg.go.dev/badge/github.com/nna774/gyazo)](https://pkg.go.dev/github.com/nna774/gyazo)

upload image to Gyazo from golang

## usage

### with oauth(recommended)

#### before use

1. create your own application from [https://gyazo.com/oauth/applications](https://gyazo.com/oauth/applications).
2. get access token from application setting or use `gyazo.AuthorizeByHTTP` or [oauth2](https://pkg.go.dev/golang.org/x/oauth2)

#### code

```go
accessToken := "xxx"
imageReader, err := os.Open("nana.png")
if err != nil {
	panic(err)
}
client := gyazo.NewOauth2Client(accessToken)
result, err := client.Upload(imageReader, nil)
fmt.Printf("image_id: %s\npermalink: %s\nthumbnail: %s\nurl: %s\ntype: %s\ncreated_at %s",
	result.ImageID, // 0123456789abcdef0123456789abcdef
	result.PermalinkURL, // https://gyazo.com/0123456789abcdef0123456789abcdef
	result.ThumbURL, // https://thumb.gyazo.com/thumb/200/...
	result.URL, // https://i.gyazo.com/0123456789abcdef0123456789abcdef.png
	result.Type, // png
	result.CreatedAt, // 2038-01-19T03:14:07+0000
)
```

### with device id

#### most simple, only upload

```go
uploader := gyazo.NewDIDUploader("")
result, err := uploader.Upload(imageReader, nil)
if err != nil {
    panic(err)
}
fmt.Println(result.PermalinkURL) // https://gyazo.com/0123456789abcdef0123456789abcdef
fmt.Println(result.DeviceID) // 0123456789abcdef0123456789abcdef
```

#### with your device id

Note: keep your device id a secret, it is like your Gyazo password

Your device id is stored at `%appdata%\Gyazo\id.txt` on Windows.

```go
uploader := gyazo.NewDIDUploader(yourDeviceID)
```

This uploader uploads to your Gyazo account.
