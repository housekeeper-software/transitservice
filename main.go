package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
	"io"
	"jingxi.cn/transitservice/cmd"
	"jingxi.cn/transitservice/conf"
	"jingxi.cn/transitservice/utils"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"time"
)

type CmdLine struct {
	isDebug  bool
	httpAddr string
	pprof    string
	confDir  string
	logDir   string

	sipServer  string
	stunServer string
	domain     string
	proxyUrl   string
	proxySUrl  string
	dbUrl      string
	dbTable    string
}

var cmdline CmdLine

func init() {
	flag.BoolVar(&cmdline.isDebug, "debug", false, "true")
	flag.StringVar(&cmdline.httpAddr, "http", "9999", "0.0.0.0:8080")
	flag.StringVar(&cmdline.pprof, "pprof", "", "0.0.0.0:6060")
	flag.StringVar(&cmdline.confDir, "conf", "", "/app/conf")
	flag.StringVar(&cmdline.logDir, "log", "", "/app/log")

	flag.StringVar(&cmdline.sipServer, "sip", "", "1.1.1.1:18888")
	flag.StringVar(&cmdline.stunServer, "stun", "", "1.1.1.1:18888")
	flag.StringVar(&cmdline.domain, "domain", "", "1.1.1.1")
	flag.StringVar(&cmdline.proxyUrl, "proxyurl", "", "http://1.1.1.1:11888")
	flag.StringVar(&cmdline.proxySUrl, "proxySUrl", "", "http://1.1.1.1:11889")
	flag.StringVar(&cmdline.dbUrl, "dbUrl", "", "opensips:opensipsrw@tcp(1.1.1.1:3306)/opensips")
	flag.StringVar(&cmdline.dbTable, "dbTable", "", "subscriber")
}

func runProfServer() {
	err := http.ListenAndServe(cmdline.pprof, nil)
	if err != nil {
		logrus.Error("start pprof server %s failed: %+v\n", cmdline.pprof, err)
	}
}

func initLog() {
	path := filepath.Join(cmdline.logDir, "transit_service.log")
	writer, _ := rotatelogs.New(
		path+".%Y%m%d%H%M",
		rotatelogs.WithLinkName(path),
		rotatelogs.WithMaxAge(time.Duration(24)*time.Hour),
		rotatelogs.WithRotationTime(time.Duration(8)*time.Hour))

	if cmdline.isDebug {
		writers := []io.Writer{
			writer,
			os.Stdout,
		}
		fileAndStdoutWriter := io.MultiWriter(writers...)
		logrus.SetOutput(fileAndStdoutWriter)
		logrus.SetLevel(logrus.InfoLevel)
	} else {
		logrus.SetOutput(writer)
		logrus.SetLevel(logrus.ErrorLevel)
	}
	logrus.SetReportCaller(true)
}

func setArgsFromCommandLine(conf *conf.ServerConfig) {
	if len(cmdline.sipServer) > 0 {
		conf.Opensips.SipServer = cmdline.sipServer
	}
	if len(cmdline.stunServer) > 0 {
		conf.Opensips.StunServer = cmdline.stunServer
	}
	if len(cmdline.domain) > 0 {
		conf.Opensips.Domain = cmdline.domain
	}
	if len(cmdline.proxyUrl) > 0 {
		conf.Transit.Url = cmdline.proxyUrl
	}
	if len(cmdline.proxySUrl) > 0 {
		conf.Transit.SUrl = cmdline.proxySUrl
	}
	if len(cmdline.dbUrl) > 0 {
		conf.Mysql.Url = cmdline.dbUrl
	}
	if len(cmdline.dbTable) > 0 {
		conf.Mysql.Table = cmdline.dbTable
	}
}

func main() {
	flag.Parse()

	initLog()

	if len(cmdline.httpAddr) < 1 {
		logrus.Fatalf("http listen address empty")
		return
	}
	if len(cmdline.pprof) > 0 {
		fmt.Printf("pprof listen on: %s\n", cmdline.pprof)
		go runProfServer()
	}

	yaml := filepath.Join(cmdline.confDir, "transit_service_conf.yaml")
	serverConf, err := conf.LoadServerConfig(yaml)
	if err != nil {
		logrus.Fatalf("read file(%s) failed: %+v", yaml, err)
		return
	}
	setArgsFromCommandLine(serverConf)

	_ = serverConf.WriteConfigToFile(filepath.Join(cmdline.logDir, "current.json"))
	_ = utils.SaveAppStartTime(cmdline.logDir)

	if !cmdline.isDebug {
		gin.SetMode(gin.ReleaseMode)
	}
	app := cmd.NewApp()
	app.Run(serverConf, cmdline.confDir, cmdline.httpAddr)
}
