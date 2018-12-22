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

// map song's download command to songId
// to deal with multi commands creating
var SongCommandMap map[string]music.Song

var bot *telebot.Bot

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s <token>\n", os.Args[0])
		return
	}

	// map init
	AudiosMap = map[string]*telebot.Audio{}
	SongCommandMap = map[string]music.Song{}

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
		_, _ = bot.Send(msg.Sender, "*Enjoy free music!* 发送歌名，聆听和下载音乐", telebot.ModeMarkdown)
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

	bot.Handle(telebot.OnText, func(msg *telebot.Message) {
		if !msg.Private() {
			return
		}

		query := msg.Text
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
			if len(song.SongId) == 0 {
				continue
			}
			command := fmt.Sprintf("/dl_%s", song.SongId)
			singersStr := getSingersStr(&song)
			songBuilder += fmt.Sprintf("🎵 *%s*\n%s \n点击 [%s](%s) 收听\n\n", song.Name, singersStr, command, command)

			SongCommandMap[command] = song

			log.Printf("handle command: %s with %s\n", command, song.SongId)
			// a inner command to send audio, use like: /dl_songId
			bot.Handle(command, func(msg *telebot.Message) {
				if !msg.Private() {
					return
				}
				// do in goroutines
				go func() {
					if !sendAudio(command, key, msg.Sender) {
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
	bot.Handle("/echo", func(msg *telebot.Message) {
		if !msg.Private() {
			return
		}
		text := msg.Payload
		text = strings.TrimSuffix(text, "吗？")
		text = strings.TrimSuffix(text, "吗?")
		text += "!"
		_, _ = bot.Send(msg.Sender, text)
	})

	bot.Start()
}

func getSingersStr(song *music.Song) string {
	singersStr := ""
	for index, singer := range song.Singers {
		if len(singer) != 0 {
			if index != 0 {
				singersStr += "/"
			}
			singersStr += singer
		}
	}
	return singersStr
}

// if the audio not upload to telegram, upload first then send.
// so this function will cast times
func sendAudio(command string, key string, receiver *telebot.User) bool {
	song := SongCommandMap[command]
	songId := song.SongId
	if len(songId) == 0 {
		return false
	}
	// audio has been uploaded
	if AudiosMap[songId] != nil {
		log.Printf("%s is in map, send directly.", songId)
		_, err := bot.Send(receiver, AudiosMap[songId])
		return err == nil
	}

	url, ok := music.GetSongUrl(songId, key)

	log.Printf("songId: %s:\n%s", songId, url)

	if ok {
		resp, err := http.Get(url)
		if err != nil {
			log.Println(err.Error())

			_, _ = bot.Send(receiver, err.Error())
			return false
		}
		songFile := &telebot.Audio{File: telebot.FromReader(resp.Body)}
		// record file and send
		AudiosMap[songId] = songFile
		_, err = bot.Send(receiver, songFile)
		// send song name for unknown songs
		_, _ = bot.Send(receiver, fmt.Sprintf("%s-%s", song.Name, getSingersStr(&song)))
		return err == nil
	}

	return false
}
