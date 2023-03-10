package bstates

import (
	"crypto/sha256"
	"encoding/binary"
	"hash"
)

// As mentioned in https://github.com/mitchellh/hashstructure/issues/32:
//
//	  """
//	  ...hashstructure uses non-cryptographic hash...
//	  ...Cryptographic hash algorithms are slow but good for detecting falsification.
//		    So, it is useful to identify data...
//	  ...non-cryptographic hash algorithms are fast and good to detect data change
//	     of the same object but very bad to detect falsification. So, it is not convenient
//	     to use them as identification values....
//	  """
//
// It is recommended to implement a custom hasher that uses cryptographic hash in
// order to better identify a struct.
// SHA256Hasher implements the hash.Hash64 interface, used by hashstructure, by calling
// sha256 hash functions.
type SHA256Hasher struct {
	hash.Hash64
	sha256Hasher hash.Hash
}

func NewSHA256Hasher() hash.Hash64 {
	return &SHA256Hasher{
		sha256Hasher: sha256.New(),
	}
}

func (s *SHA256Hasher) Write(p []byte) (n int, err error) {
	return s.sha256Hasher.Write(p)
}

func (s *SHA256Hasher) Sum(b []byte) []byte {
	return s.sha256Hasher.Sum(b)
}

func (s *SHA256Hasher) Reset() {
	s.sha256Hasher.Reset()
}

func (s *SHA256Hasher) Size() int {
	return s.sha256Hasher.Size()
}

func (s *SHA256Hasher) BlockSize() int {
	return s.sha256Hasher.BlockSize()
}

func (s *SHA256Hasher) Sum64() uint64 {
	raw := s.sha256Hasher.Sum(nil)
	return binary.BigEndian.Uint64(raw[:8])
}
