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

// Package adapters provides adapters to translate between networkservice.NetworkService{Server,Client}
package adapters

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"

	"github.com/networkservicemesh/networkservicemesh/controlplane/api/connection"
	"github.com/networkservicemesh/networkservicemesh/controlplane/api/networkservice"
)

type serverToClient struct {
	server networkservice.NetworkServiceServer
}

// NewServerToClient - returns a new networkservice.NetworkServiceClient that is a wrapper around server
func NewServerToClient(server networkservice.NetworkServiceServer) networkservice.NetworkServiceClient {
	return &serverToClient{server: server}
}

func (s *serverToClient) Request(ctx context.Context, in *networkservice.NetworkServiceRequest, opts ...grpc.CallOption) (*connection.Connection, error) {
	return s.server.Request(ctx, in)
}

func (s *serverToClient) Close(ctx context.Context, in *connection.Connection, opts ...grpc.CallOption) (*empty.Empty, error) {
	return s.server.Close(ctx, in)
}
