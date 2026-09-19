package main

import (
	"github.com/jekyulll/url_shortener/app"
)

func main() {
	a := app.New()
	if err := a.Init("./config/config.yaml"); err != nil {
		panic(err)
	}
	a.Run()
}
