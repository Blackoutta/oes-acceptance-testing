package sign

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"net/url"
	"sort"
	"strings"
)

type signer struct {
	// 签名凭证
	//credential *Credential
	// 签名标签
	label string
}

func getSignerDefault() *signer {
	return &signer{label: "HMAC-SHA1"}
}

func (signer *signer) signString(stringToSign string, accessKeySecret string) string {
	return hmacSHA1(stringToSign, accessKeySecret)
}

// RequestSignature sig request params
// urlParams URL field's query parameters eg: http://www.xxx.com?key1=value&key2=value
// then urlParams is key1=value&key2=value
// body request's body contents
func RequestSign(accessKeySecret string,
	httpMethod string, urlParams *url.Values, data ...string) string {
	// make canonicalize query string
	canonicalStr := canonicalizeQueryString(urlParams, data...)
	// make makeStringToSign
	stringToSign := makeStringToSign(httpMethod, canonicalStr)
	signedStr := sign(accessKeySecret, stringToSign)
	return signedStr
}

func canonicalizeQueryString(urlParams *url.Values, data ...string) string {
	// make canonicalize query string
	paramContent := ""
	if urlParams != nil {
		urlSortEncode := encode(*urlParams)
		paramContent = urlSortEncode
	}

	// sort encode formParams
	if len(data) > 0 {
		paramContent += data[0]
	}

	canonicalizeStr := percentEncode(paramContent)
	return canonicalizeStr
}

// percentEncode utf-8 url编码
func percentEncode(value string) string {
	strEncode := url.QueryEscape(value)
	strEncode = strings.Replace(strEncode, "+", "%20", -1)
	strEncode = strings.Replace(strEncode, "*", "%2A", -1)
	strEncode = strings.Replace(strEncode, "%7E", "~", -1)
	return strEncode
}

// makeStringToSign return a string contain request method and params that be canonicalized
func makeStringToSign(httpMethod string, canonicalizeStr string) string {
	StringToSign :=
		httpMethod + "&" + //HTTPMethod：发送请求的 HTTP 方法，例如 GET。
			percentEncode("/") + "&" + //percentEncode("/")：字符（/）UTF-8 编码得到的值，即 %2F。
			canonicalizeStr //规范化请求字符串。
	return StringToSign
}

// sign use a signer to make a signature
func sign(accessKeySecret string, stringToSign string) string {
	signer := getSignerDefault()
	return signer.signString(stringToSign, accessKeySecret)
}

func encode(v map[string][]string) string {
	if v == nil {
		return ""
	}
	var buf strings.Builder
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		vs := v[k]
		keyEscaped := k
		for _, v := range vs {
			if buf.Len() > 0 {
				buf.WriteByte('&')
			}
			buf.WriteString(keyEscaped)
			buf.WriteByte('=')
			buf.WriteString(v)
		}
	}
	return buf.String()
}

func hmacSHA1(stringToSign string, priKey string) string {
	key := []byte(priKey)
	mac := hmac.New(sha1.New, key)
	mac.Write([]byte(stringToSign))

	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return signature
}
