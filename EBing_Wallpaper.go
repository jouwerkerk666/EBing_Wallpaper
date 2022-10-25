package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"image"
	_ "image/jpeg"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/antonholmquist/jason"
	"github.com/pborman/getopt/v2"
)

var (
	Help, Verbose, Version      bool = false, false, false
	Skip, Keep, Quiet           bool = false, false, false
	Daemon                      bool = false
	Past                        int  = 0
	IDX                         string
	BingWallPaperUrl            string
	Title, Copyright, StartDate string
)

const (
	EBingVersion = "1.0-1"
	ShellToUse   = "bash"
)

func GetBingWallpaper(url string) []byte {
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	client := &http.Client{Transport: customTransport}

	res, err := client.Get(url)
	if err != nil {
		fmt.Println(err.Error())
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err.Error())
	}
	return body
}

func GetBingPicture(srcUrl string, destFile string) (int, int) {
	var (
		Xsize, Ysize int
	)
	fileName := destFile
	// Create blank file
	file, err := os.Create(fileName)
	if err != nil {
		log.Fatal(err)
	}
	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}
	resp, err := client.Get(srcUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	size, _ := io.Copy(file, resp.Body)
	defer file.Close()
	reader, err := os.Open(fileName)
	if err == nil {
		defer reader.Close()
		im, _, _ := image.DecodeConfig(reader)
		Xsize = im.Width
		Ysize = im.Height
	} else {
		fmt.Println("Error Opening file ", err)
	}
	if Verbose {
		fmt.Printf("Downloaded a picture %s with size %d Geometry:(%v:%v)\n\n", fileName, size, Xsize, Ysize)
	}
	return Xsize, Ysize
}

func Shellout(command string) (string, string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(ShellToUse, "-c", command)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func init() {
	getopt.FlagLong(&Help, "help", 'h', "Display this help")
	getopt.FlagLong(&Version, "version", 'V', "Display version")
	getopt.FlagLong(&Verbose, "verbose", 'v', "Verbose output")
	getopt.FlagLong(&Skip, "skip", 's', "Skip Enlightenment desktop check")
	getopt.FlagLong(&Keep, "keep", 'k', "Keep the tmp files, do not discard them.")
	getopt.FlagLong(&Past, "past", 'p', "is the previous [n] wallpaper from Bing. Max 7 days")
	getopt.FlagLong(&Quiet, "quiet", 'q', "No output")
	getopt.FlagLong(&Daemon, "daemon", 'd', "Run in the background every day at 09:00.")
	getopt.Parse()
	err := getopt.Getopt(nil)
	if err != nil || Help {
		if !Help {
			fmt.Fprintln(os.Stderr, err)
		}
		getopt.Usage()
		os.Exit(1)
	}
	if Quiet && Verbose {
		fmt.Println("Error, can not be Quiet and Verbose at the same time")
		os.Exit(1)
	}
	if Past > 7 {
		if !Quiet {
			fmt.Println("The max for past is 7.")
		}
		Past = 7
	}

	if Version {
		fmt.Printf("Version: %v\n", EBingVersion)
		os.Exit(0)
	}
}

func waitUntil(ctx context.Context, until time.Time) {
	timer := time.NewTimer(time.Until(until))
	defer timer.Stop()

	select {
	case <-timer.C:
		return
	case <-ctx.Done():
		return
	}
}

func main() {
	var (
		aspect      float32
		out, errout string
		cmderr      error
	)

loop:
	ctx := context.Background()
	today := time.Now()
	tomorrow := today.AddDate(0, 0, 1)
	NextRun := fmt.Sprintf("%04d-%02d-%02dT09:00:00+02:00", tomorrow.Year(), tomorrow.Month(), tomorrow.Day())
	until, _ := time.Parse(time.RFC3339, NextRun)
	DeskTop := string(os.Getenv("DESKTOP"))
	if DeskTop != "Enlightenment" && !Skip {
		fmt.Printf("Your desktop is not Enlightenment, so this is not working.\n")
		os.Exit(1)
	}

	if !Quiet {
		fmt.Printf("Change wallpaper in Enlightenment\n")
	}
	if Verbose {
		fmt.Println("Verbose output")
	}

	DestDir := fmt.Sprintf("%v/.e/e/backgrounds", os.Getenv("HOME"))
	TempDir := fmt.Sprintf("/tmp/%v/EBing_Wallpaper", os.Getenv("USER"))
	//TempDir := fmt.Sprintf("%v/Pictures/BingWallpapers", os.Getenv("HOME"))
	// Creating the directories if not exists
	err := os.MkdirAll(DestDir, os.ModePerm)
	if err != nil {
		log.Println(err)
	}
	err = os.MkdirAll(TempDir, os.ModePerm)
	if err != nil {
		log.Println(err)
	}

	IDX = getopt.GetValue("past")
	BingUrl := fmt.Sprintf("https://www.bing.com/HPImageArchive.aspx?format=js&idx=%v&n=1&mkt=en-US", IDX)
	BingJSON := GetBingWallpaper(BingUrl)
	BingWallPaperJSON, _ := jason.NewObjectFromBytes(BingJSON)
	Images, _ := BingWallPaperJSON.GetObjectArray("images")
	for _, Image := range Images {
		WallPaperUrl, _ := Image.GetString("url")
		WPStartDate, _ := Image.GetString("startdate")
		WPTitle, _ := Image.GetString("title")
		WPCopyright, _ := Image.GetString("copyright")
		BingWallPaperUrl = fmt.Sprintf("https://www.bing.com%v", WallPaperUrl)
		StartDate = fmt.Sprintf("%v", WPStartDate)
		Title = strings.ReplaceAll(WPTitle, " ", "_")
		Copyright = fmt.Sprintf("%v", WPCopyright)
	}
	ImageFile := fmt.Sprintf("%v/%v-%v.jpg", TempDir, StartDate, Title)
	EdjeFile := fmt.Sprintf("%v/bing_wallpaper_%v.edc", TempDir, StartDate)

	Xsize, Ysize := GetBingPicture(BingWallPaperUrl, ImageFile)
	if (Xsize != 0) && (Ysize != 0) {
		aspect = float32(Xsize) / float32(Ysize)
	} else {
		aspect = 0.0
	}

	EFLTemplate := fmt.Sprintf("\nimages { image: \"%v\" COMP; }\ncollections {\n  group {\n  name: \"e/desktop/background\";\n  data { item: \"style\" \"4\"; item: \"noanimation\" \"1\"; }\n  max: %v %v;\n  parts {\n    part {\n    name: \"bg\";\n    mouse_events: 0;\n    description {\n      state: \"default\" 0.0;\n      aspect: %.9f %.9f;\n      aspect_preference: NONE;\n      image { normal: \"%v\"; scale_hint: STATIC; }\n    }\n    }\n  }\n  }\n}\n", ImageFile, Xsize, Ysize, aspect, aspect, ImageFile)
	err = os.WriteFile(EdjeFile, []byte(EFLTemplate), 0755)
	if err != nil {
		log.Fatalf("Unable to open file %v for writing\n", EdjeFile)
	}
	BingWallPaperFile := fmt.Sprintf("%v/bing_wallpaper_%v.edj", DestDir, StartDate)
	edje_cc := fmt.Sprintf("edje_cc %v %v", EdjeFile, BingWallPaperFile)
	out, errout, cmderr = Shellout(edje_cc)
	if cmderr != nil {
		log.Fatalf("error: %v\n", cmderr)
	}

	if Verbose {
		fmt.Printf("StartDate        : %v\nPicture Title    : %v\nCopyright notice: %v \n", StartDate, Title, Copyright)
		fmt.Printf("edje_cc command  : %v \n", edje_cc)
		fmt.Println("--- edje_cc output ---")
		fmt.Println(out)
		fmt.Println(errout)
		fmt.Println("-----------------")
	}
	for x := 0; x <= 2; x++ {
		command := fmt.Sprintf("enlightenment_remote -desktop-bg-add -1 %v -1 -1 %v", x, BingWallPaperFile)
		cmd := exec.Command("/bin/bash", "-c", command)
		cmderr := cmd.Run()
		if cmderr != nil {
			log.Fatalf("error: %v\n", cmderr)
		}
		if Verbose {
			fmt.Printf("%v\n", command)
		}
	}
	if !Keep {
		/*
		 * Remove the edc file
		 */
		err := os.RemoveAll(TempDir)
		if err != nil {
			log.Fatal(err)
		}
	}
	if Daemon {
		waitUntil(ctx, until)
		goto loop
	}
}
