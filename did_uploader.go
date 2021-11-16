package gyazo

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
)

// DeviceIDUploader uploads to Gyazo with device id
type DeviceIDUploader struct {
	did string
}

// NewDIDUploader creates new DeviceIDUploader
func NewDIDUploader(did string) *DeviceIDUploader {
	return &DeviceIDUploader{did: did}
}

// Upload uploads to Gyazo
func (u *DeviceIDUploader) Upload(image io.Reader, metadata *UploadMetadata) (resp *UploadResponse, err error) {
	ct, body, err := createRequestBody(image, metadata)
	res, err := http.Post(didUploadEndpoint(), ct, body)
	if err != nil {
		return
	}
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	if res.StatusCode != http.StatusOK {
		err = errors.New(string(data))
		return
	}
	resp.PermalinkURI = string(data)
	resp.DeviceID = res.Header.Get("X-Gyazo-ID")
	return
}

// make sure DeviceIDUploader implements Uploader interface
// https://golang.org/doc/effective_go.html#blank_implements
var _ Uploader = (*DeviceIDUploader)(nil)
