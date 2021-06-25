package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	isoPathArray := os.Args[1:]
	isoPath := strings.Join(isoPathArray, " ")

	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)

	exSevenPath := exPath + "/7z.exe"
	fmt.Println(exSevenPath)

	emptystring := ""

	if isoPath == emptystring {
		fmt.Println("Win11-Patcher by @cryptofyre")
		fmt.Println("Usage: win11-patcher.exe C:/Users/nice/windows11.iso")
		fmt.Println("Please specify a Windows 11 .iso in the launch arguments or drag and drop a .iso into the application.")
		fmt.Println("")
		fmt.Println("DISCLAIMER: This application only allows an upgrade from Windows 10 to 11 using the Setup.exe provided in the zip after the patching process has completed. This application does not yet allow the user to install from scratch using the bootable environment.")
	} else {
		log.Printf("Win11-Patcher by @cryptofyre")
		log.Printf("DISCLAIMER: This application only allows an upgrade from Windows 10 to 11 using the Setup.exe provided in the zip after the patching process has completed. This application does not yet allow the user to install from scratch using the bootable environment.")
		log.Printf("Beginning inital check(s)")
		sevenzip, err := os.Open("7z.exe")
		if sevenzip != nil {
			log.Fatalf("Failed to find 7z.exe: %s", err)
			log.Fatalf("Please put 7z.exe next to me!")
		}
		defer sevenzip.Close()
		log.Printf("7z.exe located!")
		log.Printf("ISO: " + isoPath)

		log.Printf("Checking iso...")
		iso, err := os.Open(isoPath)
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

		extract := exec.Command(exSevenPath, "x", isoPath, tempPath7z, "-y")
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