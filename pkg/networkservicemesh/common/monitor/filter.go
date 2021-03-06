// Copyright (c) 2020 Cisco Systems, Inc.
//
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package monitor

import "github.com/networkservicemesh/networkservicemesh/controlplane/api/connection"

type monitorFilter struct {
	selector *connection.MonitorScopeSelector
	connection.MonitorConnection_MonitorConnectionsServer
}

func newMonitorFilter(selector *connection.MonitorScopeSelector, srv connection.MonitorConnection_MonitorConnectionsServer) *monitorFilter {
	return &monitorFilter{
		selector: selector,
		MonitorConnection_MonitorConnectionsServer: srv,
	}
}

func (m *monitorFilter) Send(event *connection.ConnectionEvent) error {
	rv := &connection.ConnectionEvent{
		Type:        event.Type,
		Connections: connection.FilterMapOnManagerScopeSelector(event.GetConnections(), m.selector),
	}
	if rv.Type == connection.ConnectionEventType_INITIAL_STATE_TRANSFER || len(rv.GetConnections()) > 0 {
		return m.MonitorConnection_MonitorConnectionsServer.Send(rv)
	}
	return nil
}
