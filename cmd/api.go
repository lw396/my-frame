package main

import (
	"my-frame/api"
	"my-frame/internal/repository/gorm"
	"my-frame/service"

	"github.com/urfave/cli/v2"
)

var apiCmd = &cli.Command{
	Name:  "api",
	Usage: "启动API服务",
	Flags: []cli.Flag{
		&cli.UintFlag{
			Name:    "port",
			Aliases: []string{"p"},
			Value:   1314,
			Usage:   "端口号",
		},
	},
	Before: func(c *cli.Context) error {
		var err error
		ctx, err = buildContext(c, "api")
		if err != nil {
			return err
		}
		return nil
	},
	Action: func(c *cli.Context) error {
		db, err := ctx.buildDB()
		if err != nil {
			return err
		}

		redis, err := ctx.buildRedis()
		if err != nil {
			return err
		}

		rep := gorm.New(db)

		service := service.New(
			service.WithRepository(rep),
			service.WithRedis(redis),
			service.WithLogger(ctx.buildLogger("SSO")),
		)

		port := c.Int("port")
		api := api.New(api.Config{
			App:  service,
			Port: port,
		})

		return api.Run()
	},
}
