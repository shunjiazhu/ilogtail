// Copyright 2022 iLogtail Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package helper

import (
	"errors"
	"io"
	"net"

	knet "k8s.io/apimachinery/pkg/util/net"
)

func GetFreePort() (port int, err error) {
	listener, err := net.Listen("tcp", ":0") //nolint:gosec
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

func IsErrorEOF(err error) bool {
	return errors.Is(err, io.EOF) || knet.IsProbableEOF(err)
}
