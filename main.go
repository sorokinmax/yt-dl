package main

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/kkdai/youtube"
)

const version = "v.1.0.0"

const DEBUG = false

func main() {
	fmt.Printf("yt-dl %s\n", version)
	var videoID string

	if !DEBUG {
		if len(os.Args[1:]) != 1 {
			fmt.Print("Usage:", os.Args[0], "VideoID_or_URL")
			return
		}
		videoID = os.Args[1]

		if strings.HasPrefix(videoID, "https://") {

			u, err := url.Parse(videoID)
			if err != nil {
				panic(err)
			}

			q := u.Query()
			videoID = q.Get("v")
		}

		if videoID == "" {
			fmt.Print("Video ID is required")
			return
		}
	} else {
		videoID = "QH2-TGUlwu4"
	}

	client := youtube.Client{}

	video, err := client.GetVideo(videoID)
	if err != nil {
		panic(err)
	}

	i := 0
	for _, itag := range video.Formats {
		if itag.AudioQuality != "" {
			if itag.QualityLabel != "" {
				fmt.Printf("%d: [VA] VIDEO(%s - %s) AUDIO(%s - %sHz) FORMAT(%s)\n", i, itag.Quality, itag.QualityLabel, itag.AudioQuality, itag.AudioSampleRate, itag.MimeType)
			} else {
				fmt.Printf("%d: [A] AUDIO(%s - %sHz) FORMAT(%s)\n", i, itag.AudioQuality, itag.AudioSampleRate, itag.MimeType)
			}
		} else {
			fmt.Printf("%d: [V] VIDEO(%s - %s - %skbps) FORMAT(%s)\n", i, itag.Quality, itag.QualityLabel, strconv.Itoa(itag.Bitrate/1024), itag.MimeType)
		}
		i++
	}

	fmt.Print("Choose format -> ")

	var input string
	fmt.Scanln(&input)

	var vi, _ = strconv.Atoi(input)
	resp, err := client.GetStream(video, &video.Formats[vi])
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	filename := removeCharacters(video.Title, "<>:\"/\\|?*")
	extension := strings.FieldsFunc(video.Formats[13].MimeType, Split)[1]
	filename += "." + extension
	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	counter := &WriteCounter{}
	src := io.TeeReader(resp.Body, counter)

	count, err := io.Copy(file, src)
	if err != nil {
		panic(err)
	}

	fmt.Println("Saved", count, "bytes")

}

func removeCharacters(input string, characters string) string {
	filter := func(r rune) rune {
		if strings.IndexRune(characters, r) < 0 {
			return r
		}
		return -1
	}

	return strings.Map(filter, input)

}

func Split(r rune) bool {
	return r == '/' || r == ';'
}

// WriteCounter counts the number of bytes written to it.
type WriteCounter struct {
	Total int64 // Total # of bytes written
}

// Write implements the io.Writer interface.
// Always completes and never returns an error.
func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += int64(n)
	fmt.Printf("Read %d bytes for a total of %d\n", n, wc.Total)
	return n, nil
}
