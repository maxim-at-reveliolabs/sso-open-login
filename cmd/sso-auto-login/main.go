package main

import (
	"os"

	"sso-auto-login/internal/app"
)

func main() {
	os.Exit(app.Main(os.Args[1:]))
}
