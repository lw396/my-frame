package main

import (
	"fmt"
	"my-frame/pkg/db"
	"my-frame/pkg/log"
	"my-frame/pkg/redis"
	"my-frame/pkg/snowflake"
	"my-frame/pkg/valuer"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
	"gopkg.in/ini.v1"
	"gorm.io/gorm"
)

type Context struct {
	*cli.Context
	cfg         *ini.File
	appName     string
	environment string
	podId       uint
}

func (c *Context) Section(name string) *ini.Section {
	return c.cfg.Section(name)
}

func buildContext(c *cli.Context, appName string) (*Context, error) {
	environment := getEnv()
	name := strings.ToLower(appName)
	configDir := c.String("config-dir")

	logger := log.NewConsoleLogger("LOAD")
	logger.Infof("当前环境: %s", environment)
	logger.Infof("当前应用: %s", name)
	logger.Infof("配置文件目录: %s", configDir)
	logger.Infof("开始加载配置文件...")

	fileNames := []string{
		"app.cfg",
		fmt.Sprintf("app.%s.cfg", environment),
		fmt.Sprintf("%s.cfg", name),
		fmt.Sprintf("%s.%s.cfg", name, environment),
	}

	var sources []interface{}
	for _, fileName := range fileNames {
		logger.Info("加载配置文件: ", fileName)
		sources = append(sources, filepath.Join(configDir, fileName))
	}

	opt := ini.LoadOptions{
		Loose:                   true,
		SkipUnrecognizableLines: true,
	}

	cfg := ini.Empty(opt)
	if len(sources) > 0 {
		var err error
		cfg, err = ini.LoadSources(opt, sources[0], sources[1:]...)
		if err != nil {
			return nil, err
		}
	}

	return &Context{
		Context:     c,
		cfg:         cfg,
		appName:     name,
		environment: environment,
		podId:       c.Uint("pod-id"),
	}, nil
}

func getEnv() string {
	environment := strings.ToLower(os.Getenv("MY_FRAME_ENV"))

	if environment == "" {
		environment = "develop"
	}
	return environment
}

func (c *Context) IsDebug() bool {
	return c.environment == "develop"
}

func (c *Context) buildLogger(scope string) log.Logger {
	if c.IsDebug() {
		return log.NewConsoleLogger(scope)
	}

	return log.NewLogger(log.Config{
		App:    c.appName,
		Scope:  scope,
		LogDir: c.String("log-dir"),
	})
}

func (c *Context) buildDB() (*gorm.DB, error) {
	host := valuer.Value("127.0.0.1").Try(
		os.Getenv("MYSQL_HOST"), c.Section("mysql").Key("host").String(),
	).String()
	port := valuer.Value("3306").Try(
		os.Getenv("MYSQL_PORT"), c.Section("mysql").Key("port").String(),
	).String()
	name := valuer.Value("my-frame").Try(
		os.Getenv("MYSQL_DB"), c.Section("mysql").Key("db").String(),
	).String()
	user := valuer.Value("root").Try(
		os.Getenv("MYSQL_USER"), c.Section("mysql").Key("user").String(),
	).String()
	password := valuer.Value("password").Try(
		os.Getenv("MYSQL_PASSWORD"), c.Section("mysql").Key("password").String(),
	).String()
	timezone := valuer.Value("UTC").Try(
		os.Getenv("MYSQL_TIMEZONE"), c.Section("mysql").Key("timezone").String(),
	).String()

	loc := url.QueryEscape(timezone)
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=%s",
		user, password, host, port, name, loc,
	)

	idGen, err := snowflake.NewWithConfig(snowflake.Config{
		StartTime:    1648684800000,
		WorkerIDBits: 5,
		SequenceBits: 12,
		WorkerID:     int(c.podId),
	})
	if err != nil {
		return nil, err
	}

	return db.New(
		db.WithDSN(dsn),
		db.WithIDGenerator(idGen),
		db.WithLogger(c.buildLogger("DB")),
	)
}

func (c *Context) buildRedis() (redis.RedisClient, error) {
	host := valuer.Value("127.0.0.1").Try(
		os.Getenv("REDIS_HOST"), c.Section("redis").Key("host").String(),
	).String()
	port := valuer.Value(6379).Try(
		os.Getenv("REDIS_PORT"), c.Section("redis").Key("port").MustInt(),
	).Int()
	password := valuer.Value("secret").Try(
		os.Getenv("REDIS_AUTH"), c.Section("redis").Key("auth").String(),
	).String()
	db := valuer.Value(0).Try(
		os.Getenv("REDIS_DB"), c.Section("redis").Key("db").MustInt(),
	).Int()

	return redis.NewClient(
		redis.WithAddress(host, port),
		redis.WithAuth("", password),
		redis.WithDB(db),
	)
}
