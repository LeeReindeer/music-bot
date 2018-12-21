package music

import (
	"fmt"
	"strconv"

	"github.com/tidwall/gjson"

	"gopkg.in/resty.v1"
)

type Key struct {
	Code int    `json:"code"`
	Key  string `json:"key"`
}

type Song struct {
	Name    string `json:"songname"`
	Singers []string
	SongId  string `json:"songmid"`
}

const USER_AGENT string = " Mozilla/5.0 (iPhone; CPU iPhone OS 9_1 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko) Version/9.0 Mobile/13B143 Safari/601.1"

const DEFALUT_LIST_SIZE int = 5

/*
GET: http://c.y.qq.com/soso/fcgi-bin/search_for_qq_cp
query params:
 w: query keywords
 p: page
 n: item in every page
 format = json
example: http://c.y.qq.com/soso/fcgi-bin/search_for_qq_cp?w=lust for life&p=1&n=10&format=json
referer: http://m.y.qq.com
user-agent: Mozilla/5.0 (iPhone; CPU iPhone OS 9_1 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko) Version/9.0 Mobile/13B143 Safari/601.1
*/
func GetSongList(query string, page int) []Song {
	resp, err := resty.R().
		SetQueryParams(map[string]string{
			"w":      query,
			"p":      strconv.Itoa(page),
			"n":      strconv.Itoa(DEFALUT_LIST_SIZE),
			"format": "json",
		}).
		SetHeader("user-agent", USER_AGENT).
		SetHeader("referer", "http://m.y.qq.com").
		Get("http://c.y.qq.com/soso/fcgi-bin/search_for_qq_cp")

	if err != nil || resp.Body() == nil {
		return nil
	}

	songs := make([]Song, DEFALUT_LIST_SIZE)
	i := 0

	// check code
	if gjson.GetBytes(resp.Body(), "code").Int() != 0 {
		return nil
	}

	result := gjson.GetBytes(resp.Body(), "data.song.list")
	result.ForEach(func(key, value gjson.Result) bool {
		songs[i].SongId = value.Get("songmid").String()
		songs[i].Name = value.Get("songname").String()
		songs[i].Singers = make([]string, DEFALUT_LIST_SIZE)

		// iterate to get singers' name
		singers := value.Get("singer.#.name")
		for j, name := range singers.Array() {
			songs[i].Singers[j] = name.String()
			if j >= DEFALUT_LIST_SIZE {
				break
			}
		}
		i++
		if i >= DEFALUT_LIST_SIZE {
			return false
		}
		return true
	})
	return songs
}

// GET: http://base.music.qq.com/fcgi-bin/fcg_musicexpress.fcg?json=3&guid=5150825362&format=json
// referer: http://y.qq.com
func GetSongKey() (key string, ok bool) {
	resp, err := resty.R().
		SetHeader("referer", "http://y.qq.com").
		SetHeader("user-agent", USER_AGENT).
		Get("http://base.music.qq.com/fcgi-bin/fcg_musicexpress.fcg?json=3&guid=5150825362&format=json")
	if err != nil || resp.Body() == nil {
		return "", false
	}
	keyStruct := new(Key)
	// err = json.Unmarshal(resp.Body(), &keyStruct)
	keyStruct.Code = int(gjson.GetBytes(resp.Body(), "code").Int())
	keyStruct.Key = gjson.GetBytes(resp.Body(), "key").String()
	if err != nil || keyStruct.Code != 0 {
		return "", false
	}
	return keyStruct.Key, true
}

// URL for download music, get songId from GetSongList
func GetSongUrl(songId string, key string) (string, bool) {
	quality := []string{"M800", "M500", "C400"}
	for _, q := range quality {
		url := fmt.Sprintf("http://dl.stream.qqmusic.qq.com/%s%s.mp3?vkey=%s&guid=5150825362&fromtag=1", q, songId, key)
		if IsUrlOk(url) {
			return url, true
		}
	}
	return "", false
}

// return true if the request of the url is ok
// use http head method
func IsUrlOk(url string) bool {
	resp, err := resty.R().Head(url)
	return err == nil && resp.IsSuccess()
}
