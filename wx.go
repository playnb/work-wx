package work_wx

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"strings"
	"time"
)

type wxErrorRet struct {
	ErrCode int    `json:"errcode"` //出错返回码，为0表示成功，非0表示调用失败
	ErrMsg  string `json:"errmsg"`  //返回码提示语
}

type AccessToken struct {
	wxErrorRet
	AccessToken string `json:"access_token"` //获取到的凭证，最长为512字节
	ExpiresIn   int64  `json:"expires_in"`   //凭证的有效时间（秒）
}

type WorkWx struct {
	corpID string               //企业号的标识
	secret map[uint64]*WxSecret ///企业号中的应用的Secret
}

//企业微信自建应用的AgentID是1000000开始的
//应该是每个应用都有一个Agent，但是不知道内置的Agent是多少（相关API也不需要填写）
//所以下面的BuildIn只是占位用的，统一格式
//打卡，审批有系统分配的AgentID
const (
	AgentAddressbook = uint64(100)     //通讯录管理secret。在“管理工具”-“通讯录同步”里面查看（需开启“API接口同步”）；
	AgentClientMgr   = uint64(101)     //外部联系人管理secret。在“客户联系”栏，点开“API”小按钮，即可看到。
	AgentClockIn     = uint64(3010011) //打卡
)

type WxSecret struct {
	AgentID     uint64
	Secret      string
	accessToken *AccessToken
}

func (wx *WorkWx) Init(corpID string, secret ...*WxSecret) error {
	wx.corpID = corpID
	wx.secret = make(map[uint64]*WxSecret)
	if len(secret) == 0 {
		return errors.New("至少需要一个可用的应用都没有")
	}
	for _, v := range secret {
		wx.secret[v.AgentID] = v
	}
	for _, v := range wx.secret {
		go func(s *WxSecret) {
			t := time.NewTimer(time.Millisecond)
			for {
				<-t.C
				if err := wx.getAccessToken(s); err == nil {
					t.Reset(time.Second * time.Duration(s.accessToken.ExpiresIn*2/3))
				} else {
					t.Reset(time.Second * 5)
					fmt.Printf("%d 获取Token失败 %s\n", s.AgentID, err.Error())
				}
			}
		}(v)
	}
	return nil
}

func (wx *WorkWx) getAccessToken(secret *WxSecret) error {
	getTokenUrl := `https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=$ID$&corpsecret=$SECRET$`
	getTokenUrl = strings.ReplaceAll(getTokenUrl, "$ID$", wx.corpID)
	getTokenUrl = strings.ReplaceAll(getTokenUrl, "$SECRET$", secret.Secret)
	token := &AccessToken{}
	err := getJson(getTokenUrl, token)
	if err != nil {
		return err
	}
	if token.ErrCode != 0 {
		return errors.New(token.ErrMsg)
	}
	secret.accessToken = token
	fmt.Printf("%d 拉取微信Token: %v\n", secret.AgentID, secret.accessToken)
	return nil
}

func (wx *WorkWx) AccessToken(agentID uint64) string {
	if s, ok := wx.secret[agentID]; ok {
		if s.accessToken != nil {
			return s.accessToken.AccessToken
		}
	}
	return ""
}

func (wx *WorkWx) Message(agentID uint64) *Message {
	msg := &Message{}
	msg.wx = wx
	msg.data = &MessageData{}
	msg.data.AgentID = agentID
	return msg
}

func (wx *WorkWx) ListenCallBack(r gin.IRouter, uri string, token string, encodingAESKey string) *Callback {
	cb := &Callback{}
	cb.Init(wx, token, encodingAESKey)
	r.GET(uri, cb.verifyUrl)
	r.POST(uri, cb.onMessage)
	return cb
}
