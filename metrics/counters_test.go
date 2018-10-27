package metrics

import (
	"reflect"
	"testing"
)

func getUint64(i uint64) *uint64 {
	return &i
}

func TestCounters_Count(t *testing.T) {
	type fields struct {
		m map[string]*uint64
	}
	type args struct {
		s     string
		delta uint64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   uint64
	}{
		{
			"Empty map",
			fields{
				m: map[string]*uint64{},
			},
			args{
				s:     "foo",
				delta: 1,
			},
			1,
		},
		{
			"Map with one key",
			fields{
				m: map[string]*uint64{"foo": new(uint64)},
			},
			args{
				s:     "foo",
				delta: 1,
			},
			1,
		},
		{
			"Map with 1 counter, value 10",
			fields{
				m: map[string]*uint64{"foo": getUint64(10)},
			},
			args{
				s:     "foo",
				delta: 1,
			},
			11,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCounters()
			c.m = tt.fields.m

			c.Count(tt.args.s, tt.args.delta)
			if c.Get(tt.args.s) != tt.want {
				t.Errorf("Unexpected value, want: %d, got %d", tt.want, c.Get(tt.args.s))
			}
		})
	}
}

func TestCounters_Collect(t *testing.T) {
	type fields struct {
		m map[string]*uint64
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]uint64
	}{
		{
			"Map with one key",
			fields{
				m: map[string]*uint64{
					"foo": new(uint64),
					"bar": getUint64(10),
				},
			},
			map[string]uint64{"foo": 0, "bar": 10},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCounters()
			c.m = tt.fields.m

			if got := c.Collect(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Counters.Collect() = %v, want %v", got, tt.want)
			}
		})
	}
}
