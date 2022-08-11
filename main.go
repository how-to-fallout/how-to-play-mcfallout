package main

import (
	"context"
	"fmt"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/google/go-github/v45/github"
	"golang.org/x/exp/slices"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
)

func main() {
	var blockList []string
	s := state.New("Bot " + os.Getenv("ACTION_BOT_TOKEN"))
	PRNumber, err := strconv.Atoi(os.Getenv("PR_NUMBER"))
	if err != nil {
		log.Fatalln(err)
	}
	s.AddIntents(32767)
	err = s.Open(context.Background())
	if err != nil {
		log.Fatalln(err)
		return
	}
	client := github.NewClient(nil)
	commits, _, err := client.PullRequests.ListCommits(context.Background(), "how-to-fallout", "how-to-play-mcfallout", PRNumber, nil)
	if err != nil {
		return
	}
	pr, _, err := client.PullRequests.Get(context.Background(), "how-to-fallout", "how-to-play-mcfallout", PRNumber)
	if !pr.GetMerged() {
		return
	}
	for _, commit0 := range commits {
		commit, _, _ := client.Repositories.GetCommit(context.Background(), "how-to-fallout", "how-to-play-mcfallout", commit0.GetSHA(), nil)
		for _, file := range commit.Files {
			if file.GetStatus() == "removed" {
				continue
			}
			fmt.Println(file.GetFilename())
			compile := regexp.MustCompile("channels/(\\d+)/([\\S\\s]+)\\.txt").FindStringSubmatch(file.GetFilename())
			if compile == nil {
				return
			}
			if slices.Contains(blockList, compile[1]) {
				continue
			}
			if err != nil {
				return
			}
			_, directoryContent, _, _ := client.Repositories.GetContents(context.Background(), "how-to-fallout", "how-to-play-mcfallout", "channels/"+compile[1], &github.RepositoryContentGetOptions{})
			if directoryContent != nil {
				snowflake, _ := discord.ParseSnowflake(compile[1])
				channel, _ := s.Channel(discord.ChannelID(snowflake))
				messages, _ := s.Messages(channel.ID, 20)
				var ids []discord.MessageID
				for _, message := range messages {
					ids = append(ids, message.ID)
				}
				_ = s.DeleteMessages(channel.ID, ids, "Bulk Delete For Update")
				for _, content := range directoryContent {
					resp, _ := http.Get(content.GetDownloadURL())
					data, _ := io.ReadAll(resp.Body)
					d := discord.Embed{
						Title:       content.GetName(),
						Description: string(data),
						URL:         content.GetURL(),
					}
					_, _ = s.SendEmbeds(channel.ID, d)
				}
			}
			blockList = append(blockList, compile[1])
		}

	}
}
