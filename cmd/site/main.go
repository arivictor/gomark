package main

import (
	"log"

	"github.com/arivictor/gomark"
)

func main() {
	app := gomark.App{
		ContentDir: "cmd/site/content",
		Mode:       gomark.PreRender,
	}
	if err := app.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
