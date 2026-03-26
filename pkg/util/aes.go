package util

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

func FixedLengthEncrypt(text, salt string) string {
	// 创建一个新的 HMAC 实例，使用 SHA256 算法
	// 盐（Salt）在这里作为 HMAC 的 Key
	h := hmac.New(sha256.New, []byte(salt))

	// 写入待加密数据
	h.Write([]byte(text))

	// 计算结果并转换为十六进制字符串
	// h.Sum(nil) 返回的是 []byte，长度固定为 32 字节
	return hex.EncodeToString(h.Sum(nil))
}

func AesEncrypt(orig string, key string) string {
	defer func() {
		if err := recover(); err != nil {
			return
		}
	}()
	origData := []byte(orig)
	// 转成字节数组
	bk := []byte(key)
	// 分组秘钥
	block, err := aes.NewCipher(bk)
	if err != nil {
		return ""
	}
	// 获取秘钥块的长度
	blockSize := block.BlockSize()
	// 补全码
	origData = PKCS7Padding(origData, blockSize)
	// 加密模式
	blockMode := cipher.NewCBCEncrypter(block, bk[:blockSize])
	// 创建数组
	cryted := make([]byte, len(origData))
	// 加密
	blockMode.CryptBlocks(cryted, origData)
	return base64.StdEncoding.EncodeToString(cryted)
}

func AesDecrypt(cryted string, key string) string {
	defer func() {
		if err := recover(); err != nil {
			return
		}
	}()
	// 转成字节数组
	crytedByte, err := base64.StdEncoding.DecodeString(cryted)
	if err != nil {
		return ""
	}
	bk := []byte(key)
	// 分组秘钥
	block, err := aes.NewCipher(bk)
	if err != nil {
		return ""
	}
	// 获取秘钥块的长度
	blockSize := block.BlockSize()
	// 加密模式
	blockMode := cipher.NewCBCDecrypter(block, bk[:blockSize])
	// 创建数组
	orig := make([]byte, len(crytedByte))
	// 解密
	blockMode.CryptBlocks(orig, crytedByte)
	// 去补全码
	orig = PKCS7UnPadding(orig)
	return string(orig)
}

// 补码
func PKCS7Padding(ciphertext []byte, blocksize int) []byte {
	padding := blocksize - len(ciphertext)%blocksize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

// 去码
func PKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}
