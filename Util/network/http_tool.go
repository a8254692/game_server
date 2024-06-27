package network

import (
	"BilliardServer/Util/log"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

func HttpGet(httpURL string) string {
	resp, err := http.Get(httpURL)
	if err != nil {
		log.Error("httpGet,http.Get:", err.Error())
		return ""
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("httpGet,ioutil.ReadAll:", err.Error())
		return ""
	}

	return string(body)
}
func HttpGetState(httpURL string) string {

	resp, err := http.Get(httpURL)
	if err != nil {
		log.Error("httpGet,http.Get:", err.Error())
		return "0"
	}
	defer resp.Body.Close()

	log.Info("返回状态:", resp.Header.Get("result"))
	return resp.Header.Get("result")
}

func HttpPost(postURL string, postType string, postContent string) string {
	resp, err := http.Post(postURL, postType, strings.NewReader(postContent))
	if err != nil {
		log.Error("httpPost,http.Post:", err.Error())
		return ""
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("httpPost,ioutil.ReadAll:", err.Error())
		return ""
	}

	return string(body)
}

// url.Values{"key": {"Value"}, "id": {"123"}}
func HttpPostForm(postURL string, urlValues url.Values) string {
	resp, err := http.PostForm(postURL, urlValues)

	if err != nil {
		log.Error("httpPostForm,http.PostForm:", err.Error())
		return ""
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("httpPostForm,ioutil.ReadAll:", err.Error())
		return ""
	}

	return string(body)
}
func HttpPostState(postURL string, urlValues url.Values) string {

	for i := 0; i < 3; i++ {
		resp, err := http.PostForm(postURL, urlValues)

		if err != nil {
			log.Error("httpPostForm,http.PostForm:", err.Error())
			continue
		}
		defer resp.Body.Close()

		return resp.Header.Get("state")
	}
	return "0"
}

func HttpDo() {
	//	client := &http.Client{}

	//	req, err := http.NewRequest("POST", "http://www.baidu.com", strings.NewReader("name=cjb"))
	//	if err != nil {
	//		// handle error
	//	}

	//	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	//	req.Header.Set("Cookie", "name=anny")

	//	resp, err := client.Do(req)

	//	defer resp.Body.Close()

	//	body, err := ioutil.ReadAll(resp.Body)
	//	if err != nil {
	//		// handle error
	//	}

	//	fmt.Println(string(body))
}
