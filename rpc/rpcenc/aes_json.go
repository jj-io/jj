package rpcenc

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"

	"github.com/jj-io/jj/rpc"

	"gopkg.in/logex.v1"
)

type AesEncoding struct {
	enc    rpc.Encoding
	encode cipher.Stream
	decode cipher.Stream
}

var commonIV = []byte("b1d15254f0f0417d")

func NewAesEncoding(enc rpc.Encoding, key []byte) (*AesEncoding, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, logex.Trace(err)
	}
	return &AesEncoding{
		enc:    enc,
		encode: cipher.NewCFBEncrypter(block, commonIV),
		decode: cipher.NewCFBDecrypter(block, commonIV),
	}, nil
}

func (mp *AesEncoding) Decode(r *bytes.Reader, v interface{}) error {
	buf := bytes.NewBuffer(make([]byte, 0, r.Len()))
	n, _ := r.WriteTo(buf)
	mp.decode.XORKeyStream(buf.Bytes(), buf.Bytes())
	r = bytes.NewReader(buf.Bytes()[:n])
	return logex.Trace(mp.enc.Decode(r, v))
}

func (mp *AesEncoding) Encode(w rpc.BufferWriter, v interface{}) error {
	buf := bytes.NewBuffer(make([]byte, 0, 512))
	if err := mp.enc.Encode(buf, v); err != nil {
		return logex.Trace(err)
	}
	mp.encode.XORKeyStream(buf.Bytes(), buf.Bytes())
	buf.WriteTo(w)
	return nil
}
