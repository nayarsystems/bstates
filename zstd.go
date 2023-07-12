package bstates

import (
	"bytes"
	"github.com/klauspost/compress/zstd"
	"io"

	"github.com/nayarsystems/buffer/buffer"
)

func ZstdEnc(b *buffer.Buffer) (*buffer.Buffer, error) {
	buf := new(bytes.Buffer)
	wr, err := zstd.NewWriter(buf, zstd.WithEncoderConcurrency(1))
	if err != nil {
		return nil, err
	}
	_, err = wr.Write(b.GetRawBuffer()[:b.GetByteSize()])
	wr.Close()
	if err != nil {
		return nil, err
	}
	var out []byte
	out, err = io.ReadAll(buf)
	if err != nil {
		return nil, err
	}
	outb := &buffer.Buffer{}
	outb.InitFromRawBuffer(out)
	return outb, nil
}

func ZstdDec(b *buffer.Buffer) (*buffer.Buffer, error) {
	r := bytes.NewReader(b.GetRawBuffer()[:b.GetByteSize()])
	var gzr *zstd.Decoder
	gzr, err := zstd.NewReader(r, zstd.WithDecoderConcurrency(1))
	if err != nil {
		return nil, err
	}
	var out []byte
	out, err = io.ReadAll(gzr)
	gzr.Close()
	if err != nil {
		return nil, err
	}
	outb := &buffer.Buffer{}
	outb.InitFromRawBuffer(out)
	return outb, nil
}
