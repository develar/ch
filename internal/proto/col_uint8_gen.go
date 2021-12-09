// Code generated by ./cmd/ch-gen-int, DO NOT EDIT.

package proto

import "github.com/go-faster/errors"

// ColumnUInt8 represents UInt8 column.
type ColumnUInt8 []uint8

// Type returns ColumnType of UInt8.
func (ColumnUInt8) Type() ColumnType {
	return ColumnTypeUInt8
}

// Rows returns count of rows in column.
func (c ColumnUInt8) Rows() int {
	return len(c)
}

// Reset resets data in row, preserving capacity for efficiency.
func (c *ColumnUInt8) Reset() {
	*c = (*c)[:0]
}

// EncodeColumn encodes UInt8 rows to *Buffer.
func (c ColumnUInt8) EncodeColumn(b *Buffer) {
	for _, v := range c {
		b.PutUInt8(v)
	}
}

// DecodeColumn decodes UInt8 rows from *Reader.
func (c *ColumnUInt8) DecodeColumn(r *Reader, rows int) error {
	data, err := r.ReadRaw(rows)
	if err != nil {
		return errors.Wrap(err, "read")
	}
	*c = append(*c, data...)
	return nil
}