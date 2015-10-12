package crypto

import (
	"bytes"
	"crypto/cipher"
	"fmt"
	"unsafe"
)

//#cgo CFLAGS: -O3 -Wall -I.
//#cgo LDFLAGS: -L.
//#cgo darwin,amd64 LDFLAGS: -lcrypto_darwin_amd64
//#cgo linux,amd64 LDFLAGS: -lcrypto_linux_amd64
//#cgo linux,aarch64 LDFLAGS: -lcrypto_linux_aarch64
//#cgo windows,amd64 LDFLAGS: -lcrypto_windows_amd64
//#include "openssl.h"
import "C"

type evp_ctx_ptr *C.EVP_CIPHER_CTX
type evp_cipher_ptr *C.EVP_CIPHER

type aesCipher struct {
	cipher evp_cipher_ptr
	key    []byte
}

type AESEncrypter struct {
	*aesCipher
	ctx evp_ctx_ptr
}

type AESDecrypter struct {
	*aesCipher
	ctx evp_ctx_ptr
}

type BLOCK_MODE int

const (
	MODE_CTR BLOCK_MODE = 0
	MODE_OFB BLOCK_MODE = 1
)

var (
	ERR_BAD_KEY_LENGTH = fmt.Errorf("BAD_KEY_LENGTH")
)

var aes_ciphers []evp_cipher_ptr

func init() {
	aes_ciphers = []evp_cipher_ptr{
		C.EVP_aes_128_ctr(),
		C.EVP_aes_128_ofb(),
		C.EVP_aes_192_ctr(),
		C.EVP_aes_192_ctr(), // no 192-ofb
		C.EVP_aes_256_ctr(),
		C.EVP_aes_256_ofb(),
	}
}

func NewAESCipher(key []byte, mode BLOCK_MODE) (*aesCipher, error) {
	var cipher evp_cipher_ptr
	switch len(key) {
	case 16:
		cipher = aes_ciphers[mode]
	case 24:
		cipher = aes_ciphers[2+mode]
	case 32:
		cipher = aes_ciphers[4+mode]
	default:
		return nil, ERR_BAD_KEY_LENGTH
	}
	return &aesCipher{cipher, key}, nil
}

func NewAESEncrypter(aes *aesCipher, iv []byte) (cipher.Stream, error) {
	var ctx C.EVP_CIPHER_CTX
	c_key := (*C.uint8_t)(unsafe.Pointer(&aes.key[0]))
	c_iv := (*C.uint8_t)(unsafe.Pointer(&iv[0]))
	ret := C.EVP_CipherInit_ex(&ctx, aes.cipher, nil, c_key, c_iv, 1)
	if ret == bssl_ok {
		return &AESEncrypter{
			aesCipher: aes,
			ctx:       &ctx,
		}, nil
	} else {
		return nil, fmt.Errorf(get_error())
	}
}

func NewAESDecrypter(aes *aesCipher, iv []byte) (cipher.Stream, error) {
	var ctx C.EVP_CIPHER_CTX
	c_key := (*C.uint8_t)(unsafe.Pointer(&aes.key[0]))
	c_iv := (*C.uint8_t)(unsafe.Pointer(&iv[0]))
	ret := C.EVP_CipherInit_ex(&ctx, aes.cipher, nil, c_key, c_iv, 0)
	if ret == bssl_ok {
		return &AESDecrypter{
			aesCipher: aes,
			ctx:       &ctx,
		}, nil
	} else {
		return nil, fmt.Errorf(get_error())
	}
}

func (a *AESEncrypter) XORKeyStream(dst, src []byte) {
	var out_len int
	c_out_len := (*C.int)(unsafe.Pointer(&out_len))
	c_in := (*C.uint8_t)(unsafe.Pointer(&src[0]))
	c_out := (*C.uint8_t)(unsafe.Pointer(&dst[0]))
	C.EVP_EncryptUpdate(a.ctx, c_out, c_out_len, c_in, (C.int)(len(dst)))
}

func (a *AESDecrypter) XORKeyStream(dst, src []byte) {
	var out_len int
	c_out_len := (*C.int)(unsafe.Pointer(&out_len))
	c_in := (*C.uint8_t)(unsafe.Pointer(&src[0]))
	c_out := (*C.uint8_t)(unsafe.Pointer(&dst[0]))
	C.EVP_DecryptUpdate(a.ctx, c_out, c_out_len, c_in, (C.int)(len(dst)))
}

const (
	EVP_AEAD_AES_GCM_TAG_LEN    = 16
	EVP_AEAD_AES_GCM_NONCE_LEN  = 12
	EVP_AEAD_MAX_KEY_LENGTH     = 80
	EVP_AEAD_MAX_NONCE_LENGTH   = 16
	EVP_AEAD_MAX_OVERHEAD       = 64
	EVP_AEAD_DEFAULT_TAG_LENGTH = 0
	bssl_ok                     = C.int(1)
)

type AES_GCM struct {
	ctx *C.EVP_AEAD_CTX
}

func NewAES_GCM(key []byte) (cipher.AEAD, error) {
	var key_len int
	var aead *C.EVP_AEAD
	switch len(key) {
	case 16:
		aead = C.EVP_aead_aes_128_gcm()
		key_len = 16
	case 32:
		aead = C.EVP_aead_aes_256_gcm()
		key_len = 32
	default:
		return nil, ERR_BAD_KEY_LENGTH
	}
	var ctx C.EVP_AEAD_CTX
	c_key := (*C.uint8_t)(unsafe.Pointer(&key[0]))
	// tag_len > EVP_AEAD_AES_GCM_TAG_LEN then error
	// tag_len = EVP_AEAD_DEFAULT_TAG_LENGTH then EVP_AEAD_AES_GCM_TAG_LEN
	ret := C.EVP_AEAD_CTX_init(&ctx, aead, c_key, C.size_t(key_len), EVP_AEAD_AES_GCM_TAG_LEN, nil)
	if ret == bssl_ok {
		return &AES_GCM{&ctx}, nil
	} else {
		return nil, fmt.Errorf(get_error())
	}
}

func (a *AES_GCM) NonceSize() int {
	return EVP_AEAD_AES_GCM_NONCE_LEN
}

func (a *AES_GCM) Overhead() int {
	return EVP_AEAD_AES_GCM_TAG_LEN
}

func (a *AES_GCM) Seal(dst, nonce, plaintext, data []byte) []byte {
	c_in := (*C.uint8_t)(unsafe.Pointer(&plaintext[0]))
	c_out := (*C.uint8_t)(unsafe.Pointer(&dst[0]))
	c_nonce := (*C.uint8_t)(unsafe.Pointer(&nonce[0]))
	var c_ad *C.uint8_t
	if len(data) > 0 {
		c_ad = (*C.uint8_t)(unsafe.Pointer(&data[0]))
	}
	var out_len C.size_t
	ret := C.EVP_AEAD_CTX_seal(a.ctx,
		c_out, &out_len, C.size_t(len(dst)),
		c_nonce, C.size_t(len(nonce)),
		c_in, C.size_t(len(plaintext)),
		c_ad, C.size_t(len(data)))
	if ret == bssl_ok {
		return dst[:int(out_len)]
	} else {
		panic(get_error())
	}
}

func (a *AES_GCM) Open(dst, nonce, ciphertext, data []byte) ([]byte, error) {
	c_in := (*C.uint8_t)(unsafe.Pointer(&ciphertext[0]))
	c_out := (*C.uint8_t)(unsafe.Pointer(&dst[0]))
	c_nonce := (*C.uint8_t)(unsafe.Pointer(&nonce[0]))
	var c_ad *C.uint8_t
	if len(data) > 0 {
		c_ad = (*C.uint8_t)(unsafe.Pointer(&data[0]))
	}
	var out_len C.size_t
	ret := C.EVP_AEAD_CTX_open(a.ctx,
		c_out, &out_len, C.size_t(len(dst)),
		c_nonce, C.size_t(len(nonce)),
		c_in, C.size_t(len(ciphertext)),
		c_ad, C.size_t(len(data)))
	if ret == bssl_ok {
		return dst[:int(out_len)], nil
	} else {
		return nil, fmt.Errorf(get_error())
	}
}

func get_error() string {
	packed_error := C.ERR_get_error()
	str_buf := make([]byte, C.ERR_ERROR_STRING_BUF_LEN)
	C.ERR_error_string_n(packed_error, (*C.char)(unsafe.Pointer(&str_buf[0])), C.ERR_ERROR_STRING_BUF_LEN)
	n := bytes.IndexByte(str_buf, 0)
	if n > 0 {
		return string(str_buf[:n])
	} else {
		return fmt.Sprintf("error:%x", packed_error)
	}
}

func has_aes_hardware() int {
	return int(C.EVP_has_aes_hardware())
}

func is_NEON_capable() int {
	return int(C.CRYPTO_is_NEON_capable())
}

func dumpCPUID() []uint32 {
	var cpuid [4]uint32
	cpuid = *(*[4]uint32)(unsafe.Pointer(&C.OPENSSL_ia32cap_P))
	return cpuid[:]
}
