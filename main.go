package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	var (
		input  string
		output string
		ini    string
	)

	flag.StringVar(&input, "i", "", "Input")
	flag.StringVar(&output, "o", "", "Output")
	flag.StringVar(&ini, "c", "", "comskip's ini file path")
	flag.Parse()

	base := output[:len(output)-len(filepath.Ext(output))]

	ch := make(chan error, 2)

	defer exec.Command("rm", base+".vdr", base+".chp", base+".txt", base+".logo.txt", base+"_chap.mp4").Run()

	log.Println("Start encoding and detecting commercials")
	go func() {
		ch <- exec.Command("ffmpeg", "-fflags", "+discardcorrupt", "-i", input, "-c:a", "copy", "-bsf:a", "aac_adtstoasc", "-c:v", "h264_omx", "-b:v", "5000k", "-y", output).Run()
	}()
	go func() {
		ch <- exec.Command("comskip", "-d", "255", "--ini="+ini, "--threads="+strconv.Itoa(runtime.NumCPU()), "--hwassist", "-t", input).Run()
	}()
	var failed bool
	for i := 0; i < 2; i++ {
		err := <-ch
		if err != nil {
			failed = true
		}
	}
	if failed {
		panic(errors.New("Some error occurred during encoding or detecting commercials."))
	}

	log.Println("Making chp file")
	if err := makeChapterFile(base+".vdr", base+".chp"); err != nil {
		panic(err)
	}

	log.Println("Integrating chapter")
	if err := exec.Command("MP4Box", "-chap", base+".chp", "-out", base+"_chap.mp4", output).Run(); err != nil {
		panic(err)
	}
	if err := exec.Command("mv", base+"_chap.mp4", output).Run(); err != nil {
		panic(err)
	}
}

func makeChapterFile(input, output string) error {
	b, err := ioutil.ReadFile(input)
	if err != nil {
		return err
	}
	chapterInfo := string(b)

	fOut, err := os.Create(output)
	if err != nil {
		return err
	}
	defer fOut.Close()
	timings := strings.Split(chapterInfo, "\n")

	for i, line := range timings[:len(timings)-1] {
		words := strings.Fields(line)
		chapNum := fmt.Sprintf("%02d", i+1)
		fOut.WriteString("CHAPTER" + chapNum + "=" + words[0] + "\n")
		fOut.WriteString("CHAPTER" + chapNum + "NAME=chap" + chapNum + "\n")
	}
	return nil
}
