package tools

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"BilliardServer/Util/log"

	"github.com/pkg/errors"
)

// http get
func HttpGet(httpURL string) string {
	resp, err := http.Get(httpURL)
	if err != nil {
		return ""
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	return string(body)
}

// url.Values{"key": {"Value"}, "id": {"123"}}
func HttpPostForm(postURL string, urlValues url.Values) string {
	resp, err := http.PostForm(postURL, urlValues)

	if err != nil {
		return ""
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	return string(body)
}

func Http(requestType, url, content string) (error, []byte) {
	//创建一个请求
	result := ""
	req, err := http.NewRequest(requestType, url, strings.NewReader(content))
	if err != nil {
		result = "发送http请求异常：" + err.Error()
		log.Error(result)
		return errors.New(result), nil
	}

	client := &http.Client{}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err := client.Do(req)
	if err != nil {
		result = "发送http请求失败：" + err.Error()
		log.Error(result)
		return errors.New(result), nil
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return nil, body
}
