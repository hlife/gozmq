// +build zmq_3_x

/*
  Copyright 2010-2012 Alec Thomas

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.
*/

package gozmq

/*
#cgo pkg-config: libzmq
#include <zmq.h>
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"errors"
	"unsafe"
)

const (
	RCVMORE = IntSocketOption(C.ZMQ_RCVMORE)
	SNDHWM  = IntSocketOption(C.ZMQ_SNDHWM)
	RCVHWM  = IntSocketOption(C.ZMQ_RCVHWM)

	// TODO Not documented in the man page...
	//LAST_ENDPOINT       = UInt64SocketOption(C.ZMQ_LAST_ENDPOINT)
	FAIL_UNROUTABLE     = BoolSocketOption(C.ZMQ_FAIL_UNROUTABLE)
	TCP_KEEPALIVE       = IntSocketOption(C.ZMQ_TCP_KEEPALIVE)
	TCP_KEEPALIVE_CNT   = IntSocketOption(C.ZMQ_TCP_KEEPALIVE_CNT)
	TCP_KEEPALIVE_IDLE  = IntSocketOption(C.ZMQ_TCP_KEEPALIVE_IDLE)
	TCP_KEEPALIVE_INTVL = IntSocketOption(C.ZMQ_TCP_KEEPALIVE_INTVL)
	TCP_ACCEPT_FILTER   = StringSocketOption(C.ZMQ_TCP_ACCEPT_FILTER)

	// Message options
	MORE = MessageOption(C.ZMQ_MORE)

	// Send/recv options
	DONTWAIT = SendRecvOption(C.ZMQ_DONTWAIT)

	// Deprecated aliases
	NOBLOCK = DONTWAIT
)

// Send a message to the socket.
// int zmq_send (void *s, zmq_msg_t *msg, int flags);
func (s *zmqSocket) Send(data []byte, flags SendRecvOption) error {
	var m C.zmq_msg_t
	// Copy data array into C-allocated buffer.
	size := C.size_t(len(data))

	if rc, err := C.zmq_msg_init_size(&m, size); rc != 0 {
		return casterr(err)
	}

	if size > 0 {
		// FIXME Ideally this wouldn't require a copy.
		C.memcpy(unsafe.Pointer(C.zmq_msg_data(&m)), unsafe.Pointer(&data[0]), size) // XXX I hope this works...(seems to)
	}

	if rc, err := C.zmq_sendmsg(s.s, &m, C.int(flags)); rc == -1 {
		// zmq_send did not take ownership, free message
		C.zmq_msg_close(&m)
		return casterr(err)
	}
	return nil
}

// Receive a message from the socket.
// int zmq_recv (void *s, zmq_msg_t *msg, int flags);
func (s *zmqSocket) Recv(flags SendRecvOption) (data []byte, err error) {
	// Allocate and initialise a new zmq_msg_t
	var m C.zmq_msg_t
	var rc C.int
	if rc, err = C.zmq_msg_init(&m); rc != 0 {
		err = casterr(err)
		return
	}
	defer C.zmq_msg_close(&m)
	// Receive into message
	if rc, err = C.zmq_recvmsg(s.s, &m, C.int(flags)); rc == -1 {
		err = casterr(err)
		return
	}
	err = nil
	// Copy message data into a byte array
	// FIXME Ideally this wouldn't require a copy.
	size := C.zmq_msg_size(&m)
	if size > 0 {
		data = make([]byte, int(size))
		C.memcpy(unsafe.Pointer(&data[0]), C.zmq_msg_data(&m), size)
	} else {
		data = nil
	}
	return
}

// Portability helper
func (s *zmqSocket) getRcvmore() (more bool, err error) {
	value, err := s.GetSockOptInt(RCVMORE)
	more = value != 0
	return
}

// run a zmq_proxy with in, out and capture sockets
func Proxy(in, out, capture Socket) error {
	if rc, err := C.zmq_proxy(in.apiSocket(), out.apiSocket(), capture.apiSocket()); rc != 0 {
		return casterr(err)
	}
	return errors.New("zmq_proxy() returned unexpectedly.")
}
