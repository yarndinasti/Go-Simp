package pixiv

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/JustHumanz/Go-Simp/pkg/config"
	"github.com/JustHumanz/Go-Simp/pkg/database"
	"github.com/JustHumanz/Go-Simp/pkg/engine"
	"github.com/JustHumanz/Go-Simp/pkg/network"
	"github.com/JustHumanz/Go-Simp/service/fanart/notif"
	"github.com/bwmarrin/discordgo"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
)

//Public variable
var (
	Bot         *discordgo.Session
	VtubersData database.VtubersPayload
	configfile  config.ConfigFile
)

const (
	BaseURL = "https://www.pixiv.net/en/artworks/"
)

//Start start twitter module
func Start(a *discordgo.Session, b *cron.Cron, c database.VtubersPayload, d config.ConfigFile) {
	Bot = a
	VtubersData = c
	configfile = d
	b.AddFunc(config.PixivFanart, CheckPixiv)
	log.Info("Enable Pixiv fanart module")
}

//CheckNew Check new fanart
func CheckPixiv() {
	wg := new(sync.WaitGroup)
	for _, GroupData := range VtubersData.VtuberData {
		wg.Add(1)
		go func(Group database.Group, wg *sync.WaitGroup) {
			defer wg.Done()
			for _, Member := range Group.Members {
				Pixiv := func(Payload string) error {
					var Art PixivArtworks
					log.WithFields(log.Fields{
						"Member": Member.EnName,
						"Group":  Group.GroupName,
					}).Info("Start curl pixiv")
					body, err := network.Curl(Payload, nil)
					if err != nil {
						return err
					}
					err = json.Unmarshal(body, &Art)
					if err != nil {
						return err
					}
					IsVtuber := false
					for _, tag := range Art.Body.Relatedtags {
						if strings.ToLower(tag.(string)) == strings.ToLower(Group.GroupName) {
							IsVtuber = true
						}
					}

					if Art.Body.Illustmanga.Data != nil && IsVtuber {
						for i, v := range Art.Body.Illustmanga.Data {
							var (
								v2      = v.(map[string]interface{})
								Illusts map[string]interface{}
								User    map[string]interface{}
								TextFix string
							)

							if v2["xRestrict"].(float64) == 0 {
								illusbyte, err := network.Curl(config.PixivIllustsEnd+v2["id"].(string), nil)
								if err != nil {
									return err
								}

								err = json.Unmarshal(illusbyte, &Illusts)
								if err != nil {
									return err
								}

								Body := Illusts["body"].(map[string]interface{})
								Tags := Body["tags"].(map[string]interface{})
								Img := Body["urls"].(map[string]interface{})
								FixImg := Img["original"].(string)

								usrbyte, err := network.Curl(config.PixivUserEnd+Tags["authorId"].(string), nil)
								if err != nil {
									return err
								}

								err = json.Unmarshal(usrbyte, &User)
								if err != nil {
									return err
								}

								UserBody := User["body"].(map[string]interface{})

								Desc := RemoveHtmlTag(Body["description"].(string))
								if match, _ := regexp.MatchString("http://twitter.com", Desc); match {
									TextFix = ClearTwitterURL(Desc)
								} else {
									TextFix = "**" + Body["title"].(string) + "**\n" + Desc
								}

								FixFanArt := database.DataFanart{
									PermanentURL: BaseURL + v2["id"].(string),
									Author:       v2["userName"].(string),
									AuthorAvatar: config.PixivProxy + UserBody["imageBig"].(string),
									Photos:       []string{FixImg},
									Text:         TextFix,
									PixivID:      v2["id"].(string),
									Member:       Member,
									State:        config.PixivArt,
								}

								new, err := FixFanArt.CheckPixivFanArt()
								if err != nil {
									return err
								}
								if new {
									path, err := DownloadImg(Img["mini"].(string))
									if err != nil {
										log.Error(err)
									}

									Color, err := engine.GetColor("", path)
									if err != nil {
										return err
									}
									FixFanArt.Photos[0] = config.PixivProxy + FixImg
									notif.SendNude(FixFanArt, Group, Bot, Color)
								}
							}
							if i == 10 {
								break
							}
						}
					}
					return nil
				}
				if Member.Region == "JP" {
					URL := GetPixivURL(strings.Replace(Member.JpName, " ", "_", -1))
					err := Pixiv(URL)
					if err != nil {
						log.Error(err)
					}
				} else {
					if Member.EnName != "" {
						URL := GetPixivURL(strings.Replace(Member.EnName, " ", "_", -1))
						err := Pixiv(URL)
						if err != nil {
							log.Error(err)
						}
					} else {
						URL := GetPixivURL(strings.Replace(Member.Name, " ", "_", -1))
						err := Pixiv(URL)
						if err != nil {
							log.Error(err)
						}
					}
				}
			}
		}(GroupData, wg)
	}
	wg.Wait()
}

func GetPixivURL(str string) string {
	return "https://www.pixiv.net/ajax/search/artworks/" + str + "?word=" + str + "&order=date_d&mode=all&p=1&s_mode=s_tag_full&type=all&lang=en"
}

func ClearTwitterURL(str1 string) string {
	re := regexp.MustCompile(`(https\:\/\/twitter\.com\/.*)\<`)
	submatchall := re.FindStringSubmatch(str1)
	fix := ""
	for _, element := range submatchall {
		fix = element
	}
	return fix
}

func RemoveHtmlTag(in string) string {
	// regex to match html tag
	const pattern = `(<\/?[a-zA-A]+?[^>]*\/?>)*`
	r := regexp.MustCompile(pattern)
	groups := r.FindAllString(in, -1)
	// should replace long string first
	sort.Slice(groups, func(i, j int) bool {
		return len(groups[i]) > len(groups[j])
	})
	for _, group := range groups {
		if strings.TrimSpace(group) != "" {
			in = strings.ReplaceAll(in, group, "")
		}
	}
	return in
}

func DownloadImg(u string) (string, error) {
	dir := config.TmpDir + engine.RanString()
	out, err := os.Create(dir)
	if err != nil {
		return "", err
	}

	defer out.Close()
	request, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	request.Header.Set("cache-control", "no-cache")
	request.Header.Set("User-Agent", network.RandomAgent())
	request.Header.Set("Referer", "https://www.pixiv.net")

	spaceClient := http.Client{}
	resp, err := spaceClient.Do(request.WithContext(ctx))
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New(resp.Status)
	}
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}
	return dir, nil
}