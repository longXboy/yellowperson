package util

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/emicklei/go-restful"
	"io"
	"net"
	"strings"
)

func HashPassword(Password string) string {
	sha := sha256.Sum256([]byte(Password))
	return hex.EncodeToString(sha[:])
}
func GetClientIP(req *restful.Request) string {
	forwardip := req.Request.Header.Get("X-Forwarded-For")
	if forwardip != "" {
		idx := strings.Index(forwardip, ",")
		if idx > -1 {
			return forwardip[:idx]
		} else {
			return forwardip
		}
	}
	idx := strings.Index(req.Request.RemoteAddr, ":")
	return req.Request.RemoteAddr[:idx]
}

func GetServerIP() string {
	inters, _ := net.Interfaces()
	ipaddr := make([]string, 3)
	for _, inter := range inters {
		addrs, _ := inter.Addrs()
		if addrs != nil && inter.Name != "lo" {
			ipaddr = strings.Split(addrs[0].String(), `/`)
			return ipaddr[0]
		}
	}
	return ""
}

func Str2sha1(data string) string {
	t := sha1.New()
	io.WriteString(t, data)
	return fmt.Sprintf("%x", t.Sum(nil))
}

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	// 去掉最后一个字节 unpadding 次
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func AesEncrypt(origData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origData = PKCS5Padding(origData, blockSize)
	// origData = ZeroPadding(origData, block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(origData))
	// 根据CryptBlocks方法的说明，如下方式初始化crypted也可以
	// crypted := origData
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

func AesDecrypt(crypted, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(crypted))
	// origData := crypted
	blockMode.CryptBlocks(origData, crypted)
	origData = PKCS5UnPadding(origData)
	// origData = ZeroUnPadding(origData)
	return origData, nil
}
