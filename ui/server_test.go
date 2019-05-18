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
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/html"
)

func Test_Server_ServesHTML(t *testing.T) {
	s := NewServer(55555)
	serverError := make(chan error)
	go func() {
		err := s.Serve()
		serverError <- err
	}()

	select {
	case <-time.After(time.Millisecond * 5):
	}

	resp, err := http.Get("http://:55555/")
	assert.Nil(t, err)

	defer resp.Body.Close()

	_, err = html.Parse(resp.Body)
	assert.Nil(t, err)

	s.Stop()
	assert.Nil(t, <-serverError)
}
