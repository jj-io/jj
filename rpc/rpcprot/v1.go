package rpcprot

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"sync/atomic"

	"github.com/jj-io/jj/rpc"

	"gopkg.in/logex.v1"
)

var (
	version = 1
	seq     uint64
)

type Packet struct {
	Meta *Meta
	Data *Data
}

func NewPacket(path string, data interface{}) *Packet {
	return &Packet{
		Meta: NewMeta(path),
	}
}

func (p *Packet) String() string {
	return fmt.Sprintf("meta:%+v data:%+v", p.Meta, p.Data)
}

type Meta struct {
	Version int    `json:"version,omitempty"`
	Seq     uint64 `json:"seq"`
	Path    string `json:"path,omitempty"`
	Error   string `json:"error,omitempty"`
}

func NewMeta(path string) *Meta {
	return &Meta{
		Path: path,
		Seq:  atomic.AddUint64(&seq, 1),
	}
}

func NewMetaError(seq uint64, err string) *Meta {
	return &Meta{
		Error: err,
		Seq:   seq,
	}
}

type ProtocolV1 struct {
	r *io.LimitedReader
	w io.Writer
}

func NewProtocolV1(r io.Reader, w io.Writer) Protocol {
	return &ProtocolV1{
		r: &io.LimitedReader{r, 0},
		w: w,
	}
}

func (p1 *ProtocolV1) Read(buf *bytes.Buffer, metaEnc rpc.Encoding, p *Packet) error {
	p1.r.N += 4
	var length int32
	if err := binary.Read(p1.r, binary.BigEndian, &length); err != nil {
		return logex.Trace(err)
	}
	p1.r.N += int64(length)

	n, err := buf.ReadFrom(p1.r)
	if err == nil && n != int64(length) {
		return logex.Trace(io.ErrUnexpectedEOF)
	}
	if err != nil {
		return logex.Trace(err)
	}

	br := rpc.NewBuffer(buf)
	if err := metaEnc.Decode(br, &p.Meta); err != nil {
		return logex.Trace(err, length, buf.Bytes())
	}
	data, _ := ioutil.ReadAll(br)

	p.Data = NewRawData(data)
	return nil
}

func (p1 *ProtocolV1) Write(metaEnc, bodyEnc rpc.Encoding, p *Packet) error {
	underBuf := make([]byte, 4, 512)
	buf := bytes.NewBuffer(underBuf)
	if err := metaEnc.Encode(buf, p.Meta); err != nil {
		return logex.Trace(err)
	}

	if p.Data != nil {
		if err := bodyEnc.Encode(buf, p.Data.underlay); err != nil {
			return logex.Trace(err)
		}
	}

	logex.Debug("write: ", buf.Bytes())

	binary.BigEndian.PutUint32(underBuf[:4], uint32(buf.Len()-4))
	n, err := p1.w.Write(buf.Bytes())
	if err != nil {
		return logex.Trace(err)
	}
	if n != buf.Len() {
		return logex.Trace(io.ErrShortWrite)
	}
	return nil
}
