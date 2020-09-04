package caddyprom

import (
	"fmt"
	"testing"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/stretchr/testify/assert"
)

func TestUnmarshalCaddyfile(t *testing.T) {
	metrics := &Metrics{}
	dispenser := caddyfile.NewTestDispenser(
		`prometheus 0.0.0.0:1337 {
			address 0.0.0.0:1337
		}`)
	assert.Error(t, metrics.UnmarshalCaddyfile(dispenser))
	dispenser = caddyfile.NewTestDispenser(
		`prometheus {
			bogus
		}`)
	assert.Error(t, metrics.UnmarshalCaddyfile(dispenser))

	testdata := []struct {
		caddyfile string
		expected  *Metrics
	}{
		{
			`prometheus`,
			&Metrics{Addr: "localhost:9180", Path: "/metrics"},
		},
		{
			`prometheus 0.0.0.0:1337`,
			&Metrics{Addr: "0.0.0.0:1337", Path: "/metrics"},
		},
		{
			`prometheus {
				address 0.0.0.0:1337
			}`,
			&Metrics{Addr: "0.0.0.0:1337", Path: "/metrics"},
		},
		{
			`prometheus 0.0.0.0:1337 {
				path /otherpath
			}`,
			&Metrics{Addr: "0.0.0.0:1337", Path: "/otherpath"},
		},
		{
			`prometheus {
				address 0.0.0.0:1337
				path /otherpath
			}`,
			&Metrics{Addr: "0.0.0.0:1337", Path: "/otherpath"},
		},
	}

	for i, d := range testdata {
		t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
			metrics := &Metrics{Addr: "localhost:9180", Path: "/metrics"}
			dispenser := caddyfile.NewTestDispenser(d.caddyfile)

			assert.NoError(t, metrics.UnmarshalCaddyfile(dispenser))
			assert.EqualValues(t, d.expected, metrics)
		})
	}
}
