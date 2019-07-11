package work_wx

import (
	"errors"
	"strings"
	"time"
)

type AccessToken struct {
	ErrCode     int    `json:"errcode"`      //出错返回码，为0表示成功，非0表示调用失败
	ErrMsg      string `json:"errmsg"`       //返回码提示语
	AccessToken string `json:"access_token"` //获取到的凭证，最长为512字节
	ExpiresIn   int64  `json:"expires_in"`   //凭证的有效时间（秒）
}

type WorkWx struct {
	accessToken *AccessToken
	corpID      string //企业号的标识
	corpSecret  string ///企业号中的应用的Secret
}

func (wx *WorkWx) Init(corpID string, corpSecret string) error {
	wx.corpID = corpID
	wx.corpSecret = corpSecret
	wx.getAccessToken()
	if wx.accessToken.ErrCode != 0 {
		return errors.New(wx.accessToken.ErrMsg)
	}
	go func() {
		t := time.NewTimer(time.Second * time.Duration(wx.accessToken.ExpiresIn/2))
		for {
			select {
			case <-t.C:
				wx.getAccessToken()
			}
		}
	}()
	return nil
}

func (wx *WorkWx) getAccessToken() {
	getTokenUrl := `https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=$ID$&corpsecret=$SECRET$`
	getTokenUrl = strings.ReplaceAll(getTokenUrl, "$ID$", wx.corpID)
	getTokenUrl = strings.ReplaceAll(getTokenUrl, "$SECRET$", wx.corpSecret)
	token := &AccessToken{}
	getJson(getTokenUrl, token)
	wx.accessToken = token
}

func (wx *WorkWx) Message(agentID uint64) *Message {
	msg := &Message{}
	msg.wx = wx
	msg.data = &MessageData{}
	msg.data.AgentID = agentID
	return msg
}
