// Copyright 2023 Cisco Systems, Inc. and its affiliates
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
//
// SPDX-License-Identifier: Apache-2.0

package testutils

import (
    "bufio"
    "fmt"
    "github.com/google/uuid"
    "io/fs"
)

// UUIDPool hold a list of uuid and make sure same uuid is returned each time Pop is called
type UUIDPool struct {
    Pool    []uuid.UUID
    Current int
}

// Pop return a uuid and increase index by one. This function returns error when index out of bound
func (p *UUIDPool) Pop() (uuid.UUID, error) {
    defer func() {p.Current++}()
    if p.Current >= len(p.Pool) {
        return uuid.Nil, fmt.Errorf("UUID pool exhausted")
    }
    return p.Pool[p.Current], nil
}

func (p *UUIDPool) PopOrNew() uuid.UUID {
    id, e := p.Pop()
    if e != nil {
        return uuid.New()
    }
    return id
}

func NewUUIDPool(fsys fs.FS, src string) (*UUIDPool, error) {
    f, e := fsys.Open(src)
    if e != nil {
        return nil, e
    }
    defer func() {_ = f.Close()}()

    scanner := bufio.NewScanner(f)
    scanner.Split(bufio.ScanLines)
    pool := make([]uuid.UUID, 0, 32)
    for scanner.Scan() {
        id, e := uuid.Parse(scanner.Text())
        if e == nil {
            pool = append(pool, id)
        }
    }
    if len(pool) == 0 {
        return nil, fmt.Errorf("unable to load UUIDs")
    }
    return &UUIDPool{
        Pool:    pool,
    }, nil
}
