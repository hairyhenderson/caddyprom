package main

import (
	caddycmd "github.com/caddyserver/caddy/v2/cmd"
	_ "github.com/caddyserver/caddy/v2/modules/standard"
	_ "github.com/caddyserver/jsonc-adapter"

	_ "github.com/hairyhenderson/caddy-teapot-module"
	_ "github.com/hairyhenderson/caddyprom"
)

func main() {
	caddycmd.Main()
}
