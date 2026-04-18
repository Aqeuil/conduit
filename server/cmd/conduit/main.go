package main

import (
	"conduit/internal/biz/manager"
	"conduit/internal/conf"
	"conduit/internal/server"
	"conduit/pkg/util"
	"conduit/pkg/xlog"
	"flag"
	"os"
	"path"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/http"
	"go.uber.org/zap"

	_ "conduit/internal/plugins/register"

	_ "go.uber.org/automaxprocs"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string
	// Version is the version of the compiled software.
	Version string
	// flagconf is the config flag.
	flagconf    string
	flaglogpath string

	id, _ = os.Hostname()
)

func init() {
	flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf config.yaml")
	flag.StringVar(&flaglogpath, "logpath", "", "config path, eg: -logpath ./logs/")
}

func newApp(
	logger log.Logger,
	hs *server.HttpServer,
	admin *server.ConduitServer,
	m *manager.Manager,
) *kratos.App {
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			(*http.Server)(hs),
			(*http.Server)(admin),
			m,
		),
	)
}

func main() {
	flag.Parse()
	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	logPath := bc.Log.Path
	if flaglogpath != "" {
		logPath = flaglogpath + "/" + path.Base(logPath)
	}

	_, _ = util.CreatePathDir(logPath)

	fileName := logPath
	xLogger := xlog.NewXLogger(
		xlog.WithLevel(bc.Log.GetLevel()),
		xlog.WithFileName(fileName),
		zap.AddCallerSkip(4),
	)
	defer func() {
		_ = xLogger.Sync()
	}()
	logger := log.With(xLogger,
		xlog.XesLogCallerKey, xlog.Caller(3),
		xlog.XesLogTraceIDKey, xlog.TraceID(),
	)
	app, cleanup, err := wireApp(bc.Server, bc.Etcd, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}
