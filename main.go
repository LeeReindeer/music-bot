package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"gopkg.in/tucnak/telebot.v2"

	"github.com/LeeReindeer/music-bot/music"
)

// map songId to Audio
// to avoid re-upload file
var AudiosMap map[string]*telebot.Audio

var bot *telebot.Bot

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s <token>\n", os.Args[0])
		return
	}

	AudiosMap = map[string]*telebot.Audio{}
	var err error
	bot, err = telebot.NewBot(telebot.Settings{
		Token:  os.Args[1],
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		log.Fatal(err)
		return
	}

	bot.Handle("/start", func(msg *telebot.Message) {
		if !msg.Private() {
			return
		}
		_, _ = bot.Send(msg.Sender, "*Enjoy free music!*", telebot.ModeMarkdown)
	})

	bot.Handle("/echo", func(msg *telebot.Message) {
		// all commands is personal chat only
		if !msg.Private() {
			return
		}
		_, _ = bot.Send(msg.Sender, msg.Payload)
	})

	/*
		nextPageBtn := telebot.InlineButton{
			Unique: "next_page",
			Text:   ">>>",
		}

		prevPageBtn := telebot.InlineButton{
			Unique: "prev_page",
			Text:   "<<<",
		}

		inlineBtns := [][]telebot.InlineButton{
			[]telebot.InlineButton{prevPageBtn, nextPageBtn},
		}
	*/

	bot.Handle("/music", func(msg *telebot.Message) {
		if !msg.Private() {
			return
		}

		query := msg.Payload
		songs := music.GetSongList(query, 1)
		_, _ = bot.Send(msg.Sender, songs)

		// use a same key for one query
		key, ok := music.GetSongKey()
		if !ok {
			_, _ = bot.Send(msg.Sender, "错误: 无效的Key")
			return
		}
		songBuilder := ""
		for _, song := range songs {
			command := fmt.Sprintf("/dl_%s", song.SongId)
			singersStr := ""
			for _, singer := range song.Singers {
				if len(singer) != 0 {
					singersStr += (singer + "\\")
				}
			}
			songBuilder += fmt.Sprintf("*%s* - %s \nListen: [%s](%s)\n", song.Name, singersStr, command, command)
			// a inner command to send audio, use like: /dl_songId
			bot.Handle(command, func(msg *telebot.Message) {
				if !msg.Private() {
					return
				}
				// do in goroutines
				go func() {
					if !sendAudio(song.SongId, key, msg.Sender) {
						_, _ = bot.Send(msg.Sender, "/(ㄒoㄒ)/~~抱歉，没有找到的资源")
					}
				}()
			})
		}

		// send query result
		_, _ = bot.Send(msg.Sender, songBuilder, telebot.ModeMarkdown)
		// todo paging
		// _, _ = bot.Send(msg.Sender, songs, &telebot.ReplyMarkup{
		// 	InlineKeyboard: inlineBtns,
		// })
	})

	// silly AI
	bot.Handle(telebot.OnText, func(msg *telebot.Message) {
		if !msg.Private() {
			return
		}
		text := msg.Text
		strings.TrimSuffix(text, "吗？")
		strings.TrimSuffix(text, "吗?")
		_, _ = bot.Send(msg.Sender, text)
	})

	bot.Start()
}

/*
func sendAudioTest(receiver *telebot.User) bool {
	url := "http://dl.stream.qqmusic.qq.com/M800000cWOZ64cs3Oa.mp3?vkey=B58E60DDC6ED956A84DD2F20768BD2DB74B53D008C726EC2FC7188AE8A2F2EEF0CF3D2AD9C26F94AAF41A82B29A30BE8B8481824748C6858&guid=5150825362&fromtag=1"
	// var out *os.File
	// if _, err := os.Stat("000cWOZ64cs3Oa.mp3"); !os.IsNotExist(err) {
	// download file
	// out, _ = os.Create("000cWOZ64cs3Oa.mp3")
	// defer out.Close()
	resp, err := http.Get(url)
	if err != nil {
		_, _ = bot.Send(receiver, err.Error())
		return false
	}
	// }
	// save file
	// _, _ = io.Copy(out, resp.Body)
	audio := &telebot.Audio{File: telebot.FromReader(resp.Body)}
	_, err = bot.Send(receiver, audio)
	if err != nil {
		_, _ = bot.Send(receiver, err.Error())
	}
	return err == nil
}
*/

// fixme
// if the audio not upload to telegram, upload first then send.
// so this function will cast times
func sendAudio(songId string, key string, receiver *telebot.User) bool {
	// audio has been uploaded
	if AudiosMap[songId] != nil {
		_, err := bot.Send(receiver, AudiosMap[songId])
		return err == nil
	}

	url, ok := music.GetSongUrl(songId, key)
	if ok {
		resp, err := http.Get(url)
		if err != nil {
			_, _ = bot.Send(receiver, err.Error())
			return false
		}
		songFile := &telebot.Audio{File: telebot.FromReader(resp.Body)}
		// record file and send
		AudiosMap[songId] = songFile
		_, err = bot.Send(receiver, songFile)
		return err == nil
	}

	return false
}
