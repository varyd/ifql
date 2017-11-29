package functions_test

import (
	"testing"

	"github.com/influxdata/ifql/expression"
	"github.com/influxdata/ifql/functions"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/execute/executetest"
	"github.com/influxdata/ifql/query/plan"
	"github.com/influxdata/ifql/query/plan/plantest"
	"github.com/influxdata/ifql/query/querytest"
)

func TestFilterOperation_Marshaling(t *testing.T) {
	data := []byte(`{
		"id":"filter",
		"kind":"filter",
		"spec":{
			"expression":{
				"root":{
					"type":"binary",
					"operator": "!=",
					"left":{
						"type":"reference",
						"name":"_measurement",
						"kind":"tag"
					},
					"right":{
						"type":"stringLiteral",
						"value":"mem"
					}
				}
			}
		}
	}`)
	op := &query.Operation{
		ID: "filter",
		Spec: &functions.FilterOpSpec{
			Expression: expression.Expression{
				Root: &expression.BinaryNode{
					Operator: expression.NotEqualOperator,
					Left: &expression.ReferenceNode{
						Name: "_measurement",
						Kind: "tag",
					},
					Right: &expression.StringLiteralNode{
						Value: "mem",
					},
				},
			},
		},
	}
	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestFilter_Process(t *testing.T) {
	testCases := []struct {
		name string
		spec *functions.FilterProcedureSpec
		data []execute.Block
		want []*executetest.Block
	}{
		{
			name: "$>5",
			spec: &functions.FilterProcedureSpec{
				Expression: expression.Expression{
					Root: &expression.BinaryNode{
						Operator: expression.GreaterThanOperator,
						Left: &expression.ReferenceNode{
							Name: "$",
						},
						Right: &expression.FloatLiteralNode{
							Value: 5,
						},
					},
				},
			},
			data: []execute.Block{&executetest.Block{
				Bnds: execute.Bounds{
					Start: 1,
					Stop:  3,
				},
				ColMeta: []execute.ColMeta{
					{Label: "time", Type: execute.TTime},
					{Label: "value", Type: execute.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0},
					{execute.Time(2), 6.0},
				},
			}},
			want: []*executetest.Block{{
				Bnds: execute.Bounds{
					Start: 1,
					Stop:  3,
				},
				ColMeta: []execute.ColMeta{
					{Label: "time", Type: execute.TTime},
					{Label: "value", Type: execute.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(2), 6.0},
				},
			}},
		},
		{
			name: "$>5 multiple blocks",
			spec: &functions.FilterProcedureSpec{
				Expression: expression.Expression{
					Root: &expression.BinaryNode{
						Operator: expression.GreaterThanOperator,
						Left: &expression.ReferenceNode{
							Name: "$",
						},
						Right: &expression.FloatLiteralNode{
							Value: 5,
						},
					},
				},
			},
			data: []execute.Block{
				&executetest.Block{
					Bnds: execute.Bounds{
						Start: 1,
						Stop:  3,
					},
					ColMeta: []execute.ColMeta{
						{Label: "time", Type: execute.TTime},
						{Label: "value", Type: execute.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), 3.0},
						{execute.Time(2), 6.0},
						{execute.Time(2), 1.0},
					},
				},
				&executetest.Block{
					Bnds: execute.Bounds{
						Start: 3,
						Stop:  5,
					},
					ColMeta: []execute.ColMeta{
						{Label: "time", Type: execute.TTime},
						{Label: "value", Type: execute.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(3), 3.0},
						{execute.Time(3), 2.0},
						{execute.Time(4), 8.0},
					},
				},
			},
			want: []*executetest.Block{
				{
					Bnds: execute.Bounds{
						Start: 1,
						Stop:  3,
					},
					ColMeta: []execute.ColMeta{
						{Label: "time", Type: execute.TTime},
						{Label: "value", Type: execute.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(2), 6.0},
					},
				},
				{
					Bnds: execute.Bounds{
						Start: 3,
						Stop:  5,
					},
					ColMeta: []execute.ColMeta{
						{Label: "time", Type: execute.TTime},
						{Label: "value", Type: execute.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(4), 8.0},
					},
				},
			},
		},
		{
			name: "$>5 and t1 = a and t2 = y",
			spec: &functions.FilterProcedureSpec{
				Expression: expression.Expression{
					Root: &expression.BinaryNode{
						Operator: expression.AndOperator,
						Left: &expression.BinaryNode{
							Operator: expression.GreaterThanOperator,
							Left: &expression.ReferenceNode{
								Name: "$",
							},
							Right: &expression.FloatLiteralNode{
								Value: 5,
							},
						},
						Right: &expression.BinaryNode{
							Operator: expression.AndOperator,
							Left: &expression.BinaryNode{
								Operator: expression.EqualOperator,
								Left: &expression.ReferenceNode{
									Name: "t1",
								},
								Right: &expression.StringLiteralNode{
									Value: "a",
								},
							},
							Right: &expression.BinaryNode{
								Operator: expression.EqualOperator,
								Left: &expression.ReferenceNode{
									Name: "t2",
								},
								Right: &expression.StringLiteralNode{
									Value: "y",
								},
							},
						},
					},
				},
			},
			data: []execute.Block{&executetest.Block{
				Bnds: execute.Bounds{
					Start: 1,
					Stop:  3,
				},
				ColMeta: []execute.ColMeta{
					{Label: "time", Type: execute.TTime},
					{Label: "value", Type: execute.TFloat},
					{Label: "t1", Type: execute.TString, IsTag: true, IsCommon: true},
					{Label: "t2", Type: execute.TString, IsTag: true, IsCommon: false},
				},
				Data: [][]interface{}{
					{execute.Time(1), 1.0, "a", "x"},
					{execute.Time(2), 6.0, "a", "x"},
					{execute.Time(3), 8.0, "a", "y"},
				},
			}},
			want: []*executetest.Block{{
				Bnds: execute.Bounds{
					Start: 1,
					Stop:  3,
				},
				ColMeta: []execute.ColMeta{
					{Label: "time", Type: execute.TTime},
					{Label: "value", Type: execute.TFloat},
					{Label: "t1", Type: execute.TString, IsTag: true, IsCommon: true},
					{Label: "t2", Type: execute.TString, IsTag: true, IsCommon: false},
				},
				Data: [][]interface{}{
					{execute.Time(3), 8.0, "a", "y"},
				},
			}},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			executetest.ProcessTestHelper(
				t,
				tc.data,
				tc.want,
				func(d execute.Dataset, c execute.BlockBuilderCache) execute.Transformation {
					return functions.NewFilterTransformation(d, c, tc.spec)
				},
			)
		})
	}
}

func TestFilter_PushDown_Single(t *testing.T) {
	lp := &plan.LogicalPlanSpec{
		Procedures: map[plan.ProcedureID]*plan.Procedure{
			plan.ProcedureIDFromOperationID("from"): {
				ID: plan.ProcedureIDFromOperationID("from"),
				Spec: &functions.FromProcedureSpec{
					Database: "mydb",
				},
				Parents:  nil,
				Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("range")},
			},
			plan.ProcedureIDFromOperationID("range"): {
				ID: plan.ProcedureIDFromOperationID("range"),
				Spec: &functions.RangeProcedureSpec{
					Bounds: plan.BoundsSpec{
						Stop: query.Now,
					},
				},
				Parents:  []plan.ProcedureID{plan.ProcedureIDFromOperationID("from")},
				Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("filter")},
			},
			plan.ProcedureIDFromOperationID("filter"): {
				ID: plan.ProcedureIDFromOperationID("filter"),
				Spec: &functions.FilterProcedureSpec{
					Expression: expression.Expression{
						Root: &expression.BinaryNode{
							Operator: expression.NotEqualOperator,
							Left: &expression.ReferenceNode{
								Name: "_measurement",
								Kind: "tag",
							},
							Right: &expression.StringLiteralNode{
								Value: "mem",
							},
						},
					},
				},
				Parents: []plan.ProcedureID{
					(plan.ProcedureIDFromOperationID("range")),
				},
				Children: nil,
			},
		},
		Order: []plan.ProcedureID{
			plan.ProcedureIDFromOperationID("from"),
			plan.ProcedureIDFromOperationID("range"),
			plan.ProcedureIDFromOperationID("filter"),
		},
	}

	want := &plan.PlanSpec{
		Bounds: plan.BoundsSpec{
			Stop: query.Now,
		},
		Procedures: map[plan.ProcedureID]*plan.Procedure{
			plan.ProcedureIDFromOperationID("from"): {
				ID: plan.ProcedureIDFromOperationID("from"),
				Spec: &functions.FromProcedureSpec{
					Database:  "mydb",
					BoundsSet: true,
					Bounds: plan.BoundsSpec{
						Stop: query.Now,
					},
					FilterSet: true,
					Filter: expression.Expression{
						Root: &expression.BinaryNode{
							Operator: expression.NotEqualOperator,
							Left: &expression.ReferenceNode{
								Name: "_measurement",
								Kind: "tag",
							},
							Right: &expression.StringLiteralNode{
								Value: "mem",
							},
						},
					},
				},
				Children: []plan.ProcedureID{},
			},
		},
		Results: []plan.ProcedureID{
			(plan.ProcedureIDFromOperationID("from")),
		},
		Order: []plan.ProcedureID{
			plan.ProcedureIDFromOperationID("from"),
		},
	}

	plantest.PhysicalPlanTestHelper(t, lp, want)
}

func TestFilter_PushDown_Branch(t *testing.T) {
	lp := &plan.LogicalPlanSpec{
		Procedures: map[plan.ProcedureID]*plan.Procedure{
			plan.ProcedureIDFromOperationID("from"): {
				ID: plan.ProcedureIDFromOperationID("from"),
				Spec: &functions.FromProcedureSpec{
					Database: "mydb",
				},
				Parents:  nil,
				Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("range")},
			},
			plan.ProcedureIDFromOperationID("range"): {
				ID: plan.ProcedureIDFromOperationID("range"),
				Spec: &functions.RangeProcedureSpec{
					Bounds: plan.BoundsSpec{
						Stop: query.Now,
					},
				},
				Parents: []plan.ProcedureID{plan.ProcedureIDFromOperationID("from")},
				Children: []plan.ProcedureID{
					plan.ProcedureIDFromOperationID("filterA"),
					plan.ProcedureIDFromOperationID("filterB"),
				},
			},
			plan.ProcedureIDFromOperationID("filterA"): {
				ID: plan.ProcedureIDFromOperationID("filterA"),
				Spec: &functions.FilterProcedureSpec{
					Expression: expression.Expression{
						Root: &expression.BinaryNode{
							Operator: expression.NotEqualOperator,
							Left: &expression.ReferenceNode{
								Name: "_measurement",
								Kind: "tag",
							},
							Right: &expression.StringLiteralNode{
								Value: "A",
							},
						},
					},
				},
				Parents: []plan.ProcedureID{
					(plan.ProcedureIDFromOperationID("range")),
				},
				Children: nil,
			},
			plan.ProcedureIDFromOperationID("filterB"): {
				ID: plan.ProcedureIDFromOperationID("filterB"),
				Spec: &functions.FilterProcedureSpec{
					Expression: expression.Expression{
						Root: &expression.BinaryNode{
							Operator: expression.NotEqualOperator,
							Left: &expression.ReferenceNode{
								Name: "_measurement",
								Kind: "tag",
							},
							Right: &expression.StringLiteralNode{
								Value: "B",
							},
						},
					},
				},
				Parents: []plan.ProcedureID{
					(plan.ProcedureIDFromOperationID("range")),
				},
				Children: nil,
			},
		},
		Order: []plan.ProcedureID{
			plan.ProcedureIDFromOperationID("from"),
			plan.ProcedureIDFromOperationID("range"),
			plan.ProcedureIDFromOperationID("filterA"),
			plan.ProcedureIDFromOperationID("filterB"), // FilterB is last so it will be duplicated
		},
	}

	fromID := plan.ProcedureIDFromOperationID("from")
	fromIDDup := plan.ProcedureIDForDuplicate(fromID)
	want := &plan.PlanSpec{
		Bounds: plan.BoundsSpec{
			Stop: query.Now,
		},
		Procedures: map[plan.ProcedureID]*plan.Procedure{
			fromID: {
				ID: fromID,
				Spec: &functions.FromProcedureSpec{
					Database:  "mydb",
					BoundsSet: true,
					Bounds: plan.BoundsSpec{
						Stop: query.Now,
					},
					FilterSet: true,
					Filter: expression.Expression{
						Root: &expression.BinaryNode{
							Operator: expression.NotEqualOperator,
							Left: &expression.ReferenceNode{
								Name: "_measurement",
								Kind: "tag",
							},
							Right: &expression.StringLiteralNode{
								Value: "A",
							},
						},
					},
				},
				Children: []plan.ProcedureID{},
			},
			fromIDDup: {
				ID: fromIDDup,
				Spec: &functions.FromProcedureSpec{
					Database:  "mydb",
					BoundsSet: true,
					Bounds: plan.BoundsSpec{
						Stop: query.Now,
					},
					FilterSet: true,
					Filter: expression.Expression{
						Root: &expression.BinaryNode{
							Operator: expression.NotEqualOperator,
							Left: &expression.ReferenceNode{
								Name: "_measurement",
								Kind: "tag",
							},
							Right: &expression.StringLiteralNode{
								Value: "B",
							},
						},
					},
				},
				Parents:  []plan.ProcedureID{},
				Children: []plan.ProcedureID{},
			},
		},
		Results: []plan.ProcedureID{
			fromID,
			fromIDDup,
		},
		Order: []plan.ProcedureID{
			fromID,
			fromIDDup,
		},
	}

	plantest.PhysicalPlanTestHelper(t, lp, want)
}