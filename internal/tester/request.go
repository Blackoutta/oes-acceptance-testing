package tester

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"oes-acceptance-testing/v1/pkg/helpers/conf"
	"oes-acceptance-testing/v1/pkg/helpers/sign"
	"strconv"
	"time"
)

const (
	PLATFORM_DMP string = "1"
	PLATFORM_ECP string = "2"
)

type Request struct {
	Req    *http.Request
	Method string
	Body   string
	URL    string
}

func NewRequest(path, method, platform string, body interface{}, query URLQuery) *Request {
	plb, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}

	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)

	param := url.Values{}
	for k, v := range query {
		param.Set(k, v)
	}
	param.Set("accessKeyId", conf.AccessKeyID)
	param.Set("signatureNonce", strconv.Itoa(r1.Intn(9999999999)))
	param.Set("signature", sign.RequestSign(conf.Secret, method, &param, string(plb)))
	requestURL := conf.BaseURL + path + "?" + param.Encode()

	b := bytes.NewReader(plb)

	r, err := http.NewRequest(method, requestURL, b)
	if err != nil {
		panic(err)
	}

	r.Header.Set("platform", platform)
	r.Header.Set("Content-Type", "application/json")
	r.Close = true

	return &Request{
		Req:    r,
		Method: method,
		Body:   string(plb),
		URL:    requestURL,
	}
}

func (t *Tester) Send(res interface{}) (Record, error) {
	var rec Record
	resp, err := t.Client.Do(t.Request.Req)
	if err != nil {
		return rec, err
	}

	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return rec, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return rec, fmt.Errorf("响应状态码不为200，响应为: \n%v\nURL为: %v", string(bs), t.Request.URL)
	}

	rec.ResponseString = string(bs)

	if err := json.Unmarshal(bs, res); err != nil {
		return rec, err
	}

	tt, ok := res.(Response)
	if !ok {
		log.Fatalln("req: failed type assertion")
	}

	rec.Response = tt
	rec.ReqURL = t.Request.URL
	rec.ReqBody = t.Request.Body
	rec.ReqMethod = t.Request.Method

	return rec, nil
}

type URLQuery map[string]string

func (u URLQuery) Add(key, value string) {
	u[key] = value
}
