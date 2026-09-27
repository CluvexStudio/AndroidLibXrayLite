//go:build android && arm64

package libv2ray

/*
#include <sys/auxv.h>
static unsigned long zed_hwcap(void) { return getauxval(16); }
*/
import "C"

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	_ "unsafe"
)

//go:linkname cpuHWCap internal/cpu.HWCap
var cpuHWCap uint

//go:linkname aesSupportsAES crypto/internal/fips140/aes.supportsAES
var aesSupportsAES bool

//go:linkname gcmSupportsAESGCM crypto/internal/fips140/aes/gcm.supportsAESGCM
var gcmSupportsAESGCM bool

//go:linkname sha256UseSHA2 crypto/internal/fips140/sha256.useSHA2
var sha256UseSHA2 bool

//go:linkname sha512UseSHA512 crypto/internal/fips140/sha512.useSHA512
var sha512UseSHA512 bool

const (
	hwcapAES    = 1 << 3
	hwcapPMULL  = 1 << 4
	hwcapSHA2   = 1 << 6
	hwcapSHA512 = 1 << 21
)

func init() {
	hw := uint(C.zed_hwcap())
	if hw == 0 {
		cpuAccelReport = "hwcap unavailable; crypto stays in software"
		return
	}
	if cpuHWCap == 0 {
		cpuHWCap = hw
	}

	repaired, already, failed := "", "", ""
	note := func(dst *string, name string) { *dst += name + " " }

	switch {
	case hw&hwcapAES == 0:
	case aesSupportsAES:
		note(&already, "aes")
	default:
		aesSupportsAES = true
		gcmSupportsAESGCM = hw&hwcapPMULL != 0
		if aesGcmSelfTest() {
			note(&repaired, "aes+gcm")
		} else {
			aesSupportsAES, gcmSupportsAESGCM = false, false
			note(&failed, "aes")
		}
	}

	switch {
	case hw&hwcapSHA2 == 0:
	case sha256UseSHA2:
		note(&already, "sha256")
	default:
		sha256UseSHA2 = true
		if sha256.Sum256([]byte("abc")) == [32]byte(mustHex("ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad")) {
			note(&repaired, "sha256")
		} else {
			sha256UseSHA2 = false
			note(&failed, "sha256")
		}
	}

	switch {
	case hw&hwcapSHA512 == 0:
	case sha512UseSHA512:
		note(&already, "sha512")
	default:
		sha512UseSHA512 = true
		sum := sha512.Sum512([]byte("abc"))
		want := mustHex("ddaf35a193617abacc417349ae20413112e6fa4e89a97ea20a9eeee64b55d39a2192992a274fc1a836ba3c23a3feebbd454d4423643ce80e2a9ac94fa54ca49f")
		if bytes.Equal(sum[:], want) {
			note(&repaired, "sha512")
		} else {
			sha512UseSHA512 = false
			note(&failed, "sha512")
		}
	}

	cpuAccelReport = fmt.Sprintf("hwcap=0x%x repaired=[%s] already=[%s] rejected=[%s]",
		hw, trim(repaired), trim(already), trim(failed))
}

func aesGcmSelfTest() bool {
	key := mustHex("feffe9928665731c6d6a8f9467308308feffe9928665731c6d6a8f9467308308")
	nonce := mustHex("cafebabefacedbaddecaf888")
	plain := mustHex("d9313225f88406e5a55909c5aff5269a86a7a9531534f7da2e4c303d8a318a721c3c0c95956809532fcf0e2449a6b525b16aedf5aa0de657ba637b391aafd255")
	want := mustHex("522dc1f099567d07f47f37a32a84427d643a8cdcbfe5c0c97598a2bd2555d1aa8cb08e48590dbb3da7b08b1056828838c5f61e6393ba7a0abcc9f662898015adb094dac5d93471bdec1a502270e3cc6c")
	block, err := aes.NewCipher(key)
	if err != nil {
		return false
	}
	g, err := cipher.NewGCM(block)
	if err != nil {
		return false
	}
	sealed := g.Seal(nil, nonce, plain, nil)
	if !bytes.Equal(sealed, want) {
		return false
	}
	back, err := g.Open(nil, nonce, sealed, nil)
	return err == nil && bytes.Equal(back, plain)
}

func mustHex(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic("libv2ray: bad test vector: " + err.Error())
	}
	return b
}

func trim(s string) string {
	if len(s) == 0 {
		return ""
	}
	return s[:len(s)-1]
}
