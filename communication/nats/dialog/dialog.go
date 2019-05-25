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
	"github.com/mysteriumnetwork/node/communication"
	"github.com/mysteriumnetwork/node/communication/nats/discovery"
	"github.com/mysteriumnetwork/node/identity"
)

type dialog struct {
	communication.Sender
	communication.Receiver
	peerID      identity.Identity
	peerAddress *discovery.AddressNATS
}

func (dialog *dialog) Close() error {
	if dialog.peerAddress != nil {
		dialog.peerAddress.Disconnect()
	}
	return nil
}

func (dialog *dialog) PeerID() identity.Identity {
	return dialog.peerID
}
