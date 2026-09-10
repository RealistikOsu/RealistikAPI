package common

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"github.com/valyala/fasthttp"
)

// MethodData is a struct containing the data passed over to an API method.
type MethodData struct {
	User  Token
	DB    *sqlx.DB
	Doggo *statsd.Client
	R     *redis.Client
	Ctx   *fasthttp.RequestCtx
}

// ClientIP implements a best effort algorithm to return the real client IP, it parses
// X-Real-IP and X-Forwarded-For in order to work properly with reverse-proxies such us: nginx or haproxy.
func (md MethodData) ClientIP() string {
	clientIP := strings.TrimSpace(string(md.Ctx.Request.Header.Peek("X-Real-Ip")))
	if len(clientIP) > 0 {
		return clientIP
	}
	clientIP = string(md.Ctx.Request.Header.Peek("X-Forwarded-For"))
	if index := strings.IndexByte(clientIP, ','); index >= 0 {
		clientIP = clientIP[0:index]
	}
	clientIP = strings.TrimSpace(clientIP)
	if len(clientIP) > 0 {
		return clientIP
	}
	return md.Ctx.RemoteIP().String()
}

// Err logs an error to stdout.
func (md MethodData) Err(err error) {
	_err(err, string(md.Ctx.RequestURI()))
}

// Err for peppy API calls
func Err(c *fasthttp.RequestCtx, err error) {
	_err(err, string(c.RequestURI()))
}

// WSErr is the error function for errors happening in the websockets.
func WSErr(err error) {
	_err(err, "/api/v1/ws")
}

// GenericError is just an error. Can't make a good description.
func GenericError(err error) {
	_err(err, "")
}

func _err(err error, endpoint string) {
	fmt.Println("ERROR!!!!")
	if endpoint != "" {
		fmt.Println(endpoint)
	}
	fmt.Println(err)
}

// ID retrieves the Token's owner user ID.
func (md MethodData) ID() int {
	return md.User.UserID
}

// Query is shorthand for md.C.Query.
func (md MethodData) Query(q string) string {
	return b2s(md.Ctx.QueryArgs().Peek(q))
}

// HasQuery returns true if the parameter is encountered in the querystring.
// It returns true even if the parameter is "" (the case of ?param&etc=etc)
func (md MethodData) HasQuery(q string) bool {
	return md.Ctx.QueryArgs().Has(q)
}

// Unmarshal unmarshals a request's JSON body into an interface.
func (md MethodData) Unmarshal(into interface{}) error {
	return json.Unmarshal(md.Ctx.PostBody(), into)
}

// IsBearer tells whether the current token is a Bearer (oauth) token.
func (md MethodData) IsBearer() bool {
	return md.User.ID == -1
}
