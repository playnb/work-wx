package work_wx

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/url"
	"strconv"
)

type CallbackRequest struct {
	signature string
	nonce     string
	timestamp int64
	randomStr string
	msg       string
	receiveID string
}

func (req *CallbackRequest) Parse(cb *Callback, c *gin.Context) bool {
	signature, _ := c.GetQuery("msg_signature")
	timestamp, _ := c.GetQuery("timestamp")
	nonce, _ := c.GetQuery("nonce")
	echostr, _ := c.GetQuery("echostr")

	req.nonce = nonce
	req.timestamp, _ = strconv.ParseInt(timestamp, 10, 64)
	req.signature = signature
	_ = echostr

	echostr, _ = url.PathUnescape(echostr)
	src, _ := base64.StdEncoding.DecodeString(echostr)
	dst := AesDecryptCBC([]byte(src), cb.AESKey)

	req.randomStr = string(dst[:16])
	msgLen := binary.BigEndian.Uint32(dst[16:])
	req.msg = string(dst[20 : 20+msgLen])
	req.receiveID = string(dst[20+msgLen:])

	//消息体签名校验	{dev_msg_signature=sha1(sort(token、timestamp、nonce、msg_encrypt))}
	//TODO:: 之后再实现
	return true
}

type Callback struct {
	Token  string
	AESKey []byte
	wx     *WorkWx
}

func (cb *Callback) Init(wx *WorkWx, token string, encodingAESKey string) {
	cb.wx = wx
	cb.Token = token
	cb.AESKey, _ = base64.StdEncoding.DecodeString(encodingAESKey + "=")
}

//验证URL有效性
func (cb *Callback) VerifyUrl(c *gin.Context) {
	req := &CallbackRequest{}
	if req.Parse(cb, c) == false {
		c.AbortWithStatusJSON(http.StatusUnauthorized, "signature验证失败")
		return
	}
	c.String(http.StatusOK, req.msg)
}

//接收消息请求
func (cb *Callback) OnMessage(c *gin.Context) {
	req := &CallbackRequest{}
	if req.Parse(cb, c) == false {
		c.AbortWithStatusJSON(http.StatusUnauthorized, "signature验证失败")
		return
	}
	fmt.Println("--------------")
	fmt.Println(req.msg)
	fmt.Println("--------------")
}
