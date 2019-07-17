package work_wx

import (
	"strconv"
	"strings"
)

type Department struct {
	ID       uint64 `json:"id"`
	Name     string `json:"name"`
	ParentID uint64 `json:"parentid"`
	Order    uint64 `json:"order"`
}

type departmentListRet struct {
	wxErrorRet
	Departments []*Department `json:"department"`
}

func (wx *WorkWx) ListDepartment(id uint64) []*Department {
	url := `https://qyapi.weixin.qq.com/cgi-bin/department/list?access_token=$ACCESS_TOKEN$`
	url = strings.ReplaceAll(url, "$ACCESS_TOKEN$", wx.AccessToken(AgentAddressbook))
	if id != 0 {
		url = url + "&id=" + strconv.FormatUint(id, 10)
	}
	list := &departmentListRet{}
	getJson(url, list)
	return list.Departments
}
