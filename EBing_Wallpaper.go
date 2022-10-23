package main

import (
	"bytes"
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

	"github.com/antonholmquist/jason"
	"github.com/pborman/getopt/v2"
)

var (
	Help, Verbose, Version, Noop, Keep bool = false, false, false, false, false
	Past                               int  = 0
	IDX                                string
	BingWallPaperUrl                   string
	Title, Copyright, StartDate        string
)

const (
	EBingVersion = "1.0"
	ShellToUse   = "bash"
)

func UNUSED(x ...interface{}) {}

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
	// Put content on file
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
		fmt.Printf("Downloaded a file %s with size %d Geometry:(%v:%v)\n\n", fileName, size, Xsize, Ysize)
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
	getopt.FlagLong(&Help, "help", 'h', "Display help")
	getopt.FlagLong(&Version, "version", 'V', "Display version")
	getopt.FlagLong(&Verbose, "verbose", 'v', "Verbose output")
	getopt.FlagLong(&Noop, "noop", 'n', "Do not Execute. Only shows what will be done.")
	getopt.FlagLong(&Keep, "keep", 'k', "Keep the tmp files, do not discard them.")
	getopt.FlagLong(&Past, "past", 'p', "is the previous [n] wallpaper from Bing. Max 7 days")
	getopt.Parse()
	err := getopt.Getopt(nil)
	if err != nil || Help {
		if !Help {
			fmt.Fprintln(os.Stderr, err)
		}
		getopt.Usage()
		os.Exit(1)
	}
	if Past > 7 {
		fmt.Println("The max for past is 7.")
		Past = 7
	}

	if Version {
		fmt.Printf("Version: %v\n", EBingVersion)
		os.Exit(1)
	}
}

func main() {
	var (
		aspect      float32
		out, errout string
		cmderr      error
	)

	DeskTop := string(os.Getenv("DESKTOP"))
	if DeskTop != "Enlightenment" && !Noop {
		fmt.Printf("Your desktop is not Enlightenment, so this is not working.\n")
		os.Exit(1)
	}
	fmt.Printf("Change wallpaper in Enlightenment\n")

	DestDir := fmt.Sprintf("%v/.e/e/backgrounds", os.Getenv("HOME"))
	TempDir := fmt.Sprintf("%v/Pictures/BingWallpapers", os.Getenv("HOME"))
	// Creating the directories
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
	EdjeFile := fmt.Sprintf("%v/bing_wallpaper_%v.edc", DestDir, StartDate)

	Xsize, Ysize := GetBingPicture(BingWallPaperUrl, ImageFile)
	if (Xsize != 0) && (Ysize != 0) {
		aspect = float32(Xsize) / float32(Ysize)
	} else {
		aspect = 0.0
	}

	EFLTemplate := fmt.Sprintf("\nimages { image: \"%v\" USER; }\ncollections {\n  group {\n  name: \"e/desktop/background\";\n  data { item: \"style\" \"4\"; item: \"noanimation\" \"1\"; }\n  max: %v %v;\n  parts {\n    part {\n    name: \"bg\";\n    mouse_events: 0;\n    description {\n      state: \"default\" 0.0;\n      aspect: %.9f %.9f;\n      aspect_preference: NONE;\n      image { normal: \"%v\"; scale_hint: STATIC; }\n    }\n    }\n  }\n  }\n}\n", ImageFile, Xsize, Ysize, aspect, aspect, ImageFile)
	err = os.WriteFile(EdjeFile, []byte(EFLTemplate), 0755)
	if err != nil {
		fmt.Printf("Unable to open file %v for writing\n", EdjeFile)
	}
	BingWallPaperFile := fmt.Sprintf("%v/bing_wallpaper_%v.edj", DestDir, StartDate)
	edje_cc := fmt.Sprintf("edje_cc %v %v", EdjeFile, BingWallPaperFile)
	if !Noop {
		out, errout, cmderr = Shellout(edje_cc)
		if cmderr != nil {
			log.Printf("error: %v\n", cmderr)
		}
	}

	if Verbose {
		fmt.Println("Verbose output")
		fmt.Println("--- edje_cc stdout ---")
		fmt.Println(out)
		fmt.Println("--- edje_cc stderr ---")
		fmt.Println(errout)
		fmt.Printf("%v \n %v \n %v \n", StartDate, Title, Copyright)
		fmt.Printf("%v \n", edje_cc)
		fmt.Printf("Aspect = %.1f\n", aspect)
	}
	for x := 0; x <= 2; x++ {
		command := fmt.Sprintf("enlightenment_remote -desktop-bg-add -1 %v -1 -1 %v", x, BingWallPaperFile)
		if !Noop {
			cmd := exec.Command("/bin/bash", "-c", command)
			cmderr := cmd.Run()
			if cmderr != nil {
				log.Printf("error: %v\n", cmderr)
			}
		}
		if Verbose {
			fmt.Printf("%v\n", command)
		}
	}
	if !Keep {
		err := os.Remove(EdjeFile)
		if err != nil {
			log.Fatal(err)
		}
		err = os.Remove(ImageFile)
		if err != nil {
			log.Fatal(err)
		}
	}
}
