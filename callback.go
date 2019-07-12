package work_wx

import (
	"encoding/base64"
	"encoding/binary"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/url"
	"strconv"
)

type CallbackRequest struct {
	signature string
	nonce     string
	Timestamp int64
	RandomStr string
	RawMsg    string
	ReceiveID string
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

	echostr, _ = url.PathUnescape(echostr)
	src, _ := base64.StdEncoding.DecodeString(echostr)
	dst := AesDecryptCBC([]byte(src), cb.AESKey)

	req.RandomStr = string(dst[:16])
	msgLen := binary.BigEndian.Uint32(dst[16:])
	req.RawMsg = string(dst[20 : 20+msgLen])
	req.ReceiveID = string(dst[20+msgLen:])

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
	if req.parse(cb, c) == false {
		c.AbortWithStatusJSON(http.StatusUnauthorized, "signature验证失败")
		return
	}
	if cb.OnMessage != nil {
		cb.OnMessage(c, req)
	}
}
