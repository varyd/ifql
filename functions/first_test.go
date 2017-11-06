package functions_test

import (
	"testing"

	"github.com/influxdata/ifql/functions"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/execute/executetest"
	"github.com/influxdata/ifql/query/querytest"
)

func TestFirstOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"first","kind":"first","spec":{"useRowTime":true}}`)
	op := &query.Operation{
		ID: "first",
		Spec: &functions.FirstOpSpec{
			UseRowTime: true,
		},
	}

	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestFirst_Process(t *testing.T) {
	testCases := []struct {
		name string
		data *executetest.Block
		want []execute.Row
	}{
		{
			name: "first",
			data: &executetest.Block{
				ColMeta: []execute.ColMeta{
					{Label: "time", Type: execute.TTime},
					{Label: "value", Type: execute.TFloat},
					{Label: "t1", Type: execute.TString, IsTag: true, IsCommon: true},
					{Label: "t2", Type: execute.TString, IsTag: true, IsCommon: false},
				},
				Data: [][]interface{}{
					{execute.Time(0), 0.0, "a", "y"},
					{execute.Time(10), 5.0, "a", "x"},
					{execute.Time(20), 9.0, "a", "y"},
					{execute.Time(30), 4.0, "a", "x"},
					{execute.Time(40), 6.0, "a", "y"},
					{execute.Time(50), 8.0, "a", "x"},
					{execute.Time(60), 1.0, "a", "y"},
					{execute.Time(70), 2.0, "a", "x"},
					{execute.Time(80), 3.0, "a", "y"},
					{execute.Time(90), 7.0, "a", "x"},
				},
			},
			want: []execute.Row{{
				Values: []interface{}{execute.Time(0), 0.0, "a", "y"},
			}},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			executetest.SelectorFuncTestHelper(
				t,
				new(functions.FirstSelector),
				tc.data,
				tc.want,
			)
		})
	}
}

func BenchmarkFirst(b *testing.B) {
	executetest.SelectorFuncBenchmarkHelper(b, new(functions.FirstSelector), NormalBlock)
}