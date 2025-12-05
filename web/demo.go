package web

import (
	"fmt"
	"io"
	"net/http"
)

// readBodyOnce
// 最重要的特征就是只能读取一次，无法重复读取。
// 本质上 Request Body 是一种接近 Stream 的设计。
func readBodyOnce(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "read body failed: %v", err)
		return
	}

	// 类型转换，将 []byte 转换为 string
	fmt.Fprintf(w, "read the data: %s \n", string(body))

	// 尝试再次读取，啥也读不到，但是也不会报错
	body, err = io.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "read the data one more time got error %v", err)
		return
	}

	fmt.Fprintf(w, "read the data one more time: [%s] and read data length %d \n", string(body), len(body))
}

func getBodyIsNil(w http.ResponseWriter, r *http.Request) {
	if r.GetBody == nil {
		fmt.Fprintf(w, "GetBody is nil \n")
	} else {
		fmt.Fprintf(w, "GetBody is not nil \n")
	}
}

func queryParams(w http.ResponseWriter, r *http.Request) {
	values := r.URL.Query()
	fmt.Fprintf(w, "query is %v \n", values)
}

func header(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "header is %v \n", r.Header)
}

func form(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "before parse form: %v \n", r.Form)
	err := r.ParseForm()
	if err != nil {
		fmt.Fprintf(w, "parse form error: %v \n", r.Form)
	}
	fmt.Fprintf(w, "after parse form: %v \n", r.Form)
}
