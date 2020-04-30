module caddyprom/example

go 1.14

require (
	github.com/caddyserver/caddy/v2 v2.0.0-rc.3
	github.com/caddyserver/jsonc-adapter v0.0.0-20200325004025-825ee096306c
	github.com/hairyhenderson/caddy-teapot-module v0.0.2
	github.com/hairyhenderson/caddyprom v0.0.0-00010101000000-000000000000
)

replace github.com/hairyhenderson/caddyprom => ../
