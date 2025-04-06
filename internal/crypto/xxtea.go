package crypto

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
)

const (
	delta = 0x9E3779B9
)

var (
	ErrKeyLengthInvalid = errors.New("key cannot be empty")
	ErrDataTooSmall     = errors.New("data too small")
)

// XXTEACipher XXTEA加密实现
type XXTEACipher struct {
	key []byte
}

// NewXXTEACipher 创建XXTEA加密器
func NewXXTEACipher(key []byte) (*XXTEACipher, error) {
	if len(key) == 0 {
		return nil, ErrKeyLengthInvalid
	}

	// 使用SHA-256将任意长度的密钥转换为32字节的哈希值
	hash := sha256.Sum256(key)

	// 使用哈希值的前16字节作为密钥
	k := make([]byte, 16)
	copy(k, hash[:16])

	return &XXTEACipher{
		key: k,
	}, nil
}

// Encrypt 加密数据
func (c *XXTEACipher) Encrypt(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return []byte{}, nil
	}

	v := bytesToUint32s(data)
	k := bytesToUint32s(c.key)

	n := len(v)
	if n < 2 {
		return nil, ErrDataTooSmall
	}

	encrypt(v, k)

	return uint32sToBytes(v), nil
}

// EncryptToBase64 加密数据并转为Base64
func (c *XXTEACipher) EncryptToBase64(data []byte) (string, error) {
	encrypted, err := c.Encrypt(data)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// Decrypt 解密数据
func (c *XXTEACipher) Decrypt(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return []byte{}, nil
	}

	v := bytesToUint32s(data)
	k := bytesToUint32s(c.key)

	n := len(v)
	if n < 2 {
		return nil, ErrDataTooSmall
	}

	decrypt(v, k)

	return uint32sToBytes(v), nil
}

// DecryptFromBase64 从Base64解密数据
func (c *XXTEACipher) DecryptFromBase64(data string) ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}

	return c.Decrypt(decoded)
}

// bytesToUint32s 将字节数组转换为uint32数组
func bytesToUint32s(bytes []byte) []uint32 {
	length := len(bytes)
	// 计算需要的uint32数量，并确保至少有2个
	n := (length + 3) / 4
	if n < 2 {
		n = 2
	}

	v := make([]uint32, n)

	for i := 0; i < length; i++ {
		v[i/4] |= uint32(bytes[i]) << uint32(8*(i%4))
	}

	return v
}

// uint32sToBytes 将uint32数组转换为字节数组
func uint32sToBytes(v []uint32) []byte {
	length := len(v) * 4
	bytes := make([]byte, length)

	for i := 0; i < length; i++ {
		bytes[i] = byte(v[i/4] >> uint32(8*(i%4)))
	}

	return bytes
}

// mx XXTEA算法中的MX函数
func mx(z, y, sum, p, e uint32, k []uint32) uint32 {
	return ((z>>5 ^ y<<2) + (y>>3 ^ z<<4)) ^ ((sum ^ y) + (k[p&3^e] ^ z))
}

// encrypt XXTEA加密算法
func encrypt(v, k []uint32) {
	n := len(v)
	if n < 2 {
		return
	}

	var (
		z   = v[n-1]
		y   = v[0]
		sum = uint32(0)
		e   = uint32(0)
		p   = uint32(0)
		q   = uint32(6 + 52/n)
		i   = uint32(0)
	)

	for i = 0; i < q; i++ {
		sum += delta
		e = (sum >> 2) & 3

		for p = 0; p < uint32(n-1); p++ {
			y = v[p+1]
			v[p] += mx(z, y, sum, p, e, k)
			z = v[p]
		}

		y = v[0]
		v[n-1] += mx(z, y, sum, p, e, k)
		z = v[n-1]
	}
}

// decrypt XXTEA解密算法
func decrypt(v, k []uint32) {
	n := len(v)
	if n < 2 {
		return
	}

	var (
		z   = v[n-1]
		y   = v[0]
		sum = uint32(0)
		e   = uint32(0)
		p   = uint32(0)
		q   = uint32(6 + 52/n)
		i   = uint32(0)
	)

	sum = q * delta

	for i = 0; i < q; i++ {
		e = (sum >> 2) & 3

		for p = uint32(n - 1); p > 0; p-- {
			z = v[p-1]
			v[p] -= mx(z, y, sum, p, e, k)
			y = v[p]
		}

		z = v[n-1]
		v[0] -= mx(z, y, sum, p, e, k)
		y = v[0]

		sum -= delta
	}
}
