/*
 * Copyright (C) 2017 The "MysteriumNetwork/node" Authors.
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

package dialog

import (
	"testing"

	"github.com/mysteriumnetwork/node/communication"
	"github.com/mysteriumnetwork/node/communication/nats"
	"github.com/mysteriumnetwork/node/communication/nats/discovery"
	"github.com/mysteriumnetwork/node/identity"
	"github.com/mysteriumnetwork/node/market"
	"github.com/stretchr/testify/assert"
)

var _ communication.DialogEstablisher = &dialogEstablisher{}

func TestDialogEstablisher_Factory(t *testing.T) {
	id := identity.FromAddress("123456")
	signer := &identity.SignerFake{}

	establisher := NewDialogEstablisher(id, signer)
	assert.NotNil(t, establisher)
	assert.Equal(t, id, establisher.ID)
	assert.Equal(t, signer, establisher.Signer)
}

func TestDialogEstablisher_EstablishDialog(t *testing.T) {
	myID := identity.FromAddress("0x6B21b441D0D2Fa1d86407977A3a5C6eD90Ff1A62")
	peerID := identity.FromAddress("0x0d1a35e53b7f3478d00B7C23838C0D48b2a81017")

	connection := nats.StartConnectionMock()
	connection.MockResponse(
		"peer-topic.dialog-create",
		[]byte(`{
			"payload": {"reason":200,"reasonMessage":"OK"},
            "signature": "iaV65n3kEve9+EzwWVi65qJFrb4FQZwq4yWdVH++abts3mW/xqKHpPKro7kX/liFRZgV5RHQMjE+TzPPdeJfewA="
		}`),
	)
	defer connection.Close()

	signer := &identity.SignerFake{}
	establisher := mockEstablisher(myID, connection, signer)

	dialogInstance, err := establisher.EstablishDialog(peerID, market.Contact{})
	defer dialogInstance.Close()
	assert.NoError(t, err)
	assert.NotNil(t, dialogInstance)

	dialog, ok := dialogInstance.(*dialog)
	assert.True(t, ok)

	expectedCodec := NewCodecSecured(communication.NewCodecJSON(), signer, identity.NewVerifierIdentity(peerID))
	assert.Equal(
		t,
		nats.NewSender(connection, expectedCodec, "peer-topic."+myID.Address),
		dialog.Sender,
	)
	assert.Equal(
		t,
		nats.NewReceiver(connection, expectedCodec, "peer-topic."+myID.Address),
		dialog.Receiver,
	)
}

func TestDialogEstablisher_CreateDialogWhenResponseHijacked(t *testing.T) {
	myID := identity.FromAddress("0x6B21b441D0D2Fa1d86407977A3a5C6eD90Ff1A62")
	peerID := identity.FromAddress("0x0d1a35e53b7f3478d00B7C23838C0D48b2a81017")

	connection := nats.StartConnectionMock()
	connection.MockResponse(
		"peer-topic.dialog-create",
		[]byte(`{
			"payload": {"reason":200,"reasonMessage":"OK"},
			"signature": "2Rg9KabJXdYEsMLynoeZ8+4cWjauHuZq/ydIE0NuNl1psu+AVz/8fHaqdG81CUgf2dNQHjciOVPagEb+X6//sgA="
		}`),
	)
	defer connection.Close()

	establisher := mockEstablisher(myID, connection, &identity.SignerFake{})

	dialogInstance, err := establisher.EstablishDialog(peerID, market.Contact{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dialog creation error. failed to unpack response 'peer-topic.dialog-create'. invalid message signature ")
	assert.Nil(t, dialogInstance)
}

func mockEstablisher(ID identity.Identity, connection nats.Connection, signer identity.Signer) *dialogEstablisher {
	peerTopic := "peer-topic"

	return &dialogEstablisher{
		ID:     ID,
		Signer: signer,
		peerAddressFactory: func(contact market.Contact) (*discovery.AddressNATS, error) {
			return discovery.NewAddressWithConnection(connection, peerTopic), nil
		},
	}
}
