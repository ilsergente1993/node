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

// Package bandwidth allows us to keep track of the consumer side connection speed.
package bandwidth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/mysteriumnetwork/node/consumer"
	"github.com/mysteriumnetwork/node/core/connection"
)

func Test_ThroughputStringOutput(t *testing.T) {

}

func Test_bitCountDecimal(t *testing.T) {
	tests := []struct {
		name  string
		input int64
		want  string
	}{
		{
			name:  "tests",
			input: 1000,
			want:  "1.0 kbps",
		},
		{
			name:  "tests",
			input: 1500,
			want:  "1.5 kbps",
		},
		{
			name:  "tests",
			input: 100 * 0.5,
			want:  "50 bps",
		},
		{
			name:  "tests",
			input: 1000 * 1000,
			want:  "1.0 Mbps",
		},
		{
			name:  "tests",
			input: 1000 * 1000 * 1000,
			want:  "1.0 Gbps",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := bitCountDecimal(tt.input); got != tt.want {
				t.Errorf("bitCountDecimal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ConsumeSessionEvent_ResetsOnConnect(t *testing.T) {
	tracker := Tracker{
		previousTime: time.Now(),
		previous: consumer.SessionStatistics{
			BytesReceived: 1,
			BytesSent:     1,
		},
	}
	tracker.ConsumeSessionEvent(connection.SessionEvent{
		Status: connection.SessionCreatedStatus,
	})

	assert.True(t, tracker.previousTime.IsZero())
	assert.Zero(t, tracker.previous.BytesReceived)
	assert.Zero(t, tracker.previous.BytesSent)
}

func Test_ConsumeSessionEvent_ResetsOnDisconnect(t *testing.T) {
	tracker := Tracker{
		previousTime: time.Now(),
		previous: consumer.SessionStatistics{
			BytesReceived: 1,
			BytesSent:     1,
		},
	}
	tracker.ConsumeSessionEvent(connection.SessionEvent{
		Status: connection.SessionEndedStatus,
	})

	assert.True(t, tracker.previousTime.IsZero())
	assert.Zero(t, tracker.previous.BytesReceived)
	assert.Zero(t, tracker.previous.BytesSent)
}

func Test_ConsumeStatisticsEvent_CalculatesCorrectly(t *testing.T) {
	startTime := time.Now()
	var bytesTransfered uint64 = 10000
	tracker := Tracker{
		previousTime: startTime,
	}

	tracker.ConsumeStatisticsEvent(consumer.SessionStatistics{
		BytesReceived: bytesTransfered,
		BytesSent:     bytesTransfered,
	})

	assert.NotEqual(t, tracker.previousTime, startTime)
	speed := tracker.Get()
	expected := float64(bytesTransfered) / tracker.previousTime.Sub(startTime).Seconds() * bitsInByte
	assert.Equal(t, expected, speed.Down.BitsPerSecond)
	assert.Equal(t, expected, speed.Up.BitsPerSecond)
}

func Test_ConsumeStatisticsEvent_SkipsOnZero(t *testing.T) {
	tracker := Tracker{}
	input := consumer.SessionStatistics{
		BytesReceived: 1,
		BytesSent:     2,
	}
	tracker.ConsumeStatisticsEvent(input)
	assert.False(t, tracker.previousTime.IsZero())
	assert.Equal(t, input.BytesReceived, tracker.previous.BytesReceived)
	assert.Equal(t, input.BytesSent, tracker.previous.BytesSent)
	assert.Zero(t, tracker.Get().Down.BitsPerSecond)
}
