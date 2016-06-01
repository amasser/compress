// Copyright 2015, Joe Tsai. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.md file.

package bzip2

import (
	"bytes"
	"io"
	"testing"

	"github.com/dsnet/compress/internal/testutil"
)

var testdata = []struct {
	name string
	data []byte
}{
	{"Nil", nil},
	{"Binary", testutil.MustLoadFile("../testdata/binary.bin")},
	{"Digits", testutil.MustLoadFile("../testdata/digits.txt")},
	{"Huffman", testutil.MustLoadFile("../testdata/huffman.txt")},
	{"Random", testutil.MustLoadFile("../testdata/random.bin")},
	{"Repeats", testutil.MustLoadFile("../testdata/repeats.bin")},
	{"Twain", testutil.MustLoadFile("../testdata/twain.txt")},
	{"Zeros", testutil.MustLoadFile("../testdata/zeros.bin")},
}

var levels = []struct {
	name  string
	level int
}{
	{"Speed", BestSpeed},
	{"Default", DefaultCompression},
	{"Compression", BestCompression},
}

var sizes = []struct {
	name string
	size int
}{
	{"1e4", 1e4},
	{"1e5", 1e5},
	{"1e6", 1e6},
}

func TestRoundTrip(t *testing.T) {
	for i, v := range testdata {
		var buf1, buf2 bytes.Buffer

		// Compress the input.
		wr, err := NewWriter(&buf1, nil)
		if err != nil {
			t.Errorf("test %d, NewWriter() = (_, %v), want (_, nil)", i, err)
		}
		n, err := io.Copy(wr, bytes.NewReader(v.data))
		if n != int64(len(v.data)) || err != nil {
			t.Errorf("test %d, Copy() = (%d, %v), want (%d, nil)", i, n, err, len(v.data))
		}
		if err := wr.Close(); err != nil {
			t.Errorf("test %d, Close() = %v, want nil", i, err)
		}

		// Write a canary byte to ensure this does not get read.
		buf1.WriteByte(0x7a)

		// Decompress the output.
		rd, err := NewReader(&buf1, nil)
		if err != nil {
			t.Errorf("test %d, NewReader() = (_, %v), want (_, nil)", i, err)
		}
		n, err = io.Copy(&buf2, rd)
		if n != int64(len(v.data)) || err != nil {
			t.Errorf("test %d, Copy() = (%d, %v), want (%d, nil)", i, n, err, len(v.data))
		}
		if err := rd.Close(); err != nil {
			t.Errorf("test %d, Close() = %v, want nil", i, err)
		}
		if !bytes.Equal(buf2.Bytes(), v.data) {
			t.Errorf("test %d, output data mismatch", i)
		}

		// Read back the canary byte.
		if v, _ := buf1.ReadByte(); v != 0x7a {
			t.Errorf("Read consumed more data than necessary")
		}
	}
}

func runBenchmarks(b *testing.B, f func(b *testing.B, buf []byte, lvl int)) {
	for _, td := range testdata {
		if len(td.data) == 0 {
			continue
		}
		if testing.Short() && !(td.name == "Twain" || td.name == "Digits") {
			continue
		}
		for _, tl := range levels {
			for _, ts := range sizes {
				buf := testutil.ResizeData(td.data, ts.size)
				b.Run(td.name+"/"+tl.name+"/"+ts.name, func(b *testing.B) {
					f(b, buf, tl.level)
				})
			}
		}
	}
}
