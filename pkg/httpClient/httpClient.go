package httpClient

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type OesHttpClient struct {
	ProductId    string
	DeviceId     string
	DeviceSecret string
	Token        string
	ServerAddr   string
	Client       http.Client
}

func NewOesHttpClient(productId, deviceId int, deviceSecret, serverAddr string) *OesHttpClient {
	pidStr := strconv.Itoa(productId)
	didStr := strconv.Itoa(deviceId)
	token := encryptPassword(pidStr, didStr, deviceSecret)
	return &OesHttpClient{
		ProductId:    pidStr,
		DeviceId:     didStr,
		DeviceSecret: deviceSecret,
		Token:        token,
		ServerAddr:   serverAddr,
		Client: http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// NewUpwardRequest takes in a json string as input and composes a new http request
// that is suitable for data uploading on OES platform
// payload example:
// {"temperature": 22.5, "humidity": 55.5}
func (c *OesHttpClient) NewUpwardRequest(host string, payload string) *http.Request {
	query := url.Values{}
	query.Set("clientid", c.DeviceId)
	query.Set("username", c.ProductId)
	query.Set("password", c.Token)

	pl := strings.NewReader(payload)
	url := host + "/api/v1/device/data/upload?" + query.Encode()
	r, err := http.NewRequest(http.MethodPost, url, pl)
	if err != nil {
		log.Fatalf("error while composing new request: %v", err)
	}
	return r
}

func encryptPassword(pid, did, apiKey string) string {
	query := did + "&" + pid

	h := hmac.New(sha1.New, []byte(apiKey))
	h.Write([]byte(query))

	token := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return token
}
