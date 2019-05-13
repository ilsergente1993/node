/*
 * Copyright (C) 2018 The "MysteriumNetwork/node" Authors.
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

package noop

import (
	"encoding/json"
	"errors"
	"sync"

	log "github.com/cihub/seelog"

	"github.com/mysteriumnetwork/node/core/location"
	"github.com/mysteriumnetwork/node/identity"
	"github.com/mysteriumnetwork/node/market"
	"github.com/mysteriumnetwork/node/money"
	"github.com/mysteriumnetwork/node/nat/traversal"
	"github.com/mysteriumnetwork/node/session"
)

const logPrefix = "[service-noop] "

// ErrAlreadyStarted is the error we return when the start is called multiple times
var ErrAlreadyStarted = errors.New("service already started")

// NewManager creates new instance of Noop service
func NewManager() *Manager {
	return &Manager{}
}

// Manager represents entrypoint for Noop service
type Manager struct {
	process sync.WaitGroup
}

// ProvideConfig provides the session configuration
func (manager *Manager) ProvideConfig(sessionConfig json.RawMessage, traversalParams *traversal.Params) (*session.ConfigParams, error) {
	return &session.ConfigParams{TraversalParams: traversalParams}, nil
}

// Serve starts service - does block
func (manager *Manager) Serve(providerID identity.Identity) error {
	manager.process.Add(1)
	log.Info(logPrefix, "Noop service started successfully")
	manager.process.Wait()
	return nil
}

// Stop stops service
func (manager *Manager) Stop() error {
	manager.process.Done()
	log.Info(logPrefix, "Noop service stopped")
	return nil
}

// GetProposal returns the proposal for NOOP service for given country
func GetProposal(location location.Location) market.ServiceProposal {
	return market.ServiceProposal{
		ServiceType: ServiceType,
		ServiceDefinition: ServiceDefinition{
			Location: market.Location{
				Continent: location.Continent,
				Country:   location.Country,
				City:      location.City,

				ASN:      location.ASN,
				ISP:      location.ISP,
				NodeType: location.NodeType,
			},
		},
		PaymentMethodType: PaymentMethodNoop,
		PaymentMethod: PaymentNoop{
			Price: money.NewMoney(0, money.CurrencyMyst),
		},
	}
}
