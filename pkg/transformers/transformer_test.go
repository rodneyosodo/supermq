// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package transformers_test

import (
	"testing"
	"time"

	"github.com/absmach/magistrala/pkg/transformers"
)

var now = time.Now()

func TestInt64ToUnixNano(t *testing.T) {
	cases := []struct {
		desc string
		time int64
		want int64
	}{
		{
			desc: "empty",
			time: 0,
			want: 0,
		},
		{
			desc: "unix",
			time: now.Unix(),
			want: now.Unix() * int64(time.Second),
		},
		{
			desc: "unix milli",
			time: now.UnixMilli(),
			want: now.UnixMilli() * int64(time.Millisecond),
		},
		{
			desc: "unix micro",
			time: now.UnixMicro(),
			want: now.UnixMicro() * int64(time.Microsecond),
		},
		{
			desc: "unix nano",
			time: now.UnixNano(),
			want: now.UnixNano(),
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			got := transformers.ToUnixNano(c.time)
			if got != c.want {
				t.Errorf("ToUnixNano(%d) = %d; want %d", c.time, got, c.want)
			}
			t.Logf("ToUnixNano(%d) = %d; want %d", c.time, got, c.want)
		})
	}
}

func TestFloat64ToUnixNano(t *testing.T) {
	cases := []struct {
		desc string
		time float64
		want float64
	}{
		{
			desc: "empty",
			time: 0,
			want: 0,
		},
		{
			desc: "unix",
			time: float64(now.Unix()),
			want: float64(now.Unix() * int64(time.Second)),
		},
		{
			desc: "unix milli",
			time: float64(now.UnixMilli()),
			want: float64(now.UnixMilli() * int64(time.Millisecond)),
		},
		{
			desc: "unix micro",
			time: float64(now.UnixMicro()),
			want: float64(now.UnixMicro() * int64(time.Microsecond)),
		},
		{
			desc: "unix nano",
			time: float64(now.UnixNano()),
			want: float64(now.UnixNano()),
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			got := transformers.ToUnixNano(c.time)
			if got != c.want {
				t.Errorf("ToUnixNano(%f) = %f; want %f", c.time, got, c.want)
			}
			t.Logf("ToUnixNano(%f) = %f; want %f", c.time, got, c.want)
		})
	}
}

func BenchmarkToUnixNano(b *testing.B) {
	for i := 0; i < b.N; i++ {
		transformers.ToUnixNano(now.Unix())
		transformers.ToUnixNano(now.UnixMilli())
		transformers.ToUnixNano(now.UnixMicro())
		transformers.ToUnixNano(now.UnixNano())
	}
}
