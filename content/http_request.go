package content

import (
	"bytes"
	"context"
	"crypto/sha1"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/enorith/http/contracts"
	"github.com/enorith/supports/byt"
)

type NetHttpRequest struct {
	SimpleParamRequest
	origin       *http.Request
	originWriter http.ResponseWriter
	signature    []byte
	content      []byte
}

func (n *NetHttpRequest) Context() context.Context {
	return n.origin.Context()
}

func (n *NetHttpRequest) OriginWriter() http.ResponseWriter {
	return n.originWriter
}

func (n *NetHttpRequest) Origin() *http.Request {
	return n.origin
}

func (n *NetHttpRequest) Accepts() []byte {
	return n.Header("Accept")
}

func (n *NetHttpRequest) ExceptsJson() bool {
	return byt.Contains(n.Accepts(), []byte("/json"), []byte("+json"))
}

func (n *NetHttpRequest) RequestWithJson() bool {
	return byt.Contains(n.Header("Content-Type"), []byte("application/json"))
}

func (n *NetHttpRequest) IsXmlHttpRequest() bool {
	return bytes.Equal(n.Header("X-Requested-With"), []byte("XMLHttpRequest"))
}

func (n *NetHttpRequest) GetMethod() string {
	return n.origin.Method
}

func (n *NetHttpRequest) GetPathBytes() []byte {
	return []byte(n.origin.URL.Path)
}

func (n *NetHttpRequest) GetUri() []byte {
	return []byte(n.origin.RequestURI)
}

func (n *NetHttpRequest) GetURL() *url.URL {
	return n.origin.URL
}

func (n *NetHttpRequest) Get(key string) []byte {
	q := n.origin.URL.Query().Get(key)

	if q != "" {
		return []byte(q)
	}
	n.origin.ParseForm()
	formData := n.origin.Form.Get(key)

	if formData != "" {
		return []byte(formData)
	}

	return GetJsonValue(n, key)
}

func (n *NetHttpRequest) File(key string) (contracts.UploadFile, error) {
	_, h, err := n.origin.FormFile(key)
	if err != nil {
		return nil, err
	}

	return &uploadFile{header: h}, nil
}

func (n *NetHttpRequest) GetInt(key string) (int, error) {
	str := string(n.Get(key))

	return strconv.Atoi(str)
}

func (n *NetHttpRequest) GetInt64(key string) (int64, error) {
	str := n.GetString(key)

	return strconv.ParseInt(str, 10, 64)
}

func (n *NetHttpRequest) GetUint64(key string) (uint64, error) {
	str := n.GetString(key)

	return strconv.ParseUint(str, 10, 64)
}

func (n *NetHttpRequest) GetString(key string) string {

	return string(n.Get(key))
}

func (n *NetHttpRequest) GetValue(key ...string) contracts.InputValue {
	if len(key) > 0 {
		return n.Get(key[0])
	}

	return n.GetContent()
}

/// GetClientIp get client ip
/// reverse proxy need implement
func (n *NetHttpRequest) GetClientIp() string {
	ip, _, _ := net.SplitHostPort(n.origin.RemoteAddr)

	return ExchangeIpFromProxy(ip, n)
}

func (n *NetHttpRequest) RemoteAddr() string {
	ip, _, _ := net.SplitHostPort(n.origin.RemoteAddr)

	return ip
}

func (n *NetHttpRequest) GetContent() []byte {
	if n.content != nil {
		return n.content
	}

	defer n.origin.Body.Close()
	b, _ := ioutil.ReadAll(n.origin.Body)

	n.content = b
	return b
}

func (n *NetHttpRequest) Unmarshal(to interface{}) error {
	return json.Unmarshal(n.GetContent(), to)
}

func (n *NetHttpRequest) GetSignature() []byte {
	if len(n.signature) > 0 {
		return n.signature
	}

	h := sha1.New()
	var data []byte
	data = append(data, n.GetPathBytes()...)
	data = append(data, n.Origin().Method...)
	data = append(data, n.Header("User-Agent")...)
	data = append(data, n.Authorization()...)
	data = append(data, n.GetClientIp()...)
	data = append(data, n.origin.URL.RawQuery...)
	data = append(data, n.GetContent()...)

	h.Write(data)

	n.signature = h.Sum(nil)

	return n.signature
}

func (n *NetHttpRequest) Header(key string) []byte {
	return []byte(n.HeaderString(key))
}

func (n *NetHttpRequest) HeaderString(key string) string {
	return n.origin.Header.Get(key)
}

func (n *NetHttpRequest) SetHeader(key string, value []byte) contracts.RequestContract {
	n.origin.Header.Set(key, string(value))

	return n
}

func (n *NetHttpRequest) SetHeaderString(key, value string) contracts.RequestContract {
	n.origin.Header.Set(key, value)

	return n
}

func (n *NetHttpRequest) CookieByte(key string) []byte {
	c, e := n.origin.Cookie(key)

	if e == nil {
		return []byte(c.Value)
	}

	return nil
}

func (n *NetHttpRequest) Authorization() []byte {
	return n.Header("Authorization")
}

func (n *NetHttpRequest) BearerToken() ([]byte, error) {
	auth := n.Authorization()

	if len(auth) < 7 {
		return nil, errors.New("invalid bearer token")
	}

	return bytes.TrimSpace(auth[6:]), nil
}

func (n *NetHttpRequest) ToString() string {
	firstLine := fmt.Sprintf("%s %s %s", n.GetMethod(), n.origin.URL, n.origin.Proto)

	return fmt.Sprintf("%s\r\n%s\r\n\r\n%s", firstLine,
		NetHeaderToString(n.origin.Header), n.GetContent())
}

func NewNetHttpRequest(origin *http.Request, w http.ResponseWriter) *NetHttpRequest {
	r := new(NetHttpRequest)
	r.origin = origin
	r.originWriter = w
	r.signature = []byte{}
	return r
}

func NetHeaderToString(h http.Header) string {
	var lines []string
	for k, vs := range h {
		lines = append(lines, fmt.Sprintf("%s: %s", k, strings.Join(vs, ";")))
	}
	return strings.Join(lines, "\r\n")
}
