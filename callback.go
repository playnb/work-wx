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

type wxBody struct {
	ToUserName string `xml:"ToUserName"`
	AgentID    string `xml:"AgentID"`
	Encrypt    string `xml:"Encrypt"`
}

type CallbackRequest struct {
	signature string
	nonce     string
	Timestamp int64
	RandomStr string
	RawMsg    string
	ReceiveID string
	Body      *wxBody
	MapData   map[string]string
}

func (req *CallbackRequest) IsText() bool {
	return req.GetMsgType() == "text"
}

func (req *CallbackRequest) GetContent() string {
	if c, ok := req.MapData["Content"]; ok {
		return c
	}
	return ""
}
func (req *CallbackRequest) GetAgentID() uint64 {
	if c, ok := req.MapData["AgentID"]; ok {
		aid, _ := strconv.ParseUint(c, 10, 64)
		return aid
	}
	return 0
}
func (req *CallbackRequest) GetMsgType() string {
	if c, ok := req.MapData["MsgType"]; ok {
		return c
	}
	return ""
}

func (req *CallbackRequest) parse(cb *Callback, c *gin.Context) bool {
	signature, _ := c.GetQuery("msg_signature")
	timestamp, _ := c.GetQuery("timestamp")
	nonce, _ := c.GetQuery("nonce")
	echostr, hasEchoStr := c.GetQuery("echostr")

	req.nonce = nonce
	req.Timestamp, _ = strconv.ParseInt(timestamp, 10, 64)
	req.signature = signature
	_ = echostr

	req.MapData = make(map[string]string)

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

		req.Body = &wxBody{}
		xml.Unmarshal(data, req.Body)
		src, _ := base64.StdEncoding.DecodeString(req.Body.Encrypt)
		dst := AesDecryptCBC([]byte(src), cb.AESKey)
		fmt.Println(dst)

		req.RandomStr = string(dst[:16])
		msgLen := binary.BigEndian.Uint32(dst[16:])
		req.RawMsg = string(dst[20 : 20+msgLen])
		req.ReceiveID = string(dst[20+msgLen:])

		xml.Unmarshal([]byte(req.RandomStr), req.MapData)
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
