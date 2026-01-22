// Copyright (c) 2023-2026 Matteo Pacini
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBoolFromlInt_UnmarshalJSON(t *testing.T) {

	t.Run("unmarshals 1 as true", func(t *testing.T) {
		var b BoolFromlInt
		err := b.UnmarshalJSON([]byte("1"))
		assert.NoError(t, err)
		assert.True(t, bool(b))
	})

	t.Run("unmarshals true as true", func(t *testing.T) {
		var b BoolFromlInt
		err := b.UnmarshalJSON([]byte("true"))
		assert.NoError(t, err)
		assert.True(t, bool(b))
	})

	t.Run("unmarshals 0 as false", func(t *testing.T) {
		var b BoolFromlInt
		err := b.UnmarshalJSON([]byte("0"))
		assert.NoError(t, err)
		assert.False(t, bool(b))
	})

	t.Run("unmarshals false as false", func(t *testing.T) {
		var b BoolFromlInt
		err := b.UnmarshalJSON([]byte("false"))
		assert.NoError(t, err)
		assert.False(t, bool(b))
	})

	t.Run("unmarshals null as false", func(t *testing.T) {
		var b BoolFromlInt
		err := b.UnmarshalJSON([]byte("null"))
		assert.NoError(t, err)
		assert.False(t, bool(b))
	})

	t.Run("returns error for invalid input", func(t *testing.T) {
		var b BoolFromlInt
		err := b.UnmarshalJSON([]byte("invalid"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "boolean from int unmarshal error")
	})

	t.Run("returns error for number other than 0 or 1", func(t *testing.T) {
		var b BoolFromlInt
		err := b.UnmarshalJSON([]byte("2"))
		assert.Error(t, err)
	})
}
