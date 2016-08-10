// Copyright 2014 The Serviced Authors.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package proxy

import (
	"github.com/control-center/serviced/utils"
	"github.com/zenoss/glog"

	"fmt"
	"io"
	"net"
	"strings"
	"syscall"
	"time"
)

// TCPMux is an implementation of tcp muxing RFC 1078.
type TCPMux struct {
	listener    net.Listener    // the connection this mux listens on
	connections chan net.Conn   // stream of accepted connections
	closing     chan chan error // shutdown noticiation
}

// NewTCPMux creates a new tcp mux with the given listener. If it succees, it
// is expected that this object is the owner of the listener and will close it
// when Close() is called on the TCPMux.
func NewTCPMux(listener net.Listener) (mux *TCPMux, err error) {
	if listener == nil {
		return nil, fmt.Errorf("listener can not be nil")
	}
	mux = &TCPMux{
		listener:    listener,
		connections: make(chan net.Conn),
		closing:     make(chan chan error),
	}
	go mux.loop()
	return mux, nil
}

func (mux *TCPMux) Close() {
	glog.V(5).Info("Close Called")
	close(mux.closing)
}

func (mux *TCPMux) acceptor(listener net.Listener, closing chan chan struct{}) {
	defer func() {
		close(mux.connections)
	}()
	for {
		conn, err := mux.listener.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "too many open files") {
				glog.Warningf("error accepting connections, retrying in 50 ms: %s", err)
				select {
				case <-closing:
					glog.V(5).Info("shutting down acceptor")
					return
				case <-time.After(time.Millisecond * 50):
					continue
				}
			}
			glog.Errorf("shutting down acceptor: %s", err)
			return
		}
		glog.V(5).Infof("accepted connection: %s", conn)
		select {
		case <-closing:
			glog.V(5).Info("shutting down acceptor")
			conn.Close()
			return
		case mux.connections <- conn:
		}
	}
}

func (mux *TCPMux) loop() {
	glog.V(5).Infof("entering TPCMux loop")
	closeAcceptor := make(chan chan struct{})
	go mux.acceptor(mux.listener, closeAcceptor)
	for {
		select {
		case errc := <-mux.closing:
			glog.V(5).Info("Closing mux")
			closeAcceptorAck := make(chan struct{})
			mux.listener.Close()
			closeAcceptor <- closeAcceptorAck
			errc <- nil
			return
		case conn, ok := <-mux.connections:
			if !ok {
				mux.connections = nil
				glog.V(6).Info("got nil conn, channel is closed")
				continue
			}
			glog.V(5).Info("handing mux connection")
			go mux.muxConnection(conn)
		}
	}
}

// muxConnection takes an inbound connection reads a line from it and
// then attempts to set up a connection to the service specified by the
// line. The service is specified in the form "IP:PORT\n". If the connection
// to the service is sucessful, all traffic continues to be proxied between
// two connections.
func (mux *TCPMux) muxConnection(conn net.Conn) {

	// make sure that we don't block indefinitely
	conn.SetReadDeadline(time.Now().Add(time.Second * 5))

	// Read in the mux header (a 6-byte representation of a TCP address)
	muxHeader := make([]byte, 6)
	if _, err := conn.Read(muxHeader); err != nil {
		glog.Errorf("could not read mux line: %s", err)
		conn.Close()
		return
	}
	address := utils.UnpackTCPAddressToString(muxHeader)

	// Restore the read deadline
	conn.SetReadDeadline(time.Time{})

	// Dial the requested address
	svc, err := net.Dial("tcp4", address)
	if err != nil {
		glog.Errorf("got %s => %s, could not dial to '%s' : %s", conn.LocalAddr(), conn.RemoteAddr(), address, err)
		conn.Close()
		return
	}

	// Wire up the incoming connection to the one we just dialed
	quit := make(chan bool)
	go ProxyLoop(conn, svc, quit)
}

func ProxyLoop(client net.Conn, backend net.Conn, quit chan bool) {
	event := make(chan int64)
	var broker = func(to, from net.Conn) {
		written, err := io.Copy(to, from)
		if err != nil {
			// If the socket we are writing to is shutdown with
			// SHUT_WR, forward it to the other end of the pipe:
			if err, ok := err.(*net.OpError); ok && err.Err == syscall.EPIPE {
				from.Close()
			}
		}
		to.Close()
		event <- written
	}

	go broker(client, backend)
	go broker(backend, client)

	var transferred int64 = 0
	for i := 0; i < 2; i++ {
		select {
		case written := <-event:
			transferred += written
		case <-quit:
			// Interrupt the two brokers and "join" them.
			client.Close()
			backend.Close()
			for ; i < 2; i++ {
				transferred += <-event
			}
			return
		}
	}
	//	glog.Infof("transferred %v bytes between %v", transferred, backendAddr)
	client.Close()
	backend.Close()
}
