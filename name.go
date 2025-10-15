package main

import (
	"time"
	"fmt"
	"crypto/rand"
	"crypto/sha256"
	"crypto/hmac"
	"encoding/base64"
	"encoding/hex"
	"golang.org/x/crypto/chacha20poly1305"
)

type Header []byte

type Key []byte

func demoKey(text string) (k Key) {
	h := sha256.New()
	h.Write([]byte("DEMO"))
	h.Write([]byte(text))
	k = h.Sum(nil)
	return
}

func (k Key) id() (id []byte) {
	h := hmac.New(sha256.New, k)
	h.Write([]byte("KID"))

	id = h.Sum(nil)
	return
}

func randomNonce() (n []byte) {
	n = make([]byte, chacha20poly1305.NonceSizeX)
	rand.Read(n)
	return
}

func (key Key) encrypt(plain []byte) ([]byte, error) {
	
	nonce := randomNonce()

	c, err := chacha20poly1305.NewX(key)
        if err != nil { return nil, err}

	return c.Seal(nonce, nonce, plain, nil), nil
}

func (key Key) decrypt(cipher []byte) ([]byte, error) {

	nonce := cipher[0:chacha20poly1305.NonceSizeX]
	cipher = cipher[chacha20poly1305.NonceSizeX:]
	c, err := chacha20poly1305.NewX(key)
	if err != nil { return nil, err}

	return c.Open(nil, nonce, cipher, nil)
}

func decode(key Key, h Header) (plain string, err error) {

	var d []byte
	d, err = key.decrypt(h[20:])
	plain = string(d)
	return
}

func encode(key Key, info string) (h Header, err error) {

	plain := []byte(info)

	cipher, _ := key.encrypt(plain)
	
	// keyid, obj-id, seq, cipher
	h = make([]byte, 4 + 15 + 1 + len(cipher))

	copy(h, key.id()[:4])
	rand.Read(h[4:19])
	h[19] = 0

	copy(h[20:], cipher)
	return
}



func main() {
	info := fmt.Sprintf("%d %d %s", 512, time.Now().Unix(), "foo/bar/dir1/bigfilename.txt")

	k := demoKey("mrh1")

	h, _ := encode(k, info)
	f := hex.EncodeToString(h)
	b := base64.URLEncoding.EncodeToString(h)

	fmt.Printf("chacha20poly1305.NonceSizeX: %d bytes\n", chacha20poly1305.NonceSizeX)
	fmt.Printf("plaintext: len=%d '%s'\n", len(info), info)
	fmt.Printf("base64: len=%d %s\n", len(b), b)
	fmt.Printf("hex: len:%d %s\n", len(f), f)


	bh, _ := base64.URLEncoding.DecodeString(b)
	//k[12] = 42  // corrupt the key to test authentication
	p, e := decode(k, bh)
	fmt.Printf("recovered plain: %s, %s\n", p, e)
}
