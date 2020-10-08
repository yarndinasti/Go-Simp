package bilibili

import (
	"strconv"
	"strings"
	"sync"

	database "github.com/JustHumanz/Go-simp/database"
	engine "github.com/JustHumanz/Go-simp/engine"

	log "github.com/sirupsen/logrus"
)

type Notif struct {
	TBiliData   database.InputTBiliBili
	Group       database.GroupName
	PhotosImgur string
	PhotosCount int
	MemberID    int64
}

//Push Data to discord channel
func (NotifData Notif) PushNotif(Color int) {
	Data := NotifData.TBiliData
	Group := NotifData.Group
	wg := new(sync.WaitGroup)
	ID, DiscordChannelID := database.ChannelTag(NotifData.MemberID, 1)
	for i := 0; i < len(DiscordChannelID); i++ {
		UserTagsList := database.GetUserList(ID[i], NotifData.MemberID)
		wg.Add(1)
		go func(DiscordChannel string, wg *sync.WaitGroup) {
			defer wg.Done()

			msg := ""
			repost, url, err := engine.SaucenaoCheck(strings.Split(Data.Photos, "\n")[0])
			if err != nil {
				log.Error(err)
				msg = "??????"
			} else if repost && url != nil {
				log.WithFields(log.Fields{
					"Source Img": Data.URL,
					"Sauce Img":  url,
				}).Info("Repost")
				msg = url[0]
			} else {
				log.WithFields(log.Fields{
					"Source Img": Data.URL,
					"Sauce Img":  url,
				}).Info("Ntap,Anyar cok")
				msg = "_"
			}
			if UserTagsList != nil {
				Embed := engine.NewEmbed().
					SetAuthor(strings.Title(Group.NameGroup), Group.IconURL).
					SetTitle(Data.Author).
					SetURL(Data.URL).
					SetThumbnail(Data.Avatar).
					SetDescription(Data.Text).
					SetImage(NotifData.PhotosImgur).
					AddField("User Tags", strings.Join(UserTagsList, " ")).
					AddField("Similar art", msg).
					SetFooter("1/"+strconv.Itoa(NotifData.PhotosCount)+" photos", "https://raw.githubusercontent.com/JustHumanz/Go-simp/master/Img/bilibili.png").
					InlineAllFields().
					SetColor(Color).MessageEmbed
				msg, err := BotSession.ChannelMessageSendEmbed(DiscordChannel, Embed)
				if err != nil {
					log.Error(msg, err)
				}
				err = engine.Reacting(map[string]string{
					"ChannelID": DiscordChannel,
				})
				if err != nil {
					log.Error(err)
				}
			} else {
				//DropTheBom("_", msg)
			}
		}(DiscordChannelID[i], wg)
	}
	wg.Wait()
}