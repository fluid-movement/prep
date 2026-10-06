package domain

import (
	"crypto/rand"
	"io"
	"time"
)

// Issue IDs are ULIDs: 48 bits of Unix milliseconds and 80 random bits,
// written as 26 characters of Crockford base32. They sort by creation time
// as strings, and two branches cannot create the same one.

const crockford = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

// ULIDLen is the length of an issue ID.
const ULIDLen = 26

// ValidID reports whether s is a ULID in canonical (upper case) form.
func ValidID(s string) bool {
	if len(s) != ULIDLen || s[0] > '7' { // the first character holds 3 bits
		return false
	}
	for k := 0; k < len(s); k++ {
		if !isCrockford(s[k]) {
			return false
		}
	}
	return true
}

func isCrockford(c byte) bool {
	switch {
	case c >= '0' && c <= '9':
		return true
	case c < 'A' || c > 'Z':
		return false
	}
	return c != 'I' && c != 'L' && c != 'O' && c != 'U'
}

// ULIDTime returns the creation time a valid ULID encodes.
func ULIDTime(id string) time.Time {
	var ms int64
	for k := 0; k < 10; k++ {
		ms = ms<<5 | int64(decodeCrockford(id[k]))
	}
	return time.UnixMilli(ms).UTC()
}

func decodeCrockford(c byte) byte {
	for k := 0; k < len(crockford); k++ {
		if crockford[k] == c {
			return byte(k)
		}
	}
	return 0
}

// NewULID encodes now's milliseconds and 80 bits read from entropy (nil
// means crypto/rand).
func NewULID(now time.Time, entropy io.Reader) (string, error) {
	if entropy == nil {
		entropy = rand.Reader
	}
	var b [16]byte
	ms := uint64(now.UnixMilli())
	for k := 5; k >= 0; k-- {
		b[k] = byte(ms)
		ms >>= 8
	}
	if _, err := io.ReadFull(entropy, b[6:]); err != nil {
		return "", err
	}
	return encodeULID(b), nil
}

// nextULID returns a new ULID that sorts after every existing ID created
// in the same millisecond: when one does, the newest is incremented by one
// in its random part, as monotonic ULID generators do.
func nextULID(now time.Time, entropy io.Reader, ids []string) (string, error) {
	id, err := NewULID(now, entropy)
	if err != nil {
		return "", err
	}
	newest := ""
	for _, o := range ids {
		if len(o) == ULIDLen && o[:10] == id[:10] && o > newest {
			newest = o
		}
	}
	if newest < id {
		return id, nil
	}
	b := decodeULID(newest)
	for k := 15; k >= 6; k-- {
		b[k]++
		if b[k] != 0 {
			return encodeULID(b), nil
		}
	}
	// 2^80 IDs in one millisecond: move on to the next.
	return nextULID(now.Add(time.Millisecond), entropy, ids)
}

func encodeULID(b [16]byte) string {
	var out [ULIDLen]byte
	// 128 bits as 26 five-bit groups, the first holding the top 3 bits.
	hi := uint64(b[0])<<56 | uint64(b[1])<<48 | uint64(b[2])<<40 | uint64(b[3])<<32 | uint64(b[4])<<24 | uint64(b[5])<<16 | uint64(b[6])<<8 | uint64(b[7])
	lo := uint64(b[8])<<56 | uint64(b[9])<<48 | uint64(b[10])<<40 | uint64(b[11])<<32 | uint64(b[12])<<24 | uint64(b[13])<<16 | uint64(b[14])<<8 | uint64(b[15])
	for k := ULIDLen - 1; k >= 0; k-- {
		out[k] = crockford[lo&31]
		lo = lo>>5 | hi<<59
		hi >>= 5
	}
	return string(out[:])
}

func decodeULID(id string) [16]byte {
	var hi, lo uint64
	for k := 0; k < ULIDLen; k++ {
		hi = hi<<5 | lo>>59
		lo = lo<<5 | uint64(decodeCrockford(id[k]))
	}
	var b [16]byte
	for k := 7; k >= 0; k-- {
		b[k], b[k+8] = byte(hi), byte(lo)
		hi >>= 8
		lo >>= 8
	}
	return b
}
