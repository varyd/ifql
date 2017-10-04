package execute

import (
	"context"
	"io"
	"sync"

	"github.com/influxdata/ifql/query/execute/storage"
	"github.com/influxdata/yarpc"
)

type StorageReader interface {
	Read(rs ReadSpec, start, stop Time) (BlockIterator, error)
	Close()
}

type ReadSpec struct {
	Database   string
	Predicate  *storage.Predicate
	Limit      int64
	Descending bool
}

func NewStorageReader() (StorageReader, error) {
	return &storageReader{}, nil
}

type storageReader struct {
	mu          sync.Mutex
	connections []*yarpc.ClientConn
}

func (sr *storageReader) connect() (*yarpc.ClientConn, error) {
	conn, err := yarpc.Dial("localhost:8082")
	if err != nil {
		return nil, err
	}
	sr.mu.Lock()
	sr.connections = append(sr.connections, conn)
	sr.mu.Unlock()
	return conn, nil
}

func (sr *storageReader) Read(readSpec ReadSpec, start, stop Time) (BlockIterator, error) {
	conn, err := sr.connect()
	if err != nil {
		return nil, err
	}
	c := storage.NewStorageClient(conn)

	var req storage.ReadRequest
	req.Database = readSpec.Database
	req.Predicate = readSpec.Predicate
	//	req.Limit = limit
	req.Descending = readSpec.Descending
	req.TimestampRange.Start = int64(start)
	req.TimestampRange.End = int64(stop)

	stream, err := c.Read(context.Background(), &req)
	if err != nil {
		return nil, err
	}
	bi := &storageBlockIterator{
		bounds: Bounds{
			Start: start,
			Stop:  stop,
		},
		data: &readState{
			stream: stream,
		},
	}
	return bi, nil
}

func (sr *storageReader) Close() {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	for _, c := range sr.connections {
		c.Close()
	}
}

type storageBlockIterator struct {
	bounds Bounds
	data   *readState
}

type readState struct {
	stream storage.Storage_ReadClient
	rep    storage.ReadResponse
}

type responseType int

const (
	seriesType responseType = iota
	integerPointsType
	floatPointsType
)

func (s *readState) peek() responseType {
	frame := s.rep.Frames[0]
	switch {
	case frame.GetSeries() != nil:
		return seriesType
	case frame.GetIntegerPoints() != nil:
		return integerPointsType
	case frame.GetFloatPoints() != nil:
		return floatPointsType
	default:
		panic("read response frame should have one of series, integerPoints, or floatPoints")
	}
}

func (s *readState) more() bool {
	if len(s.rep.Frames) > 0 {
		return true
	}
	if err := s.stream.RecvMsg(&s.rep); err != nil {
		if err == io.EOF {
			// We are done
			return false
		}
		//TODO add proper error handling
		return false
	}
	return true
}

func (s *readState) next() storage.ReadResponse_Frame {
	frame := s.rep.Frames[0]
	s.rep.Frames = s.rep.Frames[1:]
	return frame
}

func (bi *storageBlockIterator) Do(f func(Block)) {
	for bi.data.more() {
		if p := bi.data.peek(); p != seriesType {
			//This means the consumer didn't read all the data off the block
			continue
		}
		frame := bi.data.next()
		s := frame.GetSeries()
		tags := make(Tags)
		for _, t := range s.Tags {
			tags[string(t.Key)] = string(t.Value)
		}
		block := &storageBlock{
			bounds:  bi.bounds,
			tags:    tags,
			colMeta: [2]ColMeta{TimeCol, ValueCol},
			data:    bi.data,
			done:    make(chan struct{}),
		}
		f(block)
		// Wait until the block has been read.
		block.wait()
	}
}

type storageBlock struct {
	bounds  Bounds
	tags    Tags
	colMeta [2]ColMeta

	done chan struct{}

	data *readState
}

func (b *storageBlock) wait() {
	<-b.done
}

func (b *storageBlock) Bounds() Bounds {
	return b.bounds
}
func (b *storageBlock) Tags() Tags {
	return b.tags
}
func (b *storageBlock) Cols() []ColMeta {
	return b.colMeta[:]
}

func (b *storageBlock) Col(c int) ValueIterator {
	return &storageBlockValueIterator{
		data:    b.data,
		colMeta: b.colMeta,
		col:     c,
		done:    b.done,
	}
}

func (b *storageBlock) Times() ValueIterator {
	return b.Col(0)
}
func (b *storageBlock) Values() ValueIterator {
	return b.Col(1)
}

type storageBlockValueIterator struct {
	data *readState
	done chan<- struct{}

	// colMeta is always two columns, where the first is a TimeCol
	// and the second is any Value column.
	colMeta [2]ColMeta
	col     int

	// colBufs are the buffers for the given columns.
	colBufs [2]interface{}

	// resuable buffer for the time column
	timeBuf []Time

	// resuable buffers for the different types of values
	floatBuf  []float64
	stringBuf []string
	intBuf    []int64
}

func (b *storageBlockValueIterator) DoFloat(f func([]float64, RowReader)) {
	checkColType(b.colMeta[b.col], TFloat)
	for b.advance() {
		f(b.colBufs[b.col].([]float64), b)
	}
	close(b.done)
}
func (b *storageBlockValueIterator) DoString(f func([]string, RowReader)) {
	checkColType(b.colMeta[b.col], TString)
	for b.advance() {
		f(b.colBufs[b.col].([]string), b)
	}
	close(b.done)
}

func (b *storageBlockValueIterator) DoTime(f func([]Time, RowReader)) {
	checkColType(b.colMeta[b.col], TTime)
	for b.advance() {
		f(b.colBufs[b.col].([]Time), b)
	}
	close(b.done)
}

func (b *storageBlockValueIterator) AtFloat(i, j int) float64 {
	checkColType(b.colMeta[j], TFloat)
	return b.colBufs[j].([]float64)[i]
}
func (b *storageBlockValueIterator) AtString(i, j int) string {
	checkColType(b.colMeta[j], TString)
	return b.colBufs[j].([]string)[i]
}
func (b *storageBlockValueIterator) AtInt(i, j int) int64 {
	checkColType(b.colMeta[j], TInt)
	return b.colBufs[j].([]int64)[i]
}
func (b *storageBlockValueIterator) AtTime(i, j int) Time {
	checkColType(b.colMeta[j], TTime)
	return b.colBufs[j].([]Time)[i]
}

func (b *storageBlockValueIterator) advance() bool {
	for b.data.more() {

		//reset buffers
		b.timeBuf = b.timeBuf[0:0]
		b.floatBuf = b.floatBuf[0:0]
		b.stringBuf = b.stringBuf[0:0]
		b.intBuf = b.intBuf[0:0]

		switch b.data.peek() {
		case seriesType:
			return false
		case integerPointsType:
			// read next frame
			frame := b.data.next()
			p := frame.GetIntegerPoints()
			l := len(p.Timestamps)
			if l > cap(b.timeBuf) {
				b.timeBuf = make([]Time, l)
			} else {
				b.timeBuf = b.timeBuf[:l]
			}
			if l > cap(b.intBuf) {
				b.intBuf = make([]int64, l)
			} else {
				b.intBuf = b.intBuf[:l]
			}

			for i, c := range p.Timestamps {
				b.timeBuf[i] = Time(c)
				b.intBuf[i] = p.Values[i]
			}
			b.colBufs[0] = b.timeBuf
			b.colBufs[1] = b.intBuf
			return true
		case floatPointsType:
			// read next frame
			frame := b.data.next()
			p := frame.GetFloatPoints()

			l := len(p.Timestamps)
			if l > cap(b.timeBuf) {
				b.timeBuf = make([]Time, l)
			} else {
				b.timeBuf = b.timeBuf[:l]
			}
			if l > cap(b.floatBuf) {
				b.floatBuf = make([]float64, l)
			} else {
				b.floatBuf = b.floatBuf[:l]
			}

			for i, c := range p.Timestamps {
				b.timeBuf[i] = Time(c)
				b.floatBuf[i] = p.Values[i]
			}
			b.colBufs[0] = b.timeBuf
			b.colBufs[1] = b.floatBuf
			return true
		}
	}
	return false
}
