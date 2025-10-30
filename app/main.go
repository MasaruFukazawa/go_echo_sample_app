package main

import (
	"net/http"
	"github.com/labstack/echo/v4"
)

func main() {

	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	e.GET("/hello", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello!!")
	})

	e.GET("/goodbey", func(c echo.Context) error {
		return c.String(http.StatusOK, "Goodbey!")
	})

	e.Logger.Fatal(e.Start(":8080"))
}
