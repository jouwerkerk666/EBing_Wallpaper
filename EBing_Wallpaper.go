package main

import (
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
	Help, Verbose, Version, Noop bool = false, false, false, false
	Last                         bool = false
	Past                         int  = 1
	BingWallPaperUrl             string
	Title, Copyright, StartDate  string
)

const (
	EBingVersion = "0.0.1-beta"
)

func UNUSED(x ...interface{}) {}

func GetBingWallpaper(url string) []byte {
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	client := &http.Client{Transport: customTransport}
	//os.Setenv("HTTPS_PROXY", "http://host.docker.internal:3128")

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

func init() {
	getopt.FlagLong(&Help, "help", 'h', "Display help")
	getopt.FlagLong(&Version, "version", 'V', "Display version")
	getopt.FlagLong(&Verbose, "verbose", 'v', "Verbose output")
	getopt.FlagLong(&Noop, "noop", 'n', "Do not Execute. Only shows what will be done.")
	getopt.FlagLong(&Last, "last", 'l', "Get the latest wallpaper from Bing.")
	getopt.FlagLong(&Past, "past", 'p', "is the previous [n] wallpaper from Bing.")
	err := getopt.Getopt(nil)
	if err != nil || Help {
		if !Help {
			fmt.Fprintln(os.Stderr, err)
		}
		getopt.Usage()
		os.Exit(1)
	}
	if Version {
		fmt.Printf("Version: %v\n", EBingVersion)
		os.Exit(1)
	}
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
		fmt.Printf("DEBUG: Downloaded a file %s with size %d, (%v:%v)\n\n", fileName, size, Xsize, Ysize)
	}
	return Xsize, Ysize

}

func main() {
	var (
		aspect float32
	)
	DeskTop := string(os.Getenv("DESKTOP"))
	if DeskTop != "ENLIGHTENMENT" && !Noop {
		fmt.Printf("Your desktop is not Enlightenment, so this is not working.\n")
		os.Exit(1)
	}
	fmt.Printf("Change wallpaper in Enlightenment\n")
	BingJSON := GetBingWallpaper("https://www.bing.com/HPImageArchive.aspx?format=js&idx=0&n=1&mkt=en-US")
	BingWallPaperJSON, _ := jason.NewObjectFromBytes(BingJSON)
	Images, _ := BingWallPaperJSON.GetObjectArray("images")
	for _, Image := range Images {
		WallPaperUrl, _ := Image.GetString("url")
		WPStartDate, _ := Image.GetString("startdate")
		WPTitle, _ := Image.GetString("title")
		WPCopyright, _ := Image.GetString("copyright")

		BingWallPaperUrl = fmt.Sprintf("https://www.bing.com%v", WallPaperUrl)
		StartDate = fmt.Sprintf("%v", WPStartDate)
		//Title = fmt.Sprintf("%v", WPTitle)
		Title = strings.ReplaceAll(WPTitle, " ", "_")
		Copyright = fmt.Sprintf("%v", WPCopyright)
	}
	TempDir := fmt.Sprintf("/tmp/%v/BingWallPaper", os.Getenv("USER"))
	DestDir := fmt.Sprintf("%v/.e/e/backgrounds", os.Getenv("HOME"))
	ImageFile := fmt.Sprintf("%v/%v-%v.jpg", TempDir, StartDate, Title)
	EdjeFile := strings.ReplaceAll(ImageFile, ".jpg", ".edc")

	// Creating the directories
	err := os.MkdirAll(DestDir, os.ModePerm)
	if err != nil {
		log.Println(err)
	}
	err = os.MkdirAll(TempDir, os.ModePerm)
	if err != nil {
		log.Println(err)
	}

	Xsize, Ysize := GetBingPicture(BingWallPaperUrl, ImageFile)
	if (Xsize != 0) && (Ysize != 0) {
		aspect = float32(Xsize) / float32(Ysize)
	} else {
		aspect = 0.0
	}

	EFLTemplate := fmt.Sprintf("images { image: \"%v\" USER; }\n	collections {\n	  group {\n	  name: \"e/desktop/background\";\n 	  data { item: \"style\" \"4\"; item: \"noanimation\" \"1\"; }\n 	  max: %v %v;\n 	  parts {\n 		part {\n 		name: \"bg\";\n 		mouse_events: 0;\n		description {\n		  state: \"default\" 0.0;\n		  aspect: %.1f %.1f;\n		  aspect_preference: NONE;\n		  image { normal: \"%v\"; scale_hint: STATIC; }\n 		}\n		}\n	  }\n	  }\n	}\n	", ImageFile, Xsize, Ysize, aspect, aspect, ImageFile)
	//	edje_cc_filename = fmt.Sprintf("/tmp/%v/.edc"
	err = os.WriteFile(EdjeFile, []byte(EFLTemplate), 0755)
	if err != nil {
		fmt.Printf("Unable to open file %v for writing\n", EdjeFile)
	}
	BingWallPaperFile := fmt.Sprintf("%v/.e/e/background/bing_wallpaper_%v", os.Getenv("$HOME"), StartDate)
	edje_cc := fmt.Sprintf("edje_cc \"%v\" \"%v\"", BingWallPaperFile, EdjeFile)
	if !Noop {
		cmd := exec.Command("/bin/bash", "-c", edje_cc)
		cmderr := cmd.Run()
		if cmderr != nil {
			log.Printf("error: %v\n", cmderr)
		}
	}
	if Verbose {
		fmt.Printf("DEBUG:\n %v \n %v \n %v \n %v \n", BingWallPaperUrl, StartDate, Title, Copyright)
		fmt.Printf("DEBUG:\n %v\n\n", EFLTemplate)
		fmt.Printf("DEGUG:\n %v \n", edje_cc)
		fmt.Printf("DEBUG: Aspect = %.1f\n", aspect)
	}
	for x := 0; x <= 2; x++ {
		command := fmt.Sprintf("enlightenment_remote -desktop-bg-add -1 %v -1 -1 %v.edj", x, BingWallPaperFile)
		if !Noop {
			cmd := exec.Command("/bin/bash", "-c", command)
			cmderr := cmd.Run()
			if cmderr != nil {
				log.Printf("error: %v\n", cmderr)
			}
		}
		if Verbose {
			fmt.Printf("DEBUG: %v\n", command)
		}
	}
	if !Noop {
		rdir := fmt.Sprintf("/tmp/%v", os.Getenv("$USER"))
		err := os.RemoveAll(rdir)
		if err != nil {
			log.Fatal(err)
		}
	}
}
