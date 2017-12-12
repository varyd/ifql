package functions

import (
	"fmt"

	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/plan"
)

const MaxKind = "max"

type MaxOpSpec struct {
	UseRowTime bool `json:"useRowtime"`
}

func init() {
	query.RegisterMethod(MaxKind, createMaxOpSpec)
	query.RegisterOpSpec(MaxKind, newMaxOp)
	plan.RegisterProcedureSpec(MaxKind, newMaxProcedure, MaxKind)
	execute.RegisterTransformation(MaxKind, createMaxTransformation)
}

func createMaxOpSpec(args query.Arguments, ctx *query.Context) (query.OperationSpec, error) {
	spec := new(MaxOpSpec)
	if useRowTime, ok, err := args.GetBool("useRowTime"); err != nil {
		return nil, err
	} else if ok {
		spec.UseRowTime = useRowTime
	}

	return spec, nil
}

func newMaxOp() query.OperationSpec {
	return new(MaxOpSpec)
}

func (s *MaxOpSpec) Kind() query.OperationKind {
	return MaxKind
}

type MaxProcedureSpec struct {
	UseRowTime bool
}

func newMaxProcedure(qs query.OperationSpec) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*MaxOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}
	return &MaxProcedureSpec{
		UseRowTime: spec.UseRowTime,
	}, nil
}

func (s *MaxProcedureSpec) Kind() plan.ProcedureKind {
	return MaxKind
}
func (s *MaxProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(MaxProcedureSpec)
	ns.UseRowTime = s.UseRowTime
	return ns
}

type MaxSelector struct {
	set  bool
	rows []execute.Row
}

func createMaxTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, ctx execute.Context) (execute.Transformation, execute.Dataset, error) {
	ps, ok := spec.(*MaxProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", ps)
	}
	t, d := execute.NewRowSelectorTransformationAndDataset(id, mode, ctx.Bounds(), new(MaxSelector), ps.UseRowTime, ctx.Allocator())
	return t, d, nil
}

type MaxIntSelector struct {
	MaxSelector
	max int64
}
type MaxUIntSelector struct {
	MaxSelector
	max uint64
}
type MaxFloatSelector struct {
	MaxSelector
	max float64
}

func (s *MaxSelector) NewBoolSelector() execute.DoBoolRowSelector {
	return nil
}

func (s *MaxSelector) NewIntSelector() execute.DoIntRowSelector {
	return new(MaxIntSelector)
}

func (s *MaxSelector) NewUIntSelector() execute.DoUIntRowSelector {
	return new(MaxUIntSelector)
}

func (s *MaxSelector) NewFloatSelector() execute.DoFloatRowSelector {
	return new(MaxFloatSelector)
}

func (s *MaxSelector) NewStringSelector() execute.DoStringRowSelector {
	return nil
}

func (s *MaxSelector) Rows() []execute.Row {
	if !s.set {
		return nil
	}
	return s.rows
}

func (s *MaxSelector) selectRow(idx int, rr execute.RowReader) {
	// Capture row
	if idx >= 0 {
		s.rows = []execute.Row{execute.ReadRow(idx, rr)}
	}
}

func (s *MaxIntSelector) DoInt(vs []int64, rr execute.RowReader) {
	maxIdx := -1
	for i, v := range vs {
		if !s.set || v > s.max {
			s.set = true
			s.max = v
			maxIdx = i
		}
	}
	s.selectRow(maxIdx, rr)
}
func (s *MaxUIntSelector) DoUInt(vs []uint64, rr execute.RowReader) {
	maxIdx := -1
	for i, v := range vs {
		if !s.set || v > s.max {
			s.set = true
			s.max = v
			maxIdx = i
		}
	}
	s.selectRow(maxIdx, rr)
}
func (s *MaxFloatSelector) DoFloat(vs []float64, rr execute.RowReader) {
	maxIdx := -1
	for i, v := range vs {
		if !s.set || v > s.max {
			s.set = true
			s.max = v
			maxIdx = i
		}
	}
	s.selectRow(maxIdx, rr)
}
