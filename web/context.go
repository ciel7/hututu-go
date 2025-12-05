package web

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
)

type Context struct {
	Req        *http.Request
	Resp       http.ResponseWriter
	PathParams map[string]string

	queryValues url.Values
}

// RespJSONOK JSON成功响应
func (c *Context) RespJSONOK(val any) error {
	return c.RespJSON(http.StatusOK, val)
}

// RespJSON JSON响应
func (c *Context) RespJSON(status int, val any) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}
	c.Resp.WriteHeader(status)
	c.Resp.Header().Set("Content-Type", "application/json")
	c.Resp.Header().Set("Content-Length", strconv.Itoa(len(data)))
	n, err := c.Resp.Write(data)
	if err != nil || n != len(data) {
		//return err
		return errors.New("web: 未写入全部数据")
	}
	return err
}

func (c *Context) SetCookie(ck *http.Cookie) {
	http.SetCookie(c.Resp, ck)
}

// BindJSON
func (c *Context) BindJSON(val any) error {
	if val == nil {
		return errors.New("web: 输入不能为 nil")
	}
	if c.Req.Body == nil {
		return errors.New("web: body为 nil")
	}
	//bs, _ := io.ReadAll(c.Req.Body)
	//json.Unmarshal(bs, val)
	decoder := json.NewDecoder(c.Req.Body)
	//decoder.UseNumber()             // 设置之后，数字会用 number 表示，否则默认是 float64
	//decoder.DisallowUnknownFields() // 如果有未知字段，就会报错
	return decoder.Decode(val) // 必须传指针，否则解析后数据无法生效
}

func (c *Context) BindJSONOpt(val any, useNumber bool, disableUnknown bool) error {
	if c.Req.Body == nil {
		return errors.New("web: body 为 nil")
	}

	decoder := json.NewDecoder(c.Req.Body)
	if useNumber {
		decoder.UseNumber()
	}
	if disableUnknown {
		decoder.DisallowUnknownFields()
	}
	return decoder.Decode(val)
}

func (c *Context) FormValue(key string) (string, error) {
	err := c.Req.ParseForm()
	if err != nil {
		return "", err
	}

	// 方案1
	//vals, ok := c.Req.Form[key]
	//if !ok {
	//	return "", errors.New("web: key 不存在")
	//}
	//return vals[0], nil

	// 方案2
	return c.Req.FormValue(key), nil
}

// QueryValue Query 和 Post 比起来存在的问题是 Query 没有在首次ParseForm的时候做缓存
// 每次调用 c.Req.URL.Query().Get(key)，都会去 make 新的存储空间
// 所以我们可以在上层自己增加一个用来存储的字段
// 直接调用 c.Req.URL.Query() 用户区别不出来以下两种情况：
// 1. 有值但为空
// 2. 没有获取到值
func (c *Context) QueryValue(key string) (string, error) {
	if c.queryValues == nil {
		c.queryValues = c.Req.URL.Query()
	}

	//return c.queryValues.Get(key), nil

	vals, exist := c.queryValues[key]
	if !exist || len(vals) == 0 {
		return "", errors.New("web: key 不存在")
	}
	return vals[0], nil
}

// PathValue 路径参数
func (c *Context) PathValue(key string) (string, error) {
	val, exist := c.PathParams[key]
	if !exist {
		return "", errors.New("web: key 不存在")
	}
	return val, nil
}

type StringValue struct {
	val string
	err error
}

// StringValue
// 好处是 可以直接给 StringValue 的变量加很多转化的方法，用户想用的时候直接调用即可
func (s StringValue) AsInt64() (int64, error) {
	if s.err != nil {
		return 0, s.err
	}
	return strconv.ParseInt(s.val, 10, 64)
}

// QueryValueV1
func (c *Context) QueryValueV1(key string) StringValue {
	if c.queryValues == nil {
		c.queryValues = c.Req.URL.Query()
	}

	//return c.queryValues.Get(key), nil

	vals, exist := c.queryValues[key]
	if !exist || len(vals) == 0 {
		return StringValue{
			err: errors.New("web: key 不存在"),
		}
	}
	return StringValue{
		val: vals[0],
	}
}

// PathValueV1 路径参数
func (c *Context) PathValueV1(key string) StringValue {
	val, exist := c.PathParams[key]
	if !exist {
		return StringValue{
			err: errors.New("web: key 不存在"),
		}
	}
	return StringValue{
		val: val,
	}
}
