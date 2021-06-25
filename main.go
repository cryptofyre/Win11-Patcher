package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/urfave/cli"
)

var app = cli.NewApp()

func info() {
	app.Name = "Win11-Patcher"
	app.Usage = "Windows 11 TPM 2.0 and Secure Boot Setup.exe/Registry bypass written in Go."
	app.Author = "cryptofyre"
	app.Version = "2.0.0"
}

func commands() {
	app.Commands = []cli.Command{
		{
			Name:    "insiderpatch",
			Aliases: []string{"ins"},
			Usage:   "Changes Insider channel from Release Preview or other channel to Dev (Allows Windows 11 insider updates on incompatible devices.)",
			Action: func(c *cli.Context) {
				patchinsider()
			},
		},
		{
			Name:    "isopatch",
			Aliases: []string{"iso"},
			Usage:   "(Usage: win11-patcher.exe iso C:/Users/supbruv/Windows11.iso) Allows the upgrade to Windows 11 on Windows 10 and Windows 11 (older build) using the ISO Setup.exe on unsupported hardware.",
			Action: func(c *cli.Context) {
				emptystring := ""
				isoPath := strings.Join(c.Args(), " ")
				if isoPath == emptystring {
					fmt.Println("ERROR: Specify a path to a valid ISO. (Usage: win11-patcher.exe iso C:/Users/supbruv/Windows11.iso)")
				} else {
					patchiso(isoPath)
				}
			},
		},
		{
			Name:    "rmbuild",
			Aliases: []string{"rmb"},
			Usage:   "Remove build info in the bottom right from Windows 11 Insider Builds or Windows 10 Insider Builds.",
			Action: func(c *cli.Context) {
				removebuildinfo()
			},
		},
	}
}

func main() {
	if len(os.Args) > 1 {
		checkpath := os.Args[1]
		if strings.Contains(checkpath, ".iso") {
			patchiso(checkpath)
		} else {
			info()
			commands()
			err := app.Run(os.Args)
			if err != nil {
				log.Fatal(err)
			}
		}
	} else {
		info()
		commands()
		err := app.Run(os.Args)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func DownloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func RemoveDirectory(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

func patchinsider() {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	fileUrl := "https://cryptofyre.org/cdn/w11-insider-dev.reg"
	upgradePatch := exPath + "/w11-insider-dev.reg"
	clearterm()
	log.Printf("Insider Dev Patch initalizing.")
	log.Printf("If your not already in the Release Preview insider ring do that now. After your done press ENTER to continue.")
	fmt.Print("Press 'Enter' to continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
	log.Printf("Beginning registry patch...")
	log.Printf("Downloading registry patch from CDN.")
	errdl := DownloadFile(upgradePatch, fileUrl)
	if errdl != nil {
		panic(errdl)
	}
	log.Printf("Successfully downloaded patch(es) to " + upgradePatch)
	log.Printf("Installing and replacing registry key(s).")
	log.Printf("(If this process fails please relaunch Win11-Patcher with Administrator privilleges.)")
	regpatch := exec.Command("reg", "import", upgradePatch)
	regpatch.Stdout = os.Stdout
	regpatch.Stderr = os.Stderr
	regpatcherr := regpatch.Start()
	if regpatcherr != nil {
		log.Fatal(regpatcherr)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
	regpatcherr = regpatch.Wait()
	os.Remove(upgradePatch)
	log.Printf("Successfully imported registry key(s). Please reboot your computer then check the Insider settings in Updates & Security.")
	fmt.Print("Press 'Enter' to continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func patchiso(isopath string) {
	clearterm()
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)

	exSevenPath := exPath + "\\7z.exe"
	exSevendllPath := exPath + "\\7z.exe"

	log.Printf("Checking iso...")
	iso, err := os.Open(isopath)
	if _, err := os.Stat(exSevenPath); os.IsNotExist(err) {
		log.Println("Failed to find 7z.exe are they put alongside the executable?")
	}
	if _, err := os.Stat(exSevendllPath); os.IsNotExist(err) {
		log.Println("Failed to find 7z.dll are they put alongside the executable?")
	}
	if err != nil {
		log.Fatalf("Failed to open iso: %s", err)
	}
	defer iso.Close()
	log.Printf("Successfully checked ISO. Beginning extraction process.")

	homedir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	tempPath7z := "-o" + homedir + "/Win11-Temp/"
	tempPath := homedir + "/Win11-Temp/"

	extract := exec.Command(exSevenPath, "x", isopath, tempPath7z, "-y")
	extract.Stdout = os.Stdout
	extract.Stderr = os.Stderr
	extracterr := extract.Start()
	if extracterr != nil {
		log.Fatal(extracterr)
		log.Printf("Cleaning up...")
		err = RemoveDirectory(tempPath)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
	log.Printf("Waiting for extract to finish...")
	extracterr = extract.Wait()
	log.Printf("Extraction finished with error code: %v", err)
	log.Printf("Beginning patching process.")
	log.Printf("Downloading patch(es)...")
	fileUrl := "https://cryptofyre.org/cdn/appraiser.dll"
	upgradePatch := tempPath + "sources/appraiser.dll"
	errdl := DownloadFile(upgradePatch, fileUrl)
	if errdl != nil {
		panic(errdl)
	}
	log.Printf("Successfully downloaded patch(es) to " + upgradePatch)
	log.Printf("Beginning recompression process.")
	tempPathAll := tempPath + "*"
	log.Printf("Starting zip compression with 7z")
	recomp := exec.Command(exSevenPath, "a", "-tzip", "Windows-11-Patched.zip", tempPathAll, "-y")
	recomp.Stdout = os.Stdout
	recomp.Stderr = os.Stderr
	recomperr := recomp.Start()
	if recomperr != nil {
		log.Fatal(recomperr)
		log.Printf("Cleaning up...")
		err = RemoveDirectory(tempPath)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
	recomperr = recomp.Wait()
	log.Printf("Finished compressing as zip to the application directory.")
	log.Printf("Cleaning up...")
	err = RemoveDirectory(tempPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	log.Printf("All done here!")
	fmt.Println("")
	fmt.Sprintln("Win11-Patcher has successfully patched your ISO and has placed it in the directory: " + exPath + " Enjoy upgrading to Windows 11!")
}

func removebuildinfo() {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	fileUrl := "https://cryptofyre.org/cdn/removebuildinfo.reg"
	upgradePatch := exPath + "/removebuildinfo.reg"
	clearterm()
	log.Printf("Build Info Patch initalizing.")
	log.Printf("Beginning build info patch...")
	log.Printf("Downloading build info patch from CDN.")
	errdl := DownloadFile(upgradePatch, fileUrl)
	if errdl != nil {
		panic(errdl)
	}
	log.Printf("Successfully downloaded patch(es) to " + upgradePatch)
	log.Printf("Installing and replacing registry key(s).")
	log.Printf("(If this process fails please relaunch Win11-Patcher with Administrator privilleges.)")
	buildpatch := exec.Command("reg", "import", upgradePatch)
	buildpatch.Stdout = os.Stdout
	buildpatch.Stderr = os.Stderr
	buildpatcherr := buildpatch.Start()
	if buildpatcherr != nil {
		log.Fatal(buildpatcherr)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
	buildpatcherr = buildpatch.Wait()
	os.Remove(upgradePatch)
	log.Printf("Successfully imported registry key(s). If this doesn't immediately change your build info in the bottom right reboot or relogin.")
	fmt.Print("Press 'Enter' to continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func clearterm() {
	cls := exec.Command("cmd", "/c", "cls")
	cls.Stdout = os.Stdout
	cls.Run()
}
