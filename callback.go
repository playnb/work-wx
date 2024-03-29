package work_wx

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

type wxBodyHead struct {
	ToUserName string `xml:"ToUserName"`
	AgentID    string `xml:"AgentID"`
	Encrypt    string `xml:"Encrypt"`
}

type wxBody struct {
	ToUserName   string `xml:"ToUserName"`
	FromUserName string `xml:"FromUserName"`
	CreateTime   string `xml:"CreateTime"`
	MsgType      string `xml:"MsgType"`

	MsgId   string `xml:"MsgId"`
	AgentID string `xml:"AgentID"`

	PicUrl  string `xml:"PicUrl"`
	MediaId string `xml:"MediaId"`
	Content string `xml:"Content"`
}

type CallbackRequest struct {
	signature string
	nonce     string
	Timestamp int64
	RandomStr string
	RawMsg    string
	ReceiveID string
	BodyHead  *wxBodyHead
	Body      *wxBody
}

func (req *CallbackRequest) IsText() bool {
	return req.Body.MsgType == "text"
}
func (req *CallbackRequest) GetAgentID() uint64 {
	if req.Body != nil {
		aid, _ := strconv.ParseUint(req.Body.AgentID, 10, 64)
		return aid
	}
	return 0
}

func (req *CallbackRequest) parse(cb *Callback, c *gin.Context) bool {
	signature, _ := c.GetQuery("msg_signature")
	timestamp, _ := c.GetQuery("timestamp")
	nonce, _ := c.GetQuery("nonce")
	echostr, _ := c.GetQuery("echostr")

	req.nonce = nonce
	req.Timestamp, _ = strconv.ParseInt(timestamp, 10, 64)
	req.signature = signature
	_ = echostr

	req.Body = &wxBody{}

	if c.Request.Method == http.MethodGet {
		echostr, _ = url.PathUnescape(echostr)
		src, _ := base64.StdEncoding.DecodeString(echostr)
		dst := AesDecryptCBC([]byte(src), cb.AESKey)

		req.RandomStr = string(dst[:16])
		msgLen := binary.BigEndian.Uint32(dst[16:])
		req.RawMsg = string(dst[20 : 20+msgLen])
		req.ReceiveID = string(dst[20+msgLen:])
	}
	if c.Request.Method == http.MethodPost {
		defer c.Request.Body.Close()
		data, _ := ioutil.ReadAll(c.Request.Body)

		req.BodyHead = &wxBodyHead{}
		xml.Unmarshal(data, req.BodyHead)
		src, _ := base64.StdEncoding.DecodeString(req.BodyHead.Encrypt)
		dst := AesDecryptCBC([]byte(src), cb.AESKey)
		fmt.Println(dst)

		req.RandomStr = string(dst[:16])
		msgLen := binary.BigEndian.Uint32(dst[16:])
		req.RawMsg = string(dst[20 : 20+msgLen])
		req.ReceiveID = string(dst[20+msgLen:])

		xml.Unmarshal([]byte(req.RawMsg), req.Body)
	}
	//消息体签名校验	{dev_msg_signature=sha1(sort(token、timestamp、nonce、msg_encrypt))}
	//TODO:: 之后再实现
	return true
}

type Callback struct {
	Token  string
	AESKey []byte
	wx     *WorkWx

	OnMessage func(*gin.Context, *CallbackRequest)
}

func (cb *Callback) Init(wx *WorkWx, token string, encodingAESKey string) {
	cb.wx = wx
	cb.Token = token
	cb.AESKey, _ = base64.StdEncoding.DecodeString(encodingAESKey + "=")
}

//验证URL有效性
func (cb *Callback) verifyUrl(c *gin.Context) {
	req := &CallbackRequest{}
	if req.parse(cb, c) == false {
		c.AbortWithStatusJSON(http.StatusUnauthorized, "signature验证失败")
		return
	}
	c.String(http.StatusOK, req.RawMsg)
}

//接收消息请求
func (cb *Callback) onMessage(c *gin.Context) {
	req := &CallbackRequest{}
	fmt.Println(c.Request.RequestURI)

	if req.parse(cb, c) == false {
		c.AbortWithStatusJSON(http.StatusUnauthorized, "signature验证失败")
		return
	}
	if cb.OnMessage != nil {
		cb.OnMessage(c, req)
	}
}
