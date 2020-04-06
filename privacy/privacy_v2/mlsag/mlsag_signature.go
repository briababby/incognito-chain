package mlsag

import (
	"encoding/hex"
	"errors"

	"github.com/incognitochain/incognito-chain/privacy/operation"
)

type MlsagSig struct {
	c         operation.Scalar      // 32 bytes
	keyImages []*operation.Point    // 32 * size bytes
	r         [][]*operation.Scalar // 32 * size_1 * size_2 bytes
}

func (this *MlsagSig) ToHex() (string, error) {
	b, err := this.ToBytes()
	if err != nil {
		return "", errors.New("Error in MLSAG MlsagSig ToHex: the signature is broken (size of keyImages and r differ")
	}
	return hex.EncodeToString(b), nil
}

func (this *MlsagSig) ToBytes() ([]byte, error) {
	var b []byte
	b = append(b, MlsagPrefix)

	// Number of private keys should be up to 2^8 only (1 byte)
	var length byte = byte(len(this.keyImages))

	b = append(b, length)
	b = append(b, this.c.ToBytesS()...)
	for i := 0; i < int(length); i += 1 {
		b = append(b, this.keyImages[i].ToBytesS()...)
	}
	for i := 0; i < len(this.r); i += 1 {
		if int(length) != len(this.r[i]) {
			return []byte{}, errors.New("Error in MLSAG MlsagSig ToBytes: the signature is broken (size of keyImages and r differ)")
		}
		for j := 0; j < int(length); j += 1 {
			b = append(b, this.r[i][j].ToBytesS()...)
		}
	}
	return b, nil
}

func (this *MlsagSig) FromHex(s string) (*MlsagSig, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, errors.New("Error in MLSAG MlsagSig FromHex: the signature hex is broken")
	}
	return this.FromBytes(b)
}

// Get from byte and store to signature
func (this *MlsagSig) FromBytes(b []byte) (*MlsagSig, error) {
	if len(b)%HashSize != 2 {
		return nil, errors.New("Error in MLSAG MlsagSig FromBytes: the signature byte is broken (missing byte)")
	}
	if b[0] != MlsagPrefix {
		return nil, errors.New("Error in MLSAG MlsagSig FromBytes: the signature byte is broken (first byte is not mlsag)")
	}

	// Get size at index 0
	var m int = int(b[1])
	lenArr := len(b) - HashSize - 2 - m*32
	n := lenArr / HashSize / m

	if len(b) != 2+(1+m+m*n)*HashSize {
		return nil, errors.New("Error in MLSAG MlsagSig FromBytes: the signature byte is broken (some scalar is missing)")
	}

	if this == nil {
		this = new(MlsagSig)
	}

	offset := 2
	this.c = *new(operation.Scalar).FromBytesS(b[offset : offset+HashSize])
	offset += HashSize

	this.keyImages = make([]*operation.Point, m)
	for i := 0; i < m; i += 1 {
		val, err := new(operation.Point).FromBytesS(b[offset : offset+HashSize])
		if err != nil {
			return nil, errors.New("Error in MLSAG MlsagSig FromBytes: the signature byte is broken (keyImages is broken)")
		}
		this.keyImages[i] = val
		offset += HashSize
	}

	this.r = make([][]*operation.Scalar, n)
	for i := 0; i < n; i += 1 {
		row := make([]*operation.Scalar, m)
		for j := 0; j < m; j += 1 {
			row[j] = new(operation.Scalar).FromBytesS(b[offset : offset+HashSize])
			offset += HashSize
		}
		this.r[i] = row
	}

	return this, nil
}
