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

package service

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mysteriumnetwork/node/nat/traversal"
	"github.com/mysteriumnetwork/node/session"
)

func TestManager_StopNotPanic(t *testing.T) {
	m := Manager{}
	err := m.Stop()
	assert.NoError(t, err)
}

func TestManager_ProvideConfigNotFailOnEmptyConfig(t *testing.T) {
	m := Manager{vpnServiceConfigProvider: &mockConfigProvider{}, vpnServerPort: 1000}
	_, err := m.ProvideConfig([]byte(""), nil)
	assert.NoError(t, err)
}

func TestManager_ProvideConfigNotFailOnNilConfig(t *testing.T) {
	m := Manager{vpnServiceConfigProvider: &mockConfigProvider{}, vpnServerPort: 1000}
	_, err := m.ProvideConfig(nil, nil)
	assert.NoError(t, err)
}

type mockConfigProvider struct{}

func (cp *mockConfigProvider) ProvideConfig(sessionConfig json.RawMessage, traversalParams *traversal.Params) (*session.ConfigParams, error) {
	return &session.ConfigParams{SessionServiceConfig: traversalParams}, nil
}
