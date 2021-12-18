package proto

import (
	"fmt"

	"github.com/go-faster/errors"
)

// BlockInfo describes block.
type BlockInfo struct {
	Overflows bool
	BucketNum int
}

func (i BlockInfo) String() string {
	return fmt.Sprintf("overflows: %v, buckets: %d", i.Overflows, i.BucketNum)
}

const endField = 0 // end of field pairs

// fields of BlockInfo.
const (
	blockInfoOverflows = 1
	blockInfoBucketNum = 2
)

// Encode to Buffer.
func (i BlockInfo) Encode(b *Buffer) {
	b.PutUVarInt(blockInfoOverflows)
	b.PutBool(i.Overflows)

	b.PutUVarInt(blockInfoBucketNum)
	b.PutInt32(int32(i.BucketNum))

	b.PutUVarInt(endField)
}

func (i *BlockInfo) Decode(r *Reader) error {
	for {
		f, err := r.UVarInt()
		if err != nil {
			return errors.Wrap(err, "field id")
		}
		switch f {
		case blockInfoOverflows:
			v, err := r.Bool()
			if err != nil {
				return errors.Wrap(err, "overflows")
			}
			i.Overflows = v
		case blockInfoBucketNum:
			v, err := r.Int32()
			if err != nil {
				return errors.Wrap(err, "bucket number")
			}
			i.BucketNum = int(v)
		case endField:
			return nil
		default:
			return errors.Errorf("unknown field %d", f)
		}
	}
}

type InputColumn struct {
	Name string
	Data Input
}

type ResultColumn struct {
	Name string
	Data Result
}

func (c InputColumn) EncodeStart(buf *Buffer) {
	buf.PutString(c.Name)
	buf.PutString(string(c.Data.Type()))
}

// Input column.
type Input interface {
	Type() ColumnType
	Rows() int
	EncodeColumn(b *Buffer)
}

// Result column.
type Result interface {
	Type() ColumnType
	Rows() int
	DecodeColumn(r *Reader, rows int) error
	Reset()
}

type Block struct {
	Info    BlockInfo
	Columns int
	Rows    int
}

func (b Block) EncodeAware(buf *Buffer, version int) {
	if FeatureBlockInfo.In(version) {
		b.Info.Encode(buf)
	}

	buf.PutInt(b.Columns)
	buf.PutInt(b.Rows)
}

func (b Block) EncodeBlock(buf *Buffer, version int, input []InputColumn) error {
	b.EncodeAware(buf, version)
	for _, col := range input {
		if r := col.Data.Rows(); r != b.Rows {
			return errors.Errorf("%q has %d rows, expected %d", col.Name, r, b.Rows)
		}
		col.EncodeStart(buf)
		col.Data.EncodeColumn(buf)
	}
	return nil
}

const (
	maxColumnsInBlock = 1_000_000
	maxRowsInBlock    = 1_000_000
)

func checkRows(n int) error {
	if n < 0 || n > maxRowsInBlock {
		return errors.Errorf("invalid: %d < %d < %d",
			0, n, maxRowsInBlock,
		)
	}
	return nil
}

func (b *Block) End() bool {
	return b.Columns == 0 && b.Rows == 0
}

func (b *Block) DecodeRawBlock(r *Reader, target []ResultColumn) error {
	{
		v, err := r.Int()
		if err != nil {
			return errors.Wrap(err, "columns")
		}
		if v > maxColumnsInBlock || v < 0 {
			return errors.Errorf("invalid columns number %d", v)
		}
		b.Columns = v
	}
	{
		v, err := r.Int()
		if err != nil {
			return errors.Wrap(err, "rows")
		}
		if err := checkRows(v); err != nil {
			return errors.Wrap(err, "rows count")
		}
		b.Rows = v
	}

	if b.End() {
		// End of data, special case.
		return nil
	}

	var (
		noTarget        = len(target) == 0
		noRows          = b.Rows == 0
		columnsMismatch = b.Columns != len(target)
		allowMismatch   = noTarget && noRows
	)
	if columnsMismatch && !allowMismatch {
		return errors.Errorf("%d (columns) != %d (target)", b.Columns, len(target))
	}
	for i := 0; i < b.Columns; i++ {
		columnName, err := r.Str()
		if err != nil {
			return errors.Wrapf(err, "column [%d] name", i)
		}
		columnType, err := r.Str()
		if err != nil {
			return errors.Wrapf(err, "column [%d] type", i)
		}
		if noTarget {
			// Just reading types and names.
			continue
		}

		t := target[i]
		// Checking column name and type.
		if t.Name != columnName {
			return errors.Errorf("[%d]: unexpected column %q (%q expected)", i, columnName, t.Name)
		}
		var (
			gotType = ColumnType(columnType)
			hasType = t.Data.Type()
		)
		if gotType.Conflicts(hasType) {
			return errors.Errorf("[%d]: %s: unexpected type %q (got) instead of %q (has)",
				i, columnName, gotType, hasType,
			)
		}
		t.Data.Reset()
		if err := t.Data.DecodeColumn(r, b.Rows); err != nil {
			return errors.Wrap(err, columnName)
		}
	}
	return nil
}

func (b *Block) DecodeBlock(r *Reader, version int, target []ResultColumn) error {
	if FeatureBlockInfo.In(version) {
		if err := b.Info.Decode(r); err != nil {
			return errors.Wrap(err, "info")
		}
	}
	if err := b.DecodeRawBlock(r, target); err != nil {
		return err
	}

	return nil
}
