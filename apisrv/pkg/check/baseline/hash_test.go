package baseline

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	bytes20  []byte = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0}
	bytes20b []byte = []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	bytes32  []byte = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2}
	bytes32b []byte = []byte{2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 1}
)

func TestNewHash(t *testing.T) {
	a, err := NewHash(bytes20)
	assert.NoError(t, err)
	assert.NotNil(t, a.Sha1)
	assert.Nil(t, a.Sha256)
	assert.Equal(t, bytes20, (*a.Sha1)[:])

	b, err := NewHash(bytes32)
	assert.NoError(t, err)
	assert.Nil(t, b.Sha1)
	assert.NotNil(t, b.Sha256)
	assert.Equal(t, bytes32, (*b.Sha256)[:])

	c, err := NewHash([]byte{})
	assert.NoError(t, err)
	assert.Nil(t, c.Sha1)
	assert.Nil(t, c.Sha256)

	_, err = NewHash([]byte{1})
	assert.Error(t, err)
}

func TestIntersectHash(t *testing.T) {
	a, _ := NewHash(bytes20)
	b, _ := NewHash(bytes20)
	c := Hash{}
	d, _ := NewHash(bytes32)
	e, _ := NewHash(bytes32)
	f, _ := NewHash(bytes20b)
	g, _ := NewHash(bytes32b)

	assert.True(t, a.IntersectsWith(&a))
	assert.True(t, a.IntersectsWith(&b))
	assert.True(t, a.IntersectsWith(&c))
	assert.True(t, a.IntersectsWith(&d))
	assert.True(t, a.IntersectsWith(&e))
	assert.False(t, a.IntersectsWith(&f))
	assert.True(t, a.IntersectsWith(&g))

	assert.True(t, b.IntersectsWith(&a))
	assert.True(t, b.IntersectsWith(&b))
	assert.True(t, b.IntersectsWith(&c))
	assert.True(t, b.IntersectsWith(&d))
	assert.True(t, b.IntersectsWith(&e))
	assert.False(t, b.IntersectsWith(&f))
	assert.True(t, b.IntersectsWith(&g))

	assert.True(t, c.IntersectsWith(&a))
	assert.True(t, c.IntersectsWith(&b))
	assert.True(t, c.IntersectsWith(&c))
	assert.True(t, c.IntersectsWith(&d))
	assert.True(t, c.IntersectsWith(&e))
	assert.True(t, c.IntersectsWith(&f))
	assert.True(t, c.IntersectsWith(&g))

	assert.True(t, d.IntersectsWith(&a))
	assert.True(t, d.IntersectsWith(&b))
	assert.True(t, d.IntersectsWith(&c))
	assert.True(t, d.IntersectsWith(&d))
	assert.True(t, d.IntersectsWith(&e))
	assert.True(t, d.IntersectsWith(&f))
	assert.False(t, d.IntersectsWith(&g))

	mix1 := Hash{Sha1: dupSha1(bytes20), Sha256: dupSha256(bytes32)}
	mix2 := Hash{Sha1: dupSha1(bytes20b), Sha256: dupSha256(bytes32)}
	mix3 := Hash{Sha1: dupSha1(bytes20b), Sha256: dupSha256(bytes32b)}

	assert.True(t, mix1.IntersectsWith(&a))
	assert.True(t, mix1.IntersectsWith(&d))
	assert.False(t, mix1.IntersectsWith(&f))
	assert.False(t, mix1.IntersectsWith(&g))
	assert.True(t, mix1.IntersectsWith(&mix1))
	assert.False(t, mix1.IntersectsWith(&mix2))
	assert.False(t, mix1.IntersectsWith(&mix3))

	assert.False(t, mix2.IntersectsWith(&a))
	assert.True(t, mix2.IntersectsWith(&d))
	assert.True(t, mix2.IntersectsWith(&f))
	assert.False(t, mix2.IntersectsWith(&g))
	assert.False(t, mix2.IntersectsWith(&mix1))
	assert.True(t, mix2.IntersectsWith(&mix2))
	assert.False(t, mix2.IntersectsWith(&mix3))

	assert.False(t, mix3.IntersectsWith(&a))
	assert.False(t, mix3.IntersectsWith(&d))
	assert.True(t, mix3.IntersectsWith(&f))
	assert.True(t, mix3.IntersectsWith(&g))
	assert.False(t, mix3.IntersectsWith(&mix1))
	assert.False(t, mix3.IntersectsWith(&mix2))
	assert.True(t, mix3.IntersectsWith(&mix3))

	assert.True(t, mix1.IntersectsWith(&c))
	assert.True(t, mix2.IntersectsWith(&c))
	assert.True(t, mix3.IntersectsWith(&c))
}

func TestUnionHash(t *testing.T) {
	a, _ := NewHash(bytes20)
	assert.False(t, a.UnionWith(&a))
	assert.Equal(t, (*a.Sha1)[:], bytes20)
	assert.Nil(t, a.Sha256)

	a, _ = NewHash(bytes20)
	b, _ := NewHash(bytes20)
	assert.False(t, a.UnionWith(&b))
	assert.Equal(t, (*a.Sha1)[:], bytes20)
	assert.Nil(t, a.Sha256)

	a = Hash{}
	assert.True(t, a.UnionWith(&b))
	assert.Equal(t, (*a.Sha1)[:], bytes20)
	assert.Nil(t, a.Sha256)

	a = Hash{}
	assert.False(t, a.UnionWith(&Hash{}))
	assert.Nil(t, a.Sha256)

	a, _ = NewHash(bytes20)
	b, _ = NewHash(bytes32)
	assert.True(t, a.UnionWith(&b))
	assert.Equal(t, (*a.Sha1)[:], bytes20)
	assert.Equal(t, (*a.Sha256)[:], bytes32)

	a, _ = NewHash(bytes32)
	b, _ = NewHash(bytes20)
	assert.True(t, a.UnionWith(&b))
	assert.Equal(t, (*a.Sha1)[:], bytes20)
	assert.Equal(t, (*a.Sha256)[:], bytes32)

	a = Hash{Sha1: dupSha1(bytes20), Sha256: dupSha256(bytes32)}
	assert.False(t, a.UnionWith(&a))
	assert.Equal(t, (*a.Sha1)[:], bytes20)
	assert.Equal(t, (*a.Sha256)[:], bytes32)

	a = Hash{Sha1: dupSha1(bytes20), Sha256: dupSha256(bytes32)}
	assert.False(t, a.UnionWith(&Hash{}))
	assert.Equal(t, (*a.Sha1)[:], bytes20)
	assert.Equal(t, (*a.Sha256)[:], bytes32)

	a = Hash{Sha1: dupSha1(bytes20), Sha256: dupSha256(bytes32)}
	b, _ = NewHash(bytes20)
	c, _ := NewHash(bytes32)
	assert.False(t, a.UnionWith(&b))
	assert.False(t, a.UnionWith(&c))
	assert.Equal(t, (*a.Sha1)[:], bytes20)
	assert.Equal(t, (*a.Sha256)[:], bytes32)

	a = Hash{Sha1: dupSha1(bytes20), Sha256: dupSha256(bytes32)}
	b, _ = NewHash(bytes20b)
	c, _ = NewHash(bytes32b)
	assert.False(t, a.UnionWith(&b))
	assert.False(t, a.UnionWith(&c))
	assert.Equal(t, (*a.Sha1)[:], bytes20)
	assert.Equal(t, (*a.Sha256)[:], bytes32)
}

func TestHashSerialize(t *testing.T) {
	a := Hash{}
	b, _ := NewHash(bytes20)
	c, _ := NewHash(bytes32)
	d := Hash{Sha1: dupSha1(bytes20), Sha256: dupSha256(bytes32)}

	for _, hash := range []Hash{a, b, c, d} {
		buf, err := json.Marshal(hash)
		assert.NoError(t, err)

		fmt.Println(string(buf))
		hash2 := Hash{}
		err = json.Unmarshal(buf, &hash2)
		assert.NoError(t, err)

		assert.Equal(t, hash, hash2)
	}
}
