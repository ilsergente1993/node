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

package traversal

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	log "github.com/cihub/seelog"
	"github.com/mysteriumnetwork/node/core/port"
	"github.com/mysteriumnetwork/node/nat/event"
	"github.com/mysteriumnetwork/node/services"
	"github.com/pkg/errors"
	"golang.org/x/net/ipv4"
)

// StageName represents hole-punching stage of NAT traversal
const StageName = "hole_punching"
const prefix = "[NATPinger] "
const pingInterval = 200
const pingTimeout = 10000

var (
	errNATPunchAttemptStopped  = errors.New("NAT punch attempt stopped")
	errNATPunchAttemptTimedOut = errors.New("NAT punch attempt timed out")
)

// Pinger represents NAT pinger structure
type Pinger struct {
	pingTarget     chan *Params
	pingCancelled  chan struct{}
	stop           chan struct{}
	once           sync.Once
	natEventWaiter NatEventWaiter
	configParser   ConfigParser
	natProxy       natProxy
	portPool       PortSupplier
	consumerPort   int
	previousStage  string
	eventPublisher Publisher
}

// NatEventWaiter is responsible for waiting for nat events
type NatEventWaiter interface {
	WaitForEvent() event.Event
}

// ConfigParser is able to parse a config from given raw json
type ConfigParser interface {
	Parse(config json.RawMessage) (ip string, port int, serviceType services.ServiceType, err error)
}

// PortSupplier provides port needed to run a service on
type PortSupplier interface {
	Acquire() (port.Port, error)
}

// Publisher is responsible for publishing given events
type Publisher interface {
	Publish(topic string, data interface{})
}

// NewPingerFactory returns Pinger instance
func NewPingerFactory(waiter NatEventWaiter, parser ConfigParser, proxy natProxy, portPool PortSupplier, previousStage string, publisher Publisher) *Pinger {
	target := make(chan *Params)
	cancel := make(chan struct{})
	stop := make(chan struct{})
	return &Pinger{
		pingTarget:     target,
		pingCancelled:  cancel,
		stop:           stop,
		natEventWaiter: waiter,
		configParser:   parser,
		natProxy:       proxy,
		portPool:       portPool,
		previousStage:  previousStage,
		eventPublisher: publisher,
	}
}

type natProxy interface {
	handOff(serviceType services.ServiceType, conn *net.UDPConn)
	registerServicePort(serviceType services.ServiceType, port int)
	isAvailable(serviceType services.ServiceType) bool
}

// Params contains session parameters needed to NAT ping remote peer
type Params struct {
	RequestConfig json.RawMessage
	Port          int
	Cancel        chan struct{}
}

// Start starts NAT pinger and waits for PingTarget to ping
func (p *Pinger) Start() {
	log.Info(prefix, "Starting a NAT pinger")

	resultChannel := make(chan bool, 1)
	go func() { resultChannel <- p.waitForPreviousStageResult() }()
	select {
	case <-p.stop:
		return
	case previousStageSucceeded := <-resultChannel:
		if previousStageSucceeded {
			return
		}
	}

	for {
		select {
		case <-p.stop:
			return
		case pingParams := <-p.pingTarget:
			log.Info(prefix, "Pinging peer with: ", pingParams)

			IP, port, serviceType, err := p.configParser.Parse(pingParams.RequestConfig)
			if err != nil {
				log.Warn(prefix, errors.Wrap(err, fmt.Sprintf("unable to parse ping message: %v", pingParams)))
				continue
			}

			log.Infof("%sping target received: IP: %v, port: %v", prefix, IP, port)
			if !p.natProxy.isAvailable(serviceType) {
				log.Warn(prefix, serviceType, " NATProxy is not available for this transport protocol")
				continue
			}

			conn, err := p.getConnection(IP, port, pingParams.Port)
			if err != nil {
				log.Error(prefix, "failed to get connection: ", err)
				continue
			}

			go func() {
				err := p.ping(conn)
				if err != nil {
					log.Warn(prefix, "Error while pinging: ", err)
				}
			}()

			err = p.pingReceiver(conn, pingParams.Cancel)
			if err != nil {
				log.Error(prefix, "ping receiver error: ", err)
				continue
			}

			log.Info(prefix, "ping received, waiting for a new connection")

			go p.natProxy.handOff(serviceType, conn)
		}
	}
}

// Stop stops pinger loop
func (p *Pinger) Stop() {
	p.once.Do(func() { close(p.stop) })
}

// PingProvider pings provider determined by destination provided in sessionConfig
func (p *Pinger) PingProvider(ip string, port int, stop <-chan struct{}) error {
	log.Info(prefix, "NAT pinging to provider")

	conn, err := p.getConnection(ip, port, p.consumerPort)
	if err != nil {
		return errors.Wrap(err, "failed to get connection")
	}
	defer conn.Close()

	go func() {
		err := p.ping(conn)
		if err != nil {
			log.Warn(prefix, "Error while pinging: ", err)
		}
	}()

	time.Sleep(pingInterval * time.Millisecond)
	err = p.pingReceiver(conn, stop)
	if err != nil {
		return err
	}

	// wait for provider to setup NAT proxy connection
	time.Sleep(400 * time.Millisecond)

	return nil
}

func (p *Pinger) waitForPreviousStageResult() bool {
	for {
		event := p.natEventWaiter.WaitForEvent()
		if event.Stage == p.previousStage {
			return event.Successful
		}
	}
}

func (p *Pinger) ping(conn *net.UDPConn) error {
	n := 1

	for {
		select {
		case <-p.pingCancelled:
			return nil

		case <-time.After(pingInterval * time.Millisecond):
			log.Trace(prefix, "pinging.. ")
			// This is the essence of the TTL based udp punching.
			// We're slowly increasing the TTL so that the packet is held.
			// After a few attempts we're setting the value to 128 and assuming we're through.
			// We could stop sending ping to Consumer beyond 4 hops to prevent from possible Consumer's router's
			//  DOS block, but we plan, that Consumer at the same time will be Provider too in near future.
			if n > 4 {
				n = 128
			}

			err := p.sendPingRequest(conn, n)
			if err != nil {
				p.eventPublisher.Publish(event.Topic, event.BuildFailureEvent(StageName, err))
				return err
			}

			p.eventPublisher.Publish(event.Topic, event.BuildSuccessfulEvent(StageName))

			n++
		}
	}
}

func (p *Pinger) sendPingRequest(conn *net.UDPConn, ttl int) error {
	err := ipv4.NewConn(conn).SetTTL(ttl)
	if err != nil {
		return errors.Wrap(err, "pinger setting ttl failed")
	}

	_, err = conn.Write([]byte("continuously pinging to " + conn.RemoteAddr().String()))
	return errors.Wrap(err, "pinging request failed")
}

func (p *Pinger) getConnection(ip string, port int, pingerPort int) (*net.UDPConn, error) {
	udpAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		return nil, err
	}

	log.Info(prefix, "remote socket: ", udpAddr.String())

	conn, err := net.DialUDP("udp", &net.UDPAddr{Port: pingerPort}, udpAddr)
	if err != nil {
		return nil, err
	}

	log.Info(prefix, "local socket: ", conn.LocalAddr().String())

	return conn, nil
}

// PingTarget relays ping target address data
func (p *Pinger) PingTarget(target *Params) {
	select {
	case p.pingTarget <- target:
		return
	// do not block if ping target is not received
	case <-time.After(100 * time.Millisecond):
		return
	}
}

// BindConsumerPort binds NATPinger to source consumer port
func (p *Pinger) BindConsumerPort(port int) {
	p.consumerPort = port
}

// BindServicePort register service port to forward connection to
func (p *Pinger) BindServicePort(serviceType services.ServiceType, port int) {
	p.natProxy.registerServicePort(serviceType, port)
}

func (p *Pinger) pingReceiver(conn *net.UDPConn, stop <-chan struct{}) error {
	timeout := time.After(pingTimeout * time.Millisecond)
	for {
		select {
		case <-timeout:
			p.pingCancelled <- struct{}{}
			return errNATPunchAttemptTimedOut
		case <-stop:
			p.pingCancelled <- struct{}{}
			return errNATPunchAttemptStopped
		default:
		}

		var buf [512]byte
		n, err := conn.Read(buf[0:])
		if err != nil {
			log.Errorf(prefix, "Failed to read remote peer: %s cause: %s", conn.RemoteAddr().String(), err)
			time.Sleep(pingInterval * time.Millisecond)
			continue
		}
		fmt.Println("remote peer data received: ", string(buf[:n]))

		// send another couple of pings to remote side, because only now we have a pinghole
		// or wait for your pings to reach other end before closing pinger conn.
		select {
		case <-time.After(2 * pingInterval * time.Millisecond):
			p.pingCancelled <- struct{}{}
			return nil
		case <-stop:
			p.pingCancelled <- struct{}{}
			return errNATPunchAttemptStopped
		}
	}
}
