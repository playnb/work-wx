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

func getJson(url string, ret interface{}) error {
	client := &http.Client{}
	req, err := client.Get(url)
	if err != nil {
		return err
	}
	defer req.Body.Close()
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}
	jsoniter.Unmarshal([]byte(body), ret)
	return nil
}
