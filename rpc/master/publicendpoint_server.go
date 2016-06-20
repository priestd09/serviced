// Copyright 2016 The Serviced Authors.
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

package master

import (
	"github.com/control-center/serviced/domain/servicedefinition"
)

// Adds a port public endpoint to a service.
func (s *Server) AddPublicEndpointPort(request *PublicEndpointPortRequest, reply *servicedefinition.Port) error {
	port, err := s.f.AddPublicEndpointPort(s.context(), request.Serviceid, request.EndpointName, request.PortAddr,
		request.UseTLS, request.Protocol, request.IsEnabled, request.Restart)
	if err != nil {
		return err
	}
	*reply = *port
	return err
}

// Remove a port public endpoint from a service.
func (s *Server) RemovePublicEndpointPort(request *PublicEndpointPortRequest, _ *struct{}) error {
	return s.f.RemovePublicEndpointPort(s.context(), request.Serviceid, request.EndpointName, request.PortAddr)
}