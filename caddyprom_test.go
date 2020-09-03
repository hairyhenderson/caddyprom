package caddyprom_test

import (
	"fmt"
	"testing"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	//"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"

	. "github.com/hairyhenderson/caddyprom"
)

// This tests the different syntax options of the prometheus directive
func TestUnmarshalCaddyfileSyntax(t *testing.T) {

	testCaddyfiles := []string{
		`
			prometheus
		`,
		`
			prometheus 0.0.0.0:1337
		`,
		`
			prometheus {
				address 0.0.0.0:1337
			}
		`,
		`
			prometheus {
				address 0.0.0.0:1337
				path /otherpath
			}
		`,
	}

	// metrics struct with default values
	// TODO replace with actual metrics struct provisioned by caddy
	metrics := Metrics{
		Addr: "localhost:9180",
		Path: "/metrics",
	}

	for i, testCaddyfile := range testCaddyfiles {
		t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {

			dispenser := caddyfile.NewTestDispenser(testCaddyfile) // create a new test dispenser for every test string
			if err := metrics.UnmarshalCaddyfile(dispenser); err != nil { // check if there are errors while unmarshalling
				t.Error(err)
			}

		})
	}

}

// This tests, if the values in the caddyfile are properly assigned
// to the metrics sturct
func TestUnmarshalCaddyfileValues(t *testing.T) {

	testCaddyfiles := []string{
		`
			prometheus 0.0.0.0:1337 {
				path /otherpath
			}
		`,
		`
			prometheus {
				address 0.0.0.0:1337
				path /otherpath
			}
		`,
	}

	// metrics struct with default values
	// TODO replace with actual metrics struct provisioned by caddy
	metrics := Metrics{
		Addr: "localhost:9180",
		Path: "/metrics",
	}

	for i, testCaddyfile := range testCaddyfiles {
		t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {

			dispenser := caddyfile.NewTestDispenser(testCaddyfile) // create a new test dispenser for every test string
			metrics.UnmarshalCaddyfile(dispenser) // unmarshal errors are tested in TestUnmarshalCaddyfileSyntax
			if metrics.Addr != "0.0.0.0:1337" {
				t.Error("wrong metrics address")
			}
			if metrics.Path != "/otherpath" {
				t.Error("wrong metrics path")
			}

		})
	}

}

// This tests, if setting the address inline and as subdirective
// actually fails as it should
func TestUnmarshalCaddyfileExclusiveInlineAddress(t *testing.T) {

	dispenser := caddyfile.NewTestDispenser(`
		prometheus 0.0.0.0:1337 {
			address 0.0.0.0:1337
		}
	`)

	// metrics struct with default values
	// TODO replace with actual metrics struct provisioned by caddy
	metrics := Metrics{
		Addr: "localhost:9180",
		Path: "/metrics",
	}

	if err := metrics.UnmarshalCaddyfile(dispenser); err == nil {
		t.Error("address can be set inline and as subdirective")
	}

}