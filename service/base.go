// Copyright 2019 mlvzk
// This file is part of the piko library.
//
// The piko library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The piko library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the piko library. If not, see <http://www.gnu.org/licenses/>.

package service

import "io"

type Item struct {
	Meta             map[string]string
	DefaultName      string
	AvailableOptions map[string]([]string)
	DefaultOptions   map[string]string
}

type ServiceIterator interface {
	Next() ([]Item, error)
	HasEnded() bool
}

type Service interface {
	IsValidTarget(target string) bool
	FetchItems(target string) (ServiceIterator, error)
	Download(meta, options map[string]string) (io.Reader, error)
}

type Sized interface {
	Size() uint64
}
