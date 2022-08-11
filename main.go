package main

import (
	"context"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/bot/extras/infer"
	"github.com/google/go-github/v45/github"
	"golang.org/x/exp/slices"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	var blockList []string
	s := state.New("Bot " + os.Getenv("ACTION_BOT_TOKEN"))
	PRNumber, err := strconv.Atoi(os.Getenv("PR_NUMBER"))
	if err != nil {
		log.Fatalln(err)
	}
	s.AddIntents(32767)
	client := github.NewClient(nil)
	commits, _, err := client.PullRequests.ListCommits(context.Background(), "how-to-fallout", "how-to-play-mcfallout", PRNumber, nil)
	if err != nil {
		return
	}
	pr, _, err := client.PullRequests.Get(context.Background(), "how-to-fallout", "how-to-play-mcfallout", PRNumber)
	if !pr.GetMerged() {
		return
	}

	for i, commit := range commits {
		if i == 0 {
			for _, file := range commit.Files {
				if file.GetStatus() == "removed" {
					continue
				}
				if strings.HasPrefix(file.GetFilename(), "channels/") {
					continue
				}
				compile := regexp.MustCompile("channels/(\\d+)/([\\S\\s]+)\\.txt").FindStringSubmatch(file.GetFilename())
				if compile == nil {
					return
				}
				if slices.Contains(blockList, compile[1]) {
					continue
				}
				_, directoryContent, _, _ := client.Repositories.GetContents(context.Background(), "how-to-fallout", "how-to-play-mcfallout", "channels"+compile[1], &github.RepositoryContentGetOptions{})
				if directoryContent != nil {
					channel, _ := s.Channel(infer.ChannelID(compile[1]))
					messages, _ := s.Messages(channel.ID, 100)
					var ids []discord.MessageID
					for _, message := range messages {
						ids = append(ids, message.ID)
					}
					s.DeleteMessages(channel.ID, ids, "Bulk Delete For Update")
					for _, content := range directoryContent {
						resp, _ := http.Get(content.GetDownloadURL())
						data, _ := io.ReadAll(resp.Body)
						d := discord.Embed{
							Title:       content.GetName(),
							Description: string(data),
							URL:         content.GetURL(),
						}
						s.SendEmbeds(channel.ID, d)
					}
				}
				blockList = append(blockList, compile[1])
			}
		}

	}

}
