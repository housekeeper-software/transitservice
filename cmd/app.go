package cmd

import (
	"context"
	"github.com/sirupsen/logrus"
	"jingxi.cn/transitservice/conf"
	"jingxi.cn/transitservice/controller"
	"os"
	"os/signal"
	"syscall"
)

type App struct {
	ctx        context.Context
	cancel     context.CancelFunc
	controller *controller.Controller
}

func NewApp() *App {
	return &App{
		ctx:        nil,
		cancel:     nil,
		controller: nil,
	}
}

func (app *App) Run(serverConf *conf.ServerConfig, confDir string, httpAddr string) {
	app.ctx, app.cancel = context.WithCancel(context.Background())
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c
		app.Quit()
		logrus.Infof("server gracefully shutdown")
	}()

	app.controller = controller.NewController(serverConf, confDir)
	go func() {
		if err := app.controller.Run(httpAddr); nil != err {
			logrus.Errorf("server run failed, err: %+v", err)
			app.Quit()
		}
	}()

Loop:
	for {
		select {
		case <-app.ctx.Done():
			break Loop
		}
	}
	app.controller.Stop()
	logrus.Infof("app quit!")
}

func (app *App) Quit() {
	app.cancel()
}
