package push

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"jingxi.cn/transitservice/conf"
	"jingxi.cn/transitservice/utils"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type PushService struct {
	serverConf *conf.ServerConfig
}

func NewPushService(conf *conf.ServerConfig) *PushService {
	return &PushService{
		serverConf: conf,
	}
}

//remove tail char '_'
func getAccount(pushId string) string {
	for i := len(pushId) - 1; i >= 0; i-- {
		if pushId[i] == '_' {
			return pushId[:i]
		}
	}
	return pushId
}

//https://doc.yunxin.163.com/messaging/docs/jYxMjQ1NTk?platform=server
func (p *PushService) Push(message *IntercomMessage) ([]byte, error) {
	attachBytes, err := CreateAttach(p.serverConf.Push.MsgTag, message)
	if err != nil {
		return nil, err
	}
	payloadBytes, err := CreatePayload(message)
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Add("attach", string(attachBytes))
	params.Add("pushcontent", message.Body)
	params.Add("payload", string(payloadBytes))
	params.Add("msgtype", "0") //0：点对点自定义通知
	params.Add("from", p.serverConf.Push.AppAccid)
	params.Add("to", getAccount(message.Device))

	if len(p.serverConf.Push.Save) > 0 {
		params.Add("save", p.serverConf.Push.Save)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequest("POST", p.serverConf.Push.SendAttachMsgUrl, bytes.NewBuffer([]byte(params.Encode())))
	if err != nil {
		return nil, err
	}
	p.addHeader(req)

	resp, err := client.Do(req)
	if err != nil {
		logrus.Errorf("Request(%+v) error: %+v", params, err)
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

//https://doc.yunxin.163.com/TM5MzM5Njk/docs/jk3MzY2MTI?platform=server
func (p *PushService) addHeader(req *http.Request) {
	nonce := utils.RandString(16)
	curTime := strconv.FormatInt(time.Now().Unix(), 10)
	chechSum := checkSum(p.serverConf.Push.AppSecret, nonce, curTime)
	req.Header.Add("AppKey", p.serverConf.Push.AppKey)
	req.Header.Add("Nonce", nonce)
	req.Header.Add("CurTime", curTime)
	req.Header.Add("CheckSum", chechSum)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded;charset=utf-8")
}

//SHA1(AppSecret + Nonce + CurTime)
func checkSum(a string, b string, c string) string {
	var builder strings.Builder
	builder.WriteString(a)
	builder.WriteString(b)
	builder.WriteString(c)
	o := sha1.New()
	o.Write([]byte(builder.String()))
	return hex.EncodeToString(o.Sum(nil))
}
