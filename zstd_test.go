package bstates

import (
	"testing"

	"github.com/nayarsystems/buffer/buffer"
	"github.com/nayarsystems/buffer/shuffling"
	"github.com/stretchr/testify/require"
)

func Test_ZstdBufferTransposeCompression(t *testing.T) {
	b := &buffer.Buffer{}
	b.InitFromRawBuffer(make([]byte, 10000))

	for i := 0; i < b.GetByteSize(); i++ {
		b.GetRawBuffer()[i] = uint8(i)
	}

	bt, err := shuffling.TransposeBits(b, 8)
	require.Nil(t, err)
	require.Equal(t, b.GetByteSize(), bt.GetByteSize())

	eb, err := ZstdEnc(b)
	require.Nil(t, err)
	db, err := ZstdDec(eb)
	require.Nil(t, err)
	require.Equal(t, b, db)

	ebt, err := ZstdEnc(bt)
	require.Nil(t, err)
	dbt, err := ZstdDec(ebt)
	require.Nil(t, err)
	require.Equal(t, bt, dbt)

	require.Less(t, ebt.GetByteSize(), eb.GetByteSize())
}
