package helper

import (
	"encoding/json"
	"github.com/urvin/gokaru/internal/server/response"
	"github.com/valyala/fasthttp"
	"io/ioutil"
	"strconv"
	"strings"
)

func ServeError(ctx *fasthttp.RequestCtx, statusCode int, title string) {
	ctx.Response.Reset()
	ctx.SetStatusCode(statusCode)

	httpAccept := string(ctx.Request.Header.Peek(fasthttp.HeaderAccept))
	if strings.Contains(httpAccept, "text/html") {
		tb, err := ioutil.ReadFile("/var/gokaru/assets/error.html")
		if err != nil {
			return
		}

		ts := string(tb)
		ts = strings.Replace(ts, "#CODE#", strconv.Itoa(statusCode), -1)
		ts = strings.Replace(ts, "#TITLE#", title, -1)
		ctx.SetContentType("text/html; charset=utf-8")
		ctx.SetBodyString(ts)
	} else if strings.Contains(httpAccept, "application/json") {
		rsp := response.ErrorResponse{
			Error: title,
		}
		rspb, _ := json.Marshal(rsp)
		ctx.SetContentType("application/json; charset=utf-8")
		ctx.SetBody(rspb)
	} else {
		ctx.SetContentType("text/plain; charset=utf-8")
		ctx.SetBodyString(title)
	}
}
