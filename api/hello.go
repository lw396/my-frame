package api

import "github.com/labstack/echo/v4"

func (api *Api) hello(c echo.Context) error {
	return c.JSON(200, "Hello World")
}
