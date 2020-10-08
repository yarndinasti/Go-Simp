package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/JustHumanz/Go-simp/engine"
	"github.com/JustHumanz/Go-simp/livestream/youtube"

	"github.com/JustHumanz/Go-simp/database"

	twitterscraper "github.com/n0madic/twitter-scraper"
	log "github.com/sirupsen/logrus"
)

type Video struct {
	ID      string
	Preview string
}

func Tweet(Group string, NameID int64, Limit int) {
	var (
		a []string
	)
	if NameID > 0 {
		tmp := GetHastagMember(NameID)
		a = append(a, tmp)
	} else {
		for _, hashtag := range GetHashtag(Group) {
			if hashtag.TwitterHashtags != "" {
				a = append(a, hashtag.TwitterHashtags)
			}
		}
	}
	Hashtag := strings.Join(a, " OR ")
	log.Info(Hashtag)

	for tweet := range twitterscraper.SearchTweets(context.Background(), Hashtag+"  filter:links -filter:replies filter:media", Limit) {
		engine.BruhMoment(tweet.Error, "", false)
		Data := InputTwitter{
			TwitterData: tweet.Tweet,
			Group:       Group,
			MemberID:    NameID,
		}
		Data.FiltterTweet().InputData()
	}
}

func (Data InputTwitter) FiltterTweet() InputTwitter {
	for _, Hashtag := range GetHashtag(Data.Group) {
		matched, _ := regexp.MatchString(Hashtag.TwitterHashtags, strings.Join(Data.TwitterData.Hashtags, " "))
		if matched {
			Data.MemberID = Hashtag.MemberID
		}
	}
	return Data
}

type InputTwitter struct {
	TwitterData twitterscraper.Tweet
	Group       string
	MemberID    int64
}

func FilterYt(Dat database.Name) {
	YTChannel := strings.Split(Dat.YoutubeID, "\n")
	for _, Channel := range YTChannel {
		VideoID := youtube.GetRSS(Channel)
		body, err := engine.Curl("https://www.googleapis.com/youtube/v3/videos?part=statistics,snippet,liveStreamingDetails&fields=items(snippet(publishedAt,title,description,thumbnails(standard),channelTitle,liveBroadcastContent),liveStreamingDetails(scheduledStartTime,actualEndTime),statistics(viewCount))&id="+strings.Join(VideoID, ",")+"&key="+YtToken, nil)
		if err != nil {
			log.Error(err, string(body))
		}
		var (
			Data    YtData
			Viewers string
			yttype  string
		)
		jsonErr := json.Unmarshal(body, &Data)
		if jsonErr != nil {
			log.Fatal(jsonErr)
		}
		for i := 0; i < len(Data.Items); i++ {
			if Cover, _ := regexp.MatchString("(?m)(cover|song|feat|music|mv)", Data.Items[i].Snippet.Title); Cover {
				yttype = "Covering"
			} else {
				yttype = "Streaming"
			}
			Check, _ := database.CheckVideoID(VideoID[i])
			if Check {
				continue
			} else {
				log.Info("New video")
				//verify
				if Data.Items[i].LiveDetails.Viewers == "" {
					Viewers = Data.Items[i].Statistics.ViewCount
				} else {
					Viewers = Data.Items[i].LiveDetails.Viewers
				}
				NewData := database.YtDbData{
					Status:    Data.Items[i].Snippet.VideoStatus,
					VideoID:   VideoID[i],
					Title:     Data.Items[i].Snippet.Title,
					Thumb:     "http://i3.ytimg.com/vi/" + VideoID[i] + "/maxresdefault.jpg",
					Desc:      Data.Items[i].Snippet.Description,
					Schedul:   Data.Items[i].LiveDetails.StartTime,
					Published: Data.Items[i].Snippet.PublishedAt,
					End:       Data.Items[i].LiveDetails.EndTime,
					Type:      yttype,
					Viewers:   Viewers,
				}

				if Data.Items[i].Snippet.VideoStatus != "upcoming" || Data.Items[i].Snippet.VideoStatus != "live" {
					NewData.Status = "past"
					NewData.InputYt(Dat.ID)
				} else {
					NewData.InputYt(Dat.ID)
				}
			}
		}
	}
}

func (Data Member) YtAvatar() string {
	resp, err := http.Get("https://www.youtube.com/channel/" + Data.YtID[0] + "/about")
	engine.BruhMoment(err, "", false)

	defer resp.Body.Close()
	bit, err := ioutil.ReadAll(resp.Body)
	engine.BruhMoment(err, "", false)

	str := string(bit)
	var avatar string
	re2 := regexp.MustCompile(`(?ms)avatar.*?(http.*?)"`)
	submatchall := re2.FindAllStringSubmatch(str, -1)
	for _, element := range submatchall {
		avatar = strings.Replace(element[1], "s48", "s800", -1)
		break
	}
	return avatar
}

func (Data Member) GetYtSubs() []Subs {
	var datasubs []Subs
	for _, Yt := range Data.YtID {
		var tmp Subs
		head := []string{"Referer", "https://akshatmittal.com/youtube-realtime/"}
		body, err := engine.Curl("https://counts.live/api/youtube-subscriber-count/"+Yt+"/live", head)
		if err != nil {
			log.Error(err, string(body))
		}
		err = json.Unmarshal(body, &tmp)
		datasubs = append(datasubs, tmp)
	}
	return datasubs
}

func (Data Member) GetBiliFolow() BiliStat {
	var (
		wg   sync.WaitGroup
		stat BiliStat
	)
	if Data.BiliRoomID != 0 {
		wg.Add(3)
		go func() {
			body, err := engine.Curl("https://api.bilibili.com/x/relation/stat?vmid="+strconv.Itoa(Data.BiliBiliID), nil)
			if err != nil {
				log.Error(err)
			}
			err = json.Unmarshal(body, &stat.Follow)
			if err != nil {
				log.Error(err)
			}
			defer wg.Done()
		}()

		go func() {
			body, err := engine.Curl("https://api.bilibili.com/x/space/upstat?mid="+strconv.Itoa(Data.BiliBiliID), []string{"Cookie", "SESSDATA=" + BiliSession})
			if err != nil {
				log.Error(err)
			}
			err = json.Unmarshal(body, &stat.Like)
			if err != nil {
				log.Error(err)
			}
			defer wg.Done()
		}()

		go func() {
			baseurl := "https://api.bilibili.com/x/space/arc/search?mid=" + strconv.Itoa(Data.BiliBiliID) + "&ps=100"
			url := []string{baseurl + "&tid=1", baseurl + "&tid=3", baseurl + "&tid=4"}
			for f := 0; f < len(url); f++ {
				body, err := engine.Curl(url[f], nil)
				if err != nil {
					log.Error(err, string(body))
				}
				var video SpaceVideo
				err = json.Unmarshal(body, &video)
				if err != nil {
					log.Error(err)
				}
				stat.Video += video.Data.Page.Count
			}
			defer wg.Done()
		}()
		wg.Wait()
		return stat
	} else {
		log.WithFields(log.Fields{
			"Vtuber": Data.ENName,
		}).Info("BiliBili Space nill")
		return stat
	}
}

func (Data Member) GetTwitterFollow() int {
	profile, err := twitterscraper.GetProfile(Data.TwitterName)
	if err != nil {
		log.Error(err)
	}
	return profile.FollowersCount
}

//yep,i'm trying reduce API cost :3
func GetAvatar(YtChannel string) string {
	resp, err := http.Get("https://www.youtube.com/channel/" + YtChannel + "/about")
	engine.BruhMoment(err, "", false)

	defer resp.Body.Close()
	bit, err := ioutil.ReadAll(resp.Body)
	engine.BruhMoment(err, "", false)

	str := string(bit)
	var avatar string
	re2 := regexp.MustCompile(`(?ms)avatar.*?(http.*?)"`)
	submatchall := re2.FindAllStringSubmatch(str, -1)
	for _, element := range submatchall {
		avatar = strings.Replace(element[1], "s48", "s800", -1)
		break
	}
	return avatar
}

type BiliStat struct {
	Follow BiliFollow
	Like   LikeView
	Video  int
}

type LikeView struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	TTL     int    `json:"ttl"`
	Data    struct {
		Archive struct {
			View int `json:"view"`
		} `json:"archive"`
		Article struct {
			View int `json:"view"`
		} `json:"article"`
		Likes int `json:"likes"`
	} `json:"data"`
}

type BiliFollow struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	TTL     int    `json:"ttl"`
	Data    struct {
		Mid       int `json:"mid"`
		Following int `json:"following"`
		Whisper   int `json:"whisper"`
		Black     int `json:"black"`
		Follower  int `json:"follower"`
	} `json:"data"`
}