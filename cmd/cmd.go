// Copyright 2016 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package cmd

import (
	"github.com/juju/cmd"
	"github.com/juju/errors"
	"github.com/juju/idmclient/ussologin"
	"github.com/juju/persistent-cookiejar"
	"gopkg.in/juju/environschema.v1/form"
	"gopkg.in/macaroon-bakery.v1/httpbakery"

	"github.com/juju/juju/juju/osenv"
)

// TODO (mattyw) Http needs to be HTTP
// HttpCommand can instantiate http bakery clients using a common cookie jar.
type HttpCommand struct {
	cmd.CommandBase

	cookiejar *cookiejar.Jar
}

// newTokenStore returns a FileTokenStore for storing the USSO oauth token
// TODO (mattyw) When this function lands in core this function should be
// removed and the calls should point to jujuclient.
func newTokenStore() *ussologin.FileTokenStore {
	return ussologin.NewFileTokenStore(osenv.JujuXDGDataHomePath("store-usso-token"))
}

// NewClient returns a new HTTP bakery client for commands.
func (s *HttpCommand) NewClient(ctx *cmd.Context) (*httpbakery.Client, error) {
	if s.cookiejar == nil {
		cookieFile := cookiejar.DefaultCookieFile()
		jar, err := cookiejar.New(&cookiejar.Options{
			Filename: cookieFile,
		})
		if err != nil {
			return nil, errors.Trace(err)
		}
		s.cookiejar = jar
	}
	client := httpbakery.NewClient()
	client.Jar = s.cookiejar
	client.VisitWebPage = ussologin.VisitWebPage(newIOFiller(ctx), client.Client, newTokenStore())
	return client, nil
}

// Close saves the persistent cookie jar used by the specified httpbakery.Client.
func (s *HttpCommand) Close() error {
	if s.cookiejar != nil {
		return s.cookiejar.Save()
	}
	return nil
}

func newIOFiller(ctx *cmd.Context) *form.IOFiller {
	return &form.IOFiller{
		In:  ctx.Stdin,
		Out: ctx.Stderr,
	}
}
