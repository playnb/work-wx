package work_wx

import (
	jsoniter "github.com/json-iterator/go"
	"io/ioutil"
	"net/http"
)

var _wx *WorkWx

func init() {
	_wx = &WorkWx{}
}

func WX() *WorkWx {
	return _wx
}

func getJson(url string, ret interface{}) {
	client := &http.Client{}
	req, _ := client.Get(url)
	defer req.Body.Close()
	body, _ := ioutil.ReadAll(req.Body)
	jsoniter.Unmarshal([]byte(body), ret)
}

