package mycrypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"errors"
	"strings"
)

var (
	// ErrInvalidBlockSize indicates hash blocksize <= 0.
	ErrInvalidBlockSize = errors.New("invalid blocksize")

	// ErrInvalidPKCS7Data indicates bad input to PKCS7 pad or unpad.
	ErrInvalidPKCS7Data = errors.New("invalid PKCS7 data (empty or not padded)")

	// ErrInvalidPKCS7Padding indicates PKCS7 unpad fails to bad input.
	ErrInvalidPKCS7Padding = errors.New("invalid padding on input")
)

// pkcs7Pad right-pads the given byte slice with 1 to n bytes, where
// n is the block size. The size of the result is x times n, where x
// is at least 1.
func pkcs7Pad(b []byte, blocksize int) ([]byte, error) {
	if blocksize <= 0 {
		return nil, ErrInvalidBlockSize
	}
	if b == nil || len(b) == 0 {
		return nil, ErrInvalidPKCS7Data
	}
	n := blocksize - (len(b) % blocksize)
	pb := make([]byte, len(b)+n)
	copy(pb, b)
	copy(pb[len(b):], bytes.Repeat([]byte{byte(n)}, n))
	return pb, nil
}

// pkcs7Unpad validates and unpads data from the given bytes slice.
// The returned value will be 1 to n bytes smaller depending on the
// amount of padding, where n is the block size.
func pkcs7Unpad(b []byte, blocksize int) ([]byte, error) {
	if blocksize <= 0 {
		return nil, ErrInvalidBlockSize
	}
	if b == nil || len(b) == 0 {
		return nil, ErrInvalidPKCS7Data
	}
	if len(b)%blocksize != 0 {
		return nil, ErrInvalidPKCS7Padding
	}
	c := b[len(b)-1]
	n := int(c)
	if n == 0 || n > len(b) {
		return nil, ErrInvalidPKCS7Padding
	}
	for i := 0; i < n; i++ {
		if b[len(b)-n+i] != c {
			return nil, ErrInvalidPKCS7Padding
		}
	}
	return b[:len(b)-n], nil
}

//Decrypt decrypts ciphertext
func Decrypt(key, ciphertext string) (*string, error) {
	block, err := newCipherBlock(key)
	if err != nil {
		return nil, err
	}

	iv := key[0:16]

	// CBC mode always works in whole blocks.
	if len(key)%aes.BlockSize != 0 {
		return nil, errors.New("ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, []byte(iv))

	cipherHex, _ := hex.DecodeString(ciphertext)

	// CryptBlocks can work in-place if the two arguments are the same.
	mode.CryptBlocks(cipherHex, cipherHex)
	res := strings.TrimSpace(string(cipherHex))
	return &res, nil
}

func DecryptWithRounds(key string, ciphertext *string, rounds int) (string, error) {
	plaintext := ""
	for i := 1; i <= rounds; i++ {
		ciphertext, _ = Decrypt(key, *ciphertext)
		if i == rounds {
			plaintext = *ciphertext
		}
	}

	return plaintext, nil
}

/*//Encrypt encrypts a plaintext
func Encrypt(key, plaintext string) (string, string, error) {
	saltBs := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, saltBs); err != nil {
		return "", "", err
	}

	salt := hex.EncodeToString(saltBs)

	block, salt, err := newCipherBlock(key, salt)
	if err != nil {
		return "", "", err
	}

	ptbs, _ := pkcs7Pad([]byte(plaintext), block.BlockSize())

	if len(ptbs)%aes.BlockSize != 0 {
		return "", "", errors.New("plaintext is not a multiple of the block size")
	}

	ciphertext := make([]byte, len(ptbs))
	var iv []byte = make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", "", err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, ptbs)

	return hex.EncodeToString(iv) + ":" + hex.EncodeToString(ciphertext), salt, nil
}*/

//Creates a new cipherBlock which can be used to decrypt and encrypt.
//Block ciphers are used to encrypt/decrypt blocks of fixed sizes
func newCipherBlock(key string) (cipher.Block, error) {

	block, err := aes.NewCipher([]byte(key))

	return block, err
}
