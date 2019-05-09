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

package metrics

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockEventsTransport struct {
	sentEvent    Event
	mockResponse error
}

func buildMockEventsTransport(mockResponse error) *mockEventsTransport {
	return &mockEventsTransport{mockResponse: mockResponse}
}

func (transport *mockEventsTransport) SendEvent(event Event) error {
	transport.sentEvent = event
	return transport.mockResponse
}

func mockGatewayLoader() []map[string]string {
	return nil
}

func TestSender_SendStartupEvent_SendsToTransport(t *testing.T) {
	mockTransport := buildMockEventsTransport(nil)
	sender := &Sender{Transport: mockTransport, AppVersion: "test version", GatewayLoader: mockGatewayLoader}

	err := sender.SendStartupEvent()
	assert.NoError(t, err)

	sentEvent := mockTransport.sentEvent

	assert.Equal(t, "startup", sentEvent.EventName)
	assert.Equal(t, appInfo{Name: "myst", Version: "test version"}, sentEvent.Application)
	assert.NotZero(t, sentEvent.CreatedAt)
}

func TestSender_SendStartupEvent_ReturnsTransportErrors(t *testing.T) {
	mockTransport := buildMockEventsTransport(nil)
	mockTransport.mockResponse = errors.New("mock error")
	sender := &Sender{Transport: mockTransport, AppVersion: "test version", GatewayLoader: mockGatewayLoader}

	err := sender.SendStartupEvent()
	assert.Error(t, err)
}

func TestSender_SendNATMappingSuccessEvent_SendsToTransport(t *testing.T) {
	mockTransport := buildMockEventsTransport(nil)
	sender := &Sender{Transport: mockTransport, AppVersion: "test version", GatewayLoader: mockGatewayLoader}

	err := sender.SendNATMappingSuccessEvent("port_mapping")
	assert.NoError(t, err)

	sentEvent := mockTransport.sentEvent
	assert.Equal(t, "nat_mapping", sentEvent.EventName)
	assert.Equal(t, appInfo{Name: "myst", Version: "test version"}, sentEvent.Application)
	assert.NotZero(t, sentEvent.CreatedAt)
	assert.Equal(t, natMappingContext{Successful: true, Stage: "port_mapping"}, sentEvent.Context)
}

func TestSender_SendNATMappingSuccessEvent_ReturnsTransportErrors(t *testing.T) {
	mockTransport := buildMockEventsTransport(nil)
	mockTransport.mockResponse = errors.New("mock error")
	sender := &Sender{Transport: mockTransport, AppVersion: "test version", GatewayLoader: mockGatewayLoader}

	err := sender.SendNATMappingSuccessEvent("port_mapping")
	assert.Error(t, err)
}

func TestSender_SendNATMappingFailEvent_SendsToTransport(t *testing.T) {
	mockTransport := buildMockEventsTransport(nil)

	mockGateways := []map[string]string{
		map[string]string{
			"test": "test",
		},
	}

	sender := &Sender{Transport: mockTransport, AppVersion: "test version", GatewayLoader: func() []map[string]string {
		return mockGateways
	}}

	mockError := errors.New("mock nat mapping error")
	err := sender.SendNATMappingFailEvent("hole_punching", mockError)
	assert.NoError(t, err)

	sentEvent := mockTransport.sentEvent
	assert.Equal(t, "nat_mapping", sentEvent.EventName)
	assert.Equal(t, appInfo{Name: "myst", Version: "test version"}, sentEvent.Application)
	assert.NotZero(t, sentEvent.CreatedAt)
	c := sentEvent.Context.(natMappingContext)
	assert.False(t, c.Successful)
	assert.Equal(t, "mock nat mapping error", *c.ErrorMessage)
	assert.Equal(t, "hole_punching", c.Stage)
	assert.Equal(t, mockGateways, c.Gateways)
}

func TestSender_SendNATMappingFailEvent_ReturnsTransportErrors(t *testing.T) {
	mockTransport := buildMockEventsTransport(nil)
	mockTransport.mockResponse = errors.New("mock error")
	sender := &Sender{Transport: mockTransport, AppVersion: "test version", GatewayLoader: mockGatewayLoader}

	err := sender.SendNATMappingFailEvent("hole_punching", errors.New("mock nat mapping error"))
	assert.Error(t, err)
}
