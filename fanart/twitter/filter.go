package twitter

import (
	"database/sql"
	"regexp"
	"strconv"
	"strings"
	"sync"

	database "github.com/JustHumanz/Go-simp/database"
	log "github.com/sirupsen/logrus"
)

//check if new fanart or not
func (Data TwitterStruct) CheckNew() []Statuses {
	var tmp []Statuses
	for _, TwData := range Data.Statuses {
		var (
			id int
		)
		err := db.QueryRow(`SELECT id FROM Twitter WHERE TweetID=? `, TwData.IDStr).Scan(&id)
		if err == sql.ErrNoRows {
			tmp = append(tmp, TwData)
		} else {
			//update
			_, err := db.Exec(`Update Twitter set Likes=? Where id=? `, TwData.FavoriteCount, id)
			if err != nil {
				log.Error(err)
			}
		}
	}
	return tmp
}

//filter hashtag post
func (Data Statuses) CheckHashTag(Group []database.MemberGroupID, wg *sync.WaitGroup) {
	defer wg.Done()
	for _, hashtag := range Data.Entities.Hashtags {
		westtaiwan, _ := regexp.MatchString("(?m)(freecoco|asacocoleak|cocodidbothingwrong|freecocoandhaachama)", strings.ToLower(hashtag.Text))
		if !westtaiwan {
			for i := 0; i < len(Group); i++ {
				if "#"+hashtag.Text == Group[i].TwitterHashtags && Data.User.FollowersCount > 10 && Data.User.FriendsCount > 20 { //fuck off dummy account
					//new
					log.WithFields(log.Fields{
						"Hashtags":   Group[i].TwitterHashtags,
						"MemberName": Group[i].EnName,
					}).Info("Get new post")

					var (
						Photos    []string
						Video     string
						SendMedia string
						msg       string
					)
					for _, Media := range Data.Entities.Media {
						Photos = append(Photos, Media.MediaURLHTTPS)
					}
					for _, vid := range Data.ExtendedEntities.Media {
						if vid.VideoInfo.Variants != nil {
							Video = vid.VideoInfo.Variants[0].URL
						}
					}
					if Photos != nil && Video == "" {
						SendMedia = Photos[0]
						msg = "1/" + strconv.Itoa(len(Data.ExtendedEntities.Media)) + " photos"
					} else if Video != "" {
						SendMedia = Data.ExtendedEntities.Media[0].MediaURLHTTPS
						msg = "Video type,check original post"
					} else {
						SendMedia = "https://raw.githubusercontent.com/JustHumanz/Go-simp/master/Img/404.jpg"
						msg = "Image or Video oversize,check original post"
					}
					TwitterData := PushData{
						Twitter: database.InputTW{
							Url:      "https://twitter.com/" + Data.User.ScreenName + "/status/" + Data.IDStr,
							Author:   Data.User.Name,
							Like:     Data.FavoriteCount,
							Photos:   strings.Join(Photos, "\n"),
							Video:    Video,
							Text:     Data.Text,
							TweetID:  Data.IDStr,
							MemberID: Group[i].MemberID,
						},
						Image:      SendMedia,
						Msg:        msg,
						ScreenName: Data.User.ScreenName,
						UserName:   Data.User.Name,
						Text:       RemoveTwitterShotlink(Data.Text),
						Avatar:     (strings.Replace(Data.User.ProfileImageURLHTTPS, "_normal.jpg", ".jpg", -1)),
						Group:      Group[i],
					}
					TwitterData.Twitter.InputTwitter()
					TwitterData.SendNude()
				}
			}
		}
	}
}