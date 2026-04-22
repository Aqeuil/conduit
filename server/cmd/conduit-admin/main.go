package main

import (
	"conduit/internal/conf"
	"conduit/internal/server"
	"conduit/pkg/util"
	"conduit/pkg/xlog"
	"os"
	"path"
	"strings"

	"github.com/go-kratos/kratos/contrib/config/etcd/v2"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/env"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/http"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"

	_ "conduit/internal/plugins/register"

	_ "go.uber.org/automaxprocs"
)

var (
	Name    string
	Version string

	// optional override, e.g. via flags in your project
	flaglogpath string

	id, _ = os.Hostname()
)

const (
	EnvPrefix          = "CONDUIT_"
	EnvEtcdEndpoints   = "ETCD_ENDPOINTS" // from env: CONDUIT_ETCD_ENDPOINTS
	DefaultEtcdCfgPath = "/conduit/config/admin.yaml"
)

func newApp(
	logger log.Logger,
	hs *server.AdminServer,
) *kratos.App {
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			(*http.Server)(hs),
		),
	)
}

// LoadBootstrap does a 2-phase load:
// 1) env only: read etcd endpoints
// 2) env + etcd: load full Bootstrap (env can still override)
func LoadBootstrap() (*conf.Bootstrap, func() error, error) {
	// phase 1: env only
	c1 := config.New(config.WithSource(env.NewSource(EnvPrefix)))
	if err := c1.Load(); err != nil {
		return nil, nil, err
	}
	var endpoints string
	if err := c1.Value(EnvEtcdEndpoints).Scan(&endpoints); err != nil {
		_ = c1.Close()
		return nil, nil, err
	}
	_ = c1.Close()

	// phase 2: etcd source + env
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: strings.Split(endpoints, ","),
	})
	if err != nil {
		return nil, nil, err
	}

	es, err := etcd.New(cli, etcd.WithPath(DefaultEtcdCfgPath))
	if err != nil {
		_ = cli.Close()
		return nil, nil, err
	}

	c2 := config.New(config.WithSource(
		env.NewSource(EnvPrefix), // keep env overlay
		es,
	))
	if err := c2.Load(); err != nil {
		_ = c2.Close()
		_ = cli.Close()
		return nil, nil, err
	}

	var bc conf.Bootstrap
	if err := c2.Scan(&bc); err != nil {
		_ = c2.Close()
		_ = cli.Close()
		return nil, nil, err
	}

	cleanup := func() error {
		_ = c2.Close()
		return cli.Close()
	}
	return &bc, cleanup, nil
}

func main() {
	bc, cfgCleanup, err := LoadBootstrap()
	if err != nil {
		panic(err)
	}
	defer func() { _ = cfgCleanup() }()

	logPath := bc.Log.Path
	if flaglogpath != "" {
		logPath = flaglogpath + "/" + path.Base(logPath)
	}
	_, _ = util.CreatePathDir(logPath)

	xLogger := xlog.NewXLogger(
		xlog.WithLevel(bc.Log.GetLevel()),
		xlog.WithFileName(logPath),
		zap.AddCallerSkip(4),
	)
	defer func() { _ = xLogger.Sync() }()

	logger := log.With(
		xLogger,
		xlog.XesLogCallerKey, xlog.Caller(3),
		xlog.XesLogTraceIDKey, xlog.TraceID(),
	)

	app, cleanup, err := wireApp(bc.Server, bc.Data, bc.Etcd, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	if err := app.Run(); err != nil {
		panic(err)
	}
}
