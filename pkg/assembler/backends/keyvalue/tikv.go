//
// Copyright 2023 The GUAC Authors.
//
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

//go:build !(386 || arm || mips || darwin)

package keyvalue

import "github.com/guacsec/guac/pkg/assembler/kv/tikv"

func init() {
	// TiKV does not support 32 bit. Also darwin required CGO and cross compile
	// using xcode...
	tikvGS = tikv.GetStore
}
