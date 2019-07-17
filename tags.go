package work_wx

import (
	"strings"
)

type Tag struct {
	Name string `json:"tagname"`
	ID   uint64 `json:"tagid"`
}

type tagListRet struct {
	wxErrorRet
	Tags []*Tag `json:"taglist"`
}

func (wx *WorkWx) ListTags() []*Tag {
	url := `https://qyapi.weixin.qq.com/cgi-bin/tag/list?access_token=$ACCESS_TOKEN$`
	url = strings.ReplaceAll(url, "$ACCESS_TOKEN$", wx.AccessToken(AgentAddressbook))

	list := &tagListRet{}
	getJson(url, list)
	return list.Tags
}
