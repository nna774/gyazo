package gyazo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strconv"

	"golang.org/x/oauth2"
)

const (
	userEndpoint   = "https://api.gyazo.com/api/users/me"
	uploadEndpoint = "https://upload.gyazo.com/api/upload"
)

// Uploader is interface of uploader
type Uploader interface {
	Upload(r io.Reader, metadata *Metadata) (UploadResponse, error)
}

// HTTPAuthorizeConf is Oauth2 auth config
type HTTPAuthorizeConf struct {
	Path string
	Port int
}

// AuthorizeByHTTP is 
func AuthorizeByHTTP(conf *oauth2.Config, hconf *HTTPAuthorizeConf) (*oauth2.Token, error) {
	if hconf == nil {
		hconf = &HTTPAuthorizeConf{Path: "/", Port: 3000}
	}
	url := conf.AuthCodeURL("state", oauth2.AccessTypeOnline)
	fmt.Printf("open %v\n", url)
	codeCh := make(chan string)
	errCh := make(chan error)
	mux := http.NewServeMux()
	srv := &http.Server{Addr: fmt.Sprintf(":%d", hconf.Port), Handler: mux}
	mux.HandleFunc(hconf.Path, func(w http.ResponseWriter, r *http.Request) {
		code := r.FormValue("code")
		if code != "" {
			w.WriteHeader(200)
			w.Write([]byte("ok! back to cli\n"))
			codeCh <- code
		}
	})
	fmt.Printf("waiting callback...\n")
	go func() {
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()
	select {
	case code := <-codeCh:
		srv.Shutdown(context.Background())
		return conf.Exchange(context.Background(), code)
	case err := <-errCh:
		return nil, err
	}
}

// Oauth2Client is Gyazo client which uses oauth2
type Oauth2Client struct {
	AccessToken string
}

// NewOauth2Client creates Oauth2Client
func NewOauth2Client(accessToken string) *Oauth2Client {
	return &Oauth2Client{AccessToken: accessToken}
}

// Client gets *http.Client
func (c *Oauth2Client) Client() *http.Client {
	token := &oauth2.Token{AccessToken: c.AccessToken, TokenType: "bearer"}
	return (&oauth2.Config{}).Client(context.Background(), token)
}

// User is Gyazo user
type User struct {
	Email        string `json:"email"`
	Name         string `json:"name"`
	ProfileImage string `json:"profile_image"`
	UID          string `json:"uid"`
}

// GetCallerIdentity get caller identity
func (c *Oauth2Client) GetCallerIdentity() (User, error) {
	resp, err := c.Client().Get(userEndpoint)
	if err != nil {
		return User{}, nil
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return User{}, nil
	}
	user := struct {
		User User `json:"user"`
	}{}
	err = json.Unmarshal(data, &user)
	return user.User, err
}

// Metadata is Gyazo image metadata
type Metadata struct {
	IsPublic     bool
	CreatedAt    uint32
	RefererURI   string
	Title        string
	Desc         string
	CollectionID string
}

// UploadResponse is Gyazo upload response
type UploadResponse struct {
	ImageID      string `json:"image_id"`
	PermalinkURI string `json:"permalink_url"`
	ThumbURI     string `json:"thumb_url"`
	URI          string `json:"url"`
	Type         string `json:"type"`
}

// Upload uploads to Gyazo
func (c *Oauth2Client) Upload(image io.Reader, metadata *Metadata) (resp UploadResponse, err error) {
	buf := &bytes.Buffer{}
	mw := multipart.NewWriter(buf)
	w, err := mw.CreateFormFile("imagedata", "image")
	if err != nil {
		return
	}
	io.Copy(w, image)
	if metadata != nil {
		w, err = mw.CreateFormField("metadata_is_public")
		if err != nil {
			return
		}
		io.WriteString(w, strconv.FormatBool(metadata.IsPublic))
		if metadata.RefererURI != "" {
			w, err = mw.CreateFormField("referer_url")
			if err != nil {
				return
			}
			io.WriteString(w, metadata.RefererURI)
		}
		// TODO: title, desc, created_at, collection_id
	}
	err = mw.Close()
	if err != nil {
		return
	}
	res, err := c.Client().Post(uploadEndpoint, mw.FormDataContentType(), buf)
	if err != nil {
		return
	}
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &resp)
	return
}
