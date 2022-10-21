package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

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
	EBingVersion = "0.0.0-alpha1"
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

func main() {
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
		Title = fmt.Sprintf("%v", WPTitle)
		Copyright = fmt.Sprintf("%v", WPCopyright)
	}
	TempDir := fmt.Sprintf("/tmp/%v/BingWallPaper", os.Getenv("USER"))
	DestDir := fmt.Sprintf("%v/.e/e/backgrounds", os.Getenv("HOME"))
	ImageFile := fmt.Sprintf("%v/%v-%v.jpg", TempDir, StartDate, Title)
	aspect := "0.0"

	EFLTemplate := fmt.Sprintf("images { image: \"%v\" USER; }\n	collections {\n	  group {\n	  name: \"e/desktop/background\"; 	  data { item: \"style\" \"4\"; item: \"noanimation\" \"1\"; }\n 	  max: %v %v;\n 	  parts {\n 		part {\n 		name: \"bg\";\n 		mouse_events: 0;\n		description {\n		  state: \"default\" 0.0;\n		  aspect: %v %v;\n		  aspect_preference: NONE;\n		  image { normal: \"%v\"; scale_hint: STATIC; }\n 		}\n		}\n	  }\n	  }\n	}\n	", ImageFile, "1900", "1280", ImageFile, aspect, aspect)
	UNUSED(DestDir, EFLTemplate)
	err := os.MkdirAll(DestDir, os.ModePerm)
	if err != nil {
		log.Println(err)
	}
	err = os.MkdirAll(TempDir, os.ModePerm)
	if err != nil {
		log.Println(err)
	}
	if Verbose {
		fmt.Printf("DEBUG:\n %v \n %v \n %v \n %v \n", BingWallPaperUrl, StartDate, Title, Copyright)
		fmt.Printf("DEBUG:\n %v", EFLTemplate)
	}
}
