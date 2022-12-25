package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"jingxi.cn/transitservice/conf"
	"jingxi.cn/transitservice/opensips"
	"jingxi.cn/transitservice/push"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type ProxyConf struct {
	Url  string `json:"url"`
	SUrl string `json:"surl"`
}

type Result struct {
	Status  int    `json:"result"`
	Message string `json:"message"`
}

type Controller struct {
	serverConf *conf.ServerConfig
	sipConf    []byte //sip.json bytebuffer
	confDir    string //conf directory
	proxyConf  ProxyConf
	subscriber *opensips.SubService //opensips subscriber service
	push       *push.PushService    //push to Yunxin
	srv        *http.Server         //http service
	keyword    *push.Keyword        //push message text replace
	rw         sync.RWMutex
}

func NewController(conf *conf.ServerConfig, confDir string) *Controller {
	return &Controller{serverConf: conf,
		sipConf:    nil,
		confDir:    confDir,
		subscriber: nil,
		push:       nil,
		srv:        nil,
		keyword:    nil,
	}
}

func (c *Controller) loadConfig() error {
	sipConf, err := opensips.LoadSipIceTemplate(filepath.Join(c.confDir, "sip.json"))
	if err != nil {
		return err
	}
	sipConf.ReplaceSipServer(c.serverConf.Opensips.SipServer)
	sipConf.ReplaceStunServer(c.serverConf.Opensips.StunServer)

	c.sipConf, err = json.Marshal(sipConf)
	if err != nil {
		return err
	}

	c.keyword, err = push.LoadKeyword(filepath.Join(c.confDir, "message.json"))
	if err != nil {
		return err
	}

	c.subscriber = opensips.NewSubService(c.serverConf)
	c.push = push.NewPushService(c.serverConf)
	c.proxyConf = ProxyConf{
		Url:  c.serverConf.Transit.Url,
		SUrl: c.serverConf.Transit.SUrl,
	}
	return nil
}

func NoResponse(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{
		"status": 404,
		"error":  "404, page not exists!",
	})
}

func (c *Controller) Run(httpAddr string) error {
	err := c.loadConfig()
	if err != nil {
		return err
	}

	router := gin.Default()

	router.GET("/keepalive", c.keepAliveHandlerFunc)
	router.POST("/push", c.pushHandlerFunc)
	router.POST("/opensip/v2/register", c.registerHandlerFunc)
	router.GET("/reload", c.reloadHandlerFunc)
	router.NoRoute(NoResponse)

	c.srv = &http.Server{
		Addr:    httpAddr,
		Handler: router,
	}
	fmt.Printf("http server listen on: %s\n", httpAddr)

	if err = c.srv.ListenAndServe(); err != nil {
		logrus.Errorf("gin ListenAndServe(%s) error: %+v", httpAddr, err)
		return err
	}

	return c.subscriber.Close()
}

func (c *Controller) Stop() {
	if c.srv == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := c.srv.Shutdown(ctx); err != nil {
		logrus.Errorf("gin Shutdown error: %+v", err)
	}
}

func (c *Controller) keepAliveHandlerFunc(ctx *gin.Context) {
	logrus.Infof("/keepalive called")
	var query strings.Builder
	paramPairs := ctx.Request.URL.Query()
	for key, values := range paramPairs {
		query.WriteString(fmt.Sprintf("%v=%v\n", key, values))
	}
	logrus.Infof("%s", query.String())
	ctx.JSON(http.StatusOK, c.proxyConf)
}

func (c *Controller) pushHandlerFunc(ctx *gin.Context) {
	logrus.Infof("/push called")
	var message push.IntercomMessage
	if err := ctx.ShouldBindJSON(&message); err != nil {
		ctx.JSON(http.StatusBadRequest, Result{
			Status:  http.StatusBadRequest,
			Message: "Content empty or Content format invalid",
		})
		return
	}
	//replace title and body
	c.replaceIntercomMessage(&message)

	data, err := c.push.Push(&message)
	if err != nil {
		logrus.Errorf("push to yunxin error: %+v", err)
		ctx.JSON(http.StatusServiceUnavailable, Result{
			Status:  http.StatusServiceUnavailable,
			Message: string(data),
		})
		return
	}
	//https://doc.yunxin.163.com/messaging/docs/jYxMjQ1NTk?platform=server
	//code: 414,403,500
	type PushResult struct {
		Code int    `json:"code"`
		Desc string `json:"desc"`
	}
	var pushResult PushResult
	err = json.Unmarshal(data, &pushResult)
	if err != nil {
		logrus.Errorf("push to yunxin error: %+v", err)
		ctx.JSON(http.StatusServiceUnavailable, Result{
			Status:  http.StatusServiceUnavailable,
			Message: string(data),
		})
		return
	}
	if pushResult.Code != 200 {
		logrus.Errorf("push to yunxin error code:%d", pushResult.Code)
		ctx.JSON(http.StatusServiceUnavailable, Result{
			Status:  pushResult.Code,
			Message: string(data),
		})
		return
	}
	logrus.Infof("push to yunxin success: %s", string(data))
	ctx.String(http.StatusOK, string(data))
}

func (c *Controller) registerHandlerFunc(ctx *gin.Context) {
	logrus.Infof("/v2/register called")
	data := ctx.PostForm("data")
	if len(data) < 1 {
		ctx.JSON(http.StatusBadRequest, Result{
			Status:  http.StatusBadRequest,
			Message: "PostForm(data) empty",
		})
		return
	}
	var r opensips.UserRequest
	err := json.Unmarshal([]byte(data), &r)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, Result{
			Status:  http.StatusBadRequest,
			Message: "Invalid JSON format",
		})
		return
	}
	username := r.User

	if len(username) < 1 {
		username = opensips.CreateUserId(r.Did, r.Client.ClientId, r.Client.SerialNumber)
	}
	user := opensips.NewUser(c.serverConf.Opensips.Domain, username, r.Pwd)

	db_user, err, ok := c.subscriber.GetUser(username)
	if err != nil && !ok {
		//database operation failed
		ctx.JSON(http.StatusInternalServerError, Result{
			Status:  http.StatusInternalServerError,
			Message: "Database operation failed When Query User",
		})
		return
	}
	if err != nil {
		//user does not exist,add it
		err = c.subscriber.AddUser(user)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, Result{
				Status:  http.StatusInternalServerError,
				Message: "Database operation failed When Add User",
			})
			return
		}
		c.createRegisterResponse(ctx, user)
		return
	}
	if !opensips.IsUserValid(db_user, c.serverConf.Opensips.Domain) {
		//user in database not valid,so we update
		err = c.subscriber.UpdateUser(user)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, Result{
				Status:  http.StatusInternalServerError,
				Message: "Database operation failed When update User",
			})
			return
		}
		c.createRegisterResponse(ctx, user)
		return
	}

	//user existed and user valid
	c.createRegisterResponse(ctx, db_user)
}

func (c *Controller) reloadHandlerFunc(ctx *gin.Context) {
	logrus.Infof("/reload called")
	keyword, err := push.LoadKeyword(filepath.Join(c.confDir, "message.json"))
	if err != nil {
		msg := fmt.Sprintf("open %s failed", filepath.Join(c.confDir, "message.json"))
		ctx.JSON(http.StatusInternalServerError, Result{
			Status:  http.StatusInternalServerError,
			Message: msg,
		})
		return
	}
	c.rw.Lock()
	c.keyword = keyword
	c.rw.Unlock()

	ctx.JSON(http.StatusOK, Result{
		Status:  http.StatusOK,
		Message: "success",
	})
}

func (c *Controller) createRegisterResponse(ctx *gin.Context, user *opensips.User) {
	//for deep copy
	var o opensips.SipIceConfig
	err := json.Unmarshal(c.sipConf, &o)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Result{
			Status:  http.StatusInternalServerError,
			Message: "Unmarshal failed",
		})
		return
	}
	o.ReplaceUser(c.serverConf.Opensips.SipServer, user.Username, user.Password)
	ctx.JSON(http.StatusOK, o)
}

func (c *Controller) replaceIntercomMessage(message *push.IntercomMessage) {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.keyword.Replace(message)
}
