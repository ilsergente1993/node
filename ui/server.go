/*
 * Copyright (C) 2019 The "MysteriumNetwork/node" Authors.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package ui

import (
	"context"
	"fmt"
	"net/http"
	"time"

	log "github.com/cihub/seelog"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	godvpnweb "github.com/mysteriumnetwork/go-dvpn-web"
	"github.com/pkg/errors"
)

const logPrefix = "[dvpn-web-server] "

// Server represents our web UI server
type Server struct {
	srv *http.Server
}

// NewServer creates a new instance of the server for the given port
func NewServer(port int) *Server {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(cors.Default())
	r.StaticFS("/", godvpnweb.Assets)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%v", port),
		Handler: r,
	}

	return &Server{
		srv: srv,
	}
}

// Serve starts the server
func (s *Server) Serve() error {
	log.Info(logPrefix, "server starting on: ", s.srv.Addr)
	err := s.srv.ListenAndServe()
	if err != http.ErrServerClosed {
		return errors.Wrap(err, "dvpn web server crashed")
	}
	return nil
}

// Stop stops the server
func (s *Server) Stop() {
	// give the server a few seconds to shut down properly in case a request is waiting somewhere
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := s.srv.Shutdown(ctx)
	if err != nil {
		log.Error(logPrefix, "server exit error: ", err)
	}
	log.Info(logPrefix, "server exited")
}
