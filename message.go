package work_wx

import (
	"bytes"
	jsoniter "github.com/json-iterator/go"
	"io/ioutil"
	"net/http"
	"strings"
)

type MessageRet struct {
	ErrCode      int    `json:"errcode"`
	ErrMsg       string `json:"errmsg"`
	invaliduser  string `json:"invaliduser"`  //"userid1|userid2", // 不区分大小写，返回的列表都统一转为小写
	invalidparty string `json:"invalidparty"` //"partyid1|partyid2",
	invalidtag   string `json:"invalidtag"`   //"tagid1|tagid2"
}

type Text struct {
	Content string `json:"content"`
}
type Markdown struct {
	Content string `json:"content"`
}
type Image struct {
	MediaID string `json:"media_id"`
}
type Voice struct {
	MediaID string `json:"media_id"`
}
type Video struct {
	MediaID     string `json:"media_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}
type File struct {
	MediaID string `json:"media_id"`
}
type TextCard struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Url         string `json:"url"`
	BtnTxt      string `json:"btntxt"`
}

//图文消息
//图文消息（mpnews）
//markdown消息
//小程序通知消息
//任务卡片消息

type MessageData struct {
	ToUser  string `json:"touser"`  //"UserID1|UserID2|UserID3"
	ToParty string `json:"toparty"` //"PartyID1|PartyID2"
	ToTag   string `json:"totag"`   //"TagID1 | TagID2"
	MsgType string `json:"msgtype"` //text
	AgentID uint64 `json:"agentid"`
	Safe    uint64 `json:"safe"`

	Text     *Text     `json:"text,omitempty"`
	Markdown *Markdown `json:"markdown,omitempty"`
	Image    *Image    `json:"image,omitempty"`
	Voice    *Voice    `json:"voice,omitempty"`
	Video    *Video    `json:"video,omitempty"`
	File     *File     `json:"file,omitempty"`
	TextCard *TextCard `json:"textcard,omitempty"`
}
type Message struct {
	data *MessageData
	wx   *WorkWx
}

func (m *Message) AgentID(agentID uint64) *Message {
	m.data.AgentID = agentID
	return m
}

func (m *Message) ToUser(users ...string) *Message {
	for _, v := range users {
		if len(m.data.ToUser) > 0 {
			m.data.ToUser += "|"
		}
		m.data.ToUser += v
	}
	return m
}

func (m *Message) ToParty(parties ...string) *Message {
	for _, v := range parties {
		if len(m.data.ToParty) > 0 {
			m.data.ToParty += "|"
		}
		m.data.ToParty += v
	}
	return m
}
func (m *Message) ToTag(tags ...string) *Message {
	for _, v := range tags {
		if len(m.data.ToTag) > 0 {
			m.data.ToTag += "|"
		}
		m.data.ToTag += v
	}
	return m
}

func (m *Message) Text(text *Text) *Message {
	m.data.Text = text
	m.data.MsgType = "text"
	return m
}
func (m *Message) Markdown(md *Markdown) *Message {
	m.data.Markdown = md
	m.data.MsgType = "markdown"
	return m
}
func (m *Message) TextCard(tc *TextCard) *Message {
	m.data.TextCard = tc
	m.data.MsgType = "textcard"
	return m
}

func (m *Message) Send() *MessageRet {
	ret := &MessageRet{}

	url := `https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=$ACCESS_TOKEN$`
	url = strings.ReplaceAll(url, "$ACCESS_TOKEN$", m.wx.accessToken.AccessToken)
	client := &http.Client{}
	body, _ := jsoniter.Marshal(m.data)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		ret.ErrCode = -1
		ret.ErrMsg = "发送请求失败"
		return ret
	}
	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(req.Body)
	jsoniter.Unmarshal([]byte(data), ret)
	return ret
}
