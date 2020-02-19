/*
 * Go implementation of Google city hash (MIT license)
 * https://code.google.com/p/cityhash/
 *
 * MIT License http://www.opensource.org/licenses/mit-license.php
 *
 * I don't even want to pretend to understand the details of city hash.
 * I am only reproducing the logic in Go as faithfully as I can.
 *
 */

package cityhash

import (
	"encoding/binary"
)

func unalignedLoad64(p []byte) (result uint64) {
	return binary.LittleEndian.Uint64(p)
}

func unalignedLoad32(p []byte) (result uint32) {
	return binary.LittleEndian.Uint32(p)
}

func bswap64(x uint64) uint64 {
	// Copied from netbsd's bswap64.c
	return ((x << 56) & 0xff00000000000000) |
		((x << 40) & 0x00ff000000000000) |
		((x << 24) & 0x0000ff0000000000) |
		((x << 8) & 0x000000ff00000000) |
		((x >> 8) & 0x00000000ff000000) |
		((x >> 24) & 0x0000000000ff0000) |
		((x >> 40) & 0x000000000000ff00) |
		((x >> 56) & 0x00000000000000ff)
}

func bswap32(x uint32) uint32 {
	// Copied from netbsd's bswap32.c
	return ((x << 24) & 0xff000000) |
		((x << 8) & 0x00ff0000) |
		((x >> 8) & 0x0000ff00) |
		((x >> 24) & 0x000000ff)
}

func uint32InExpectedOrder(x uint32) uint32 {
	/*
		if !little {
			return bswap32(x)
		}
	*/

	return x
}

func uint64InExpectedOrder(x uint64) uint64 {
	/*
		if !little {
			return bswap64(x)
		}
	*/

	return x
}

// If I understand the original code correctly, it's expecting to load either 8 or 4
// byes in little endian order. so let's just simplify it a bit since we will do that
// anyway..
// https://code.google.com/p/cityhash/source/browse/trunk/src/city.cc#112
func fetch64(p []byte, index int) uint64 {
	return binary.LittleEndian.Uint64(p[index:])
}

func fetch32(p []byte, index int) uint32 {
	return binary.LittleEndian.Uint32(p[index:])
}

const (
	k0 uint64 = 0xc3a5c85c97cb3127
	k1 uint64 = 0xb492b66fbe98f273
	k2 uint64 = 0x9ae16a3b2f90404f
	k3 uint64 = 0xc949d7c7509e6557 // Magic number copied from gthub.com/Atvaark/CityHash/blob/master/CityHash/CityHash.cs
	c1 uint32 = 0xcc9e2d51
	c2 uint32 = 0x1b873593
)

func fmix(h uint32) uint32 {
	h ^= h >> 16
	h *= 0x85ebca6b
	h ^= h >> 13
	h *= 0xc2b2ae35
	h ^= h >> 16
	return h
}

func hash128to64(x Uint128) uint64 {
	// Murmur-inspired hashing.
	const kMul uint64 = 0x9ddfea08eb382d69
	var a uint64 = (x.Lower64() ^ x.Higher64()) * kMul
	a ^= (a >> 47)
	var b uint64 = (x.Higher64() ^ a) * kMul
	b ^= (b >> 47)
	b *= kMul
	return b
}

func rotate64(val uint64, shift uint32) uint64 {
	// Avoid shifting by 64: doing so yields an undefined result.
	if shift != 0 {
		return ((val >> shift) | (val << (64 - shift)))
	}

	return val
}

func rotateByAtLeast1(val uint64, shift uint) uint64 {
	return (val >> shift) | (val << (64 - shift))
}

func rotate32(val uint32, shift uint32) uint32 {
	// Avoid shifting by 32: doing so yields an undefined result.
	if shift != 0 {
		return ((val >> shift) | (val << (32 - shift)))
	}

	return val
}

func swap64(a, b *uint64) {
	*a, *b = *b, *a
}

func swap32(a, b *uint32) {
	*a, *b = *b, *a
}

func permute3(a, b, c *uint32) {
	swap32(a, b)
	swap32(a, c)
}

func mur(a, h uint32) uint32 {
	a *= c1
	a = rotate32(a, 17)
	a *= c2
	h ^= a
	h = rotate32(h, 19)

	//return h * 5 + 0xe6546b64
	z := h*5 + 0xe6546b64
	return z
}

func hash32Len13to24(s []byte) uint32 {
	length := len(s)
	var a uint32 = fetch32(s, (length>>1)-4)
	var b uint32 = fetch32(s, 4)
	var c uint32 = fetch32(s, length-8)
	var d uint32 = fetch32(s, length>>1)
	var e uint32 = fetch32(s, 0)
	var f uint32 = fetch32(s, length-4)
	var h uint32 = uint32(length)

	return fmix(mur(f, mur(e, mur(d, mur(c, mur(b, mur(a, h)))))))
}

/*private static uint32 Hash32Len13To24(byte[] s)
        {
            uint len = (uint) s.Length;
            uint32 a = Fetch32(s, (len >> 1) - 4);
            uint32 b = Fetch32(s, +4);
            uint32 c = Fetch32(s, +len - 8);
            uint32 d = Fetch32(s, +(len >> 1));
            uint32 e = Fetch32(s, 0);
            uint32 f = Fetch32(s, +len - 4);
            uint32 h = len;
            return Fmix(Mur(f, Mur(e, Mur(d, Mur(c, Mur(b, Mur(a, h)))))));
		}
*/

func hash32Len0to4(s []byte) uint32 {
	var b, c uint32 = 0, 9

	for _, v := range s {
		b = uint32(int64(b)*int64(c1) + int64(int8(v)))
		c ^= b
	}

	return fmix(mur(b, mur(uint32(len(s)), c)))
}

func hash32Len5to12(s []byte) uint32 {
	length := len(s)
	var a, b, c uint32 = uint32(length), uint32(length) * 5, 9
	var d uint32 = b

	a += fetch32(s, 0)
	b += fetch32(s, length-4)
	c += fetch32(s, (length>>1)&4)

	return fmix(mur(c, mur(b, mur(a, d))))
}

/*
 private static uint32 Hash32Len5To12(byte[] s)
        {
            uint len = (uint) s.Length;
            uint32 a = len, b = len*5, c = 9, d = b;
            a += Fetch32(s, 0);
            b += Fetch32(s, len - 4);
            c += Fetch32(s, ((len >> 1) & 4));
            return Fmix(Mur(c, Mur(b, Mur(a, d))));
        }
*/

func CityHash32(s []byte) uint32 {
	length := len(s)
	if length <= 4 {
		return hash32Len0to4(s)
	} else if length <= 12 {
		return hash32Len5to12(s)
	} else if length <= 24 {
		return hash32Len13to24(s)
	}

	// length > 24
	var h uint32 = uint32(length)
	var g uint32 = c1 * uint32(length)
	var f uint32 = g
	var a0 uint32 = rotate32(fetch32(s, length-4)*c1, 17) * c2
	var a1 uint32 = rotate32(fetch32(s, length-8)*c1, 17) * c2
	var a2 uint32 = rotate32(fetch32(s, length-16)*c1, 17) * c2
	var a3 uint32 = rotate32(fetch32(s, length-12)*c1, 17) * c2
	var a4 uint32 = rotate32(fetch32(s, length-20)*c1, 17) * c2
	h ^= a0
	h = rotate32(h, 19)
	h = h*5 + 0xe6546b64
	h ^= a2
	h = rotate32(h, 19)
	h = h*5 + 0xe6546b64
	g ^= a1
	g = rotate32(g, 19)
	g = g*5 + 0xe6546b64
	g ^= a3
	g = rotate32(g, 19)
	g = g*5 + 0xe6546b64
	f += a4
	f = rotate32(f, 19)
	f = f*5 + 0xe6546b64

	var iters int = (length - 1) / 20
	for {
		var a0 uint32 = rotate32(fetch32(s, 0)*c1, 17) * c2
		var a1 uint32 = fetch32(s, 4)
		var a2 uint32 = rotate32(fetch32(s, 8)*c1, 17) * c2
		var a3 uint32 = rotate32(fetch32(s, 12)*c1, 17) * c2
		var a4 uint32 = fetch32(s, 16)
		h ^= a0
		h = rotate32(h, 18)
		h = h*5 + 0xe6546b64
		f += a1
		f = rotate32(f, 19)
		f = f * c1
		g += a2
		g = rotate32(g, 18)
		g = g*5 + 0xe6546b64
		h ^= a3 + a1
		h = rotate32(h, 19)
		h = h*5 + 0xe6546b64
		g ^= a4
		g = bswap32(g) * 5
		h += a4 * 5
		h = bswap32(h)
		f += a0
		permute3(&f, &h, &g)
		s = s[20:]

		iters--
		if iters == 0 {
			break
		}
	}

	g = rotate32(g, 11) * c1
	g = rotate32(g, 17) * c1
	f = rotate32(f, 11) * c1
	f = rotate32(f, 17) * c1
	h = rotate32(h+g, 19)
	h = h*5 + 0xe6546b64
	h = rotate32(h, 17) * c1
	h = rotate32(h+f, 19)
	h = h*5 + 0xe6546b64
	h = rotate32(h, 17) * c1
	return h
}

func shiftMix(val uint64) uint64 {
	return val ^ (val >> 47)
}

type Uint128 [2]uint64

func (this *Uint128) setLower64(l uint64) {
	this[0] = l
}

func (this *Uint128) setHigher64(h uint64) {
	this[1] = h
}

func (this Uint128) Lower64() uint64 {
	return this[0]
}

func (this Uint128) Higher64() uint64 {
	return this[1]
}

func (this Uint128) Bytes() []byte {
	b := make([]byte, 16)
	binary.LittleEndian.PutUint64(b, this[0])
	binary.LittleEndian.PutUint64(b[8:], this[1])
	return b
}

func hashLen16(u, v uint64) uint64 {
	return hash128to64(Uint128{u, v})
}

func hashLen0to16(s []byte, offset int) uint64 {
	length := len(s) - offset
	if length > 8 {
		var a uint64 = fetch64(s, offset)
		var b uint64 = fetch64(s, offset+length-8)
		return hashLen16(a, rotateByAtLeast1(b+uint64(length), uint(length))) ^ b
	}

	if length >= 4 {
		var a uint64 = uint64(fetch32(s, offset))
		return hashLen16(uint64(length)+(a<<3), uint64(fetch32(s, offset+length-4)))
	}

	if length > 0 {
		var a uint8 = uint8(s[0])
		var b uint8 = uint8(s[length>>1])
		var c uint8 = uint8(s[length-1])
		var y uint32 = uint32(a) + (uint32(b) << 8)
		var z uint32 = uint32(length) + (uint32(c) << 2)
		return shiftMix(uint64(y)*k2^uint64(z)*k3) * k2
	}

	return k2
}

/*        private static uint64 HashLen0To16(byte[] s, int offset)
{
    int len = s.Length - offset;
    if (len > 8)
    {
        uint64 a = Fetch64(s, offset);
        uint64 b = Fetch64(s, offset + len - 8);
        return HashLen16(a, RotateByAtLeast1(b + (ulong) len, len)) ^ b;
    }
    if (len >= 4)
    {
        uint64 a = Fetch32(s, offset);
        return HashLen16((uint) len + (a << 3), Fetch32(s, offset + len - 4));
    }
    if (len > 0)
    {
        uint8 a = s[offset];
        uint8 b = s[offset + (len >> 1)];
        uint8 c = s[offset + (len - 1)];
        uint32 y = a + ((uint32) b << 8);
        uint32 z = (uint) len + ((uint32) c << 2);
        return ShiftMix(y*K2 ^ z*K3)*K2;
    }
    return K2;
}
*/

func hashLen17to32(s []byte) uint64 {
	length := len(s)
	var a uint64 = fetch64(s, 0) * k1
	var b uint64 = fetch64(s, 8)
	var c uint64 = fetch64(s, length-8) * k2
	var d uint64 = fetch64(s, length-16) * k0
	return hashLen16(rotate64(a-b, 43)+rotate64(c, 30)+d, a+rotate64(b^k3, 20)-c+uint64(length))
	//HashLen16(Rotate(a - b, 43) + Rotate(c, 30) + d, a + Rotate(b ^ K3, 20) - c + len) (From Atvaark's fork)
}

func weakHashLen32WithSeeds(w, x, y, z, a, b uint64) Uint128 {
	a += w
	b = rotate64(b+a+z, 21)
	var c uint64 = a
	a += x
	a += y
	b += rotate64(a, 44)
	return Uint128{a + z, b + c}
}

func weakHashLen32WithSeeds_3(s []byte, offset int, a, b uint64) Uint128 {
	return weakHashLen32WithSeeds(fetch64(s, offset), fetch64(s, offset+8), fetch64(s, offset+16), fetch64(s, offset+24), a, b)
}

func hashLen33to64(s []byte) uint64 {
	length := len(s)
	var z uint64 = fetch64(s, 24)
	var a uint64 = fetch64(s, 0) + (uint64(length)+fetch64(s, length-16))*k0
	var b uint64 = rotate64(a+z, 52)
	var c uint64 = rotate64(a, 37)
	a += fetch64(s, 8)
	c += rotate64(a, 7)
	a += fetch64(s, 16)
	var vf uint64 = a + z
	var vs uint64 = b + rotate64(a, 31) + c
	a = fetch64(s, 16) + fetch64(s, length-32)
	z = fetch64(s, length-8)
	b = rotate64(a+z, 52)
	c = rotate64(a, 37)
	a += fetch64(s, length-24)
	c += rotate64(a, 7)
	a += fetch64(s, length-16)
	var wf uint64 = a + z
	var ws uint64 = b + rotate64(a, 31) + c
	var r uint64 = shiftMix((vf+ws)*k2 + (wf+vs)*k0)
	return shiftMix(r*k0+vs) * k2
}

func CityHash64(s []byte) uint64 {
	length := len(s)
	if length <= 32 {
		if length <= 16 {
			return hashLen0to16(s, 0)
		} else {
			return hashLen17to32(s)
		}
	} else if length <= 64 {
		return hashLen33to64(s)
	}

	var x uint64 = fetch64(s, length-40)
	var y uint64 = fetch64(s, length-16) + fetch64(s, length-56)
	var z uint64 = hashLen16(fetch64(s, length-48)+uint64(length), fetch64(s, length-24))
	var v Uint128 = weakHashLen32WithSeeds_3(s, length-64, uint64(length), z)
	var w Uint128 = weakHashLen32WithSeeds_3(s, length-32, y+k1, x)
	x = x*k1 + fetch64(s, 0)

	length = (length - 1) & ^(63)
	for {
		x = rotate64(x+y+v.Lower64()+fetch64(s, 8), 37) * k1
		y = rotate64(y+v.Higher64()+fetch64(s, 48), 42) * k1
		x ^= w.Higher64()
		y += v.Lower64() + fetch64(s, 40)
		z = rotate64(z+w.Lower64(), 33) * k1
		v = weakHashLen32WithSeeds_3(s, 0, v.Higher64()*k1, x+w.Lower64())
		w = weakHashLen32WithSeeds_3(s, 32, z+w.Higher64(), y+fetch64(s, 16))
		swap64(&z, &x)
		s = s[64:]
		length -= 64

		if length == 0 {
			break
		}
	}

	return hashLen16(hashLen16(v.Lower64(), w.Lower64())+shiftMix(y)*k1+z, hashLen16(v.Higher64(), w.Higher64())+x)
}

func CityHash64WithSeed(s []byte, seed uint64) uint64 {
	return CityHash64WithSeeds(s, k2, seed)
}

func CityHash64WithSeeds(s []byte, seed0, seed1 uint64) uint64 {
	return hashLen16(CityHash64(s)-seed0, seed1)
}
