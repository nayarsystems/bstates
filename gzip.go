package bstates

import (
	"bytes"
	"compress/gzip"
	"io"

	"github.com/nayarsystems/buffer/buffer"
)

func GzipEnc(b *buffer.Buffer) (*buffer.Buffer, error) {
	buf := new(bytes.Buffer)
	wr, err := gzip.NewWriterLevel(buf, gzip.BestCompression)
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

func GzipDec(b *buffer.Buffer) (*buffer.Buffer, error) {
	r := bytes.NewReader(b.GetRawBuffer()[:b.GetByteSize()])
	var gzr *gzip.Reader
	gzr, err := gzip.NewReader(r)
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
