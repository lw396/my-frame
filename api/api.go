package api

import (
	"fmt"
	"my-frame/service"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Api struct {
	port    int
	service *service.Service
}

type Config struct {
	App  *service.Service
	Port int
}

func New(c Config) *Api {
	return &Api{
		port:    c.Port,
		service: c.App,
	}
}

func (api *Api) Run() error {
	e := echo.New()

	e.HTTPErrorHandler = HTTPErrorHandler
	e.Validator = NewValidator()

	e.Use(middleware.Recover())

	e.GET("/hello", api.hello)

	return e.Start(fmt.Sprintf(":%d", api.port))
}
