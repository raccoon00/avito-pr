package main

import (
	"log"
	"os"

	"github.com/raccoon00/avito-pr/internal/app"
)

func main() {
	log.SetOutput(os.Stdout)
	app.Run()
}
