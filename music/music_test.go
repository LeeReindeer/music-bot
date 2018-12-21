package music

import (
	"testing"

	"gotest.tools/assert"
)

func TestGetSongKey(t *testing.T) {
	Key, ok := GetSongKey()
	t.Log(Key, ok)
	assert.Assert(t, ok)
}

func TestGetSongList(t *testing.T) {
	songs := GetSongList("lust for life", 1)
	assert.Assert(t, songs != nil)

	for _, song := range songs {
		t.Logf("%s(%s) :%s", song.Name, song.SongId, song.Singers[0])
	}
}

func TestIsUrlOk(t *testing.T) {
	ok := IsUrlOk("http://dl.stream.qqmusic.qq.com/M800000cWOZ64cs3Oa.mp3?vkey=B58E60DDC6ED956A84DD2F20768BD2DB74B53D008C726EC2FC7188AE8A2F2EEF0CF3D2AD9C26F94AAF41A82B29A30BE8B8481824748C6858&guid=5150825362&fromtag=1")
	assert.Assert(t, ok)
}

func TestGetSongUrl(t *testing.T) {
	songs := GetSongList("lust for life", 1)
	assert.Assert(t, songs != nil)

	key, ok := GetSongKey()
	t.Log(key, ok)
	assert.Assert(t, ok)

	url, ok := GetSongUrl(songs[1].SongId, key)
	assert.Assert(t, ok)
	t.Log("url: ", url)
}
