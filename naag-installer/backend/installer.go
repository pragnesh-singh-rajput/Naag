package backend

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

type NaagInstaller struct{}

func NewNaagInstaller() *NaagInstaller {
	return &NaagInstaller{}
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = nil
	return cmd.Run()
}

func (n *NaagInstaller) DisableSecurity() string {
	output := ""

	// Disable Defender
	defender := [][]string{
		{"add", `HKLM\SOFTWARE\Policies\Microsoft\Windows Defender`, "/v", "DisableAntiSpyware", "/t", "REG_DWORD", "/d", "1", "/f"},
		{"add", `HKLM\SOFTWARE\Policies\Microsoft\Windows Defender\Real-Time Protection`, "/v", "DisableRealtimeMonitoring", "/t", "REG_DWORD", "/d", "1", "/f"},
	}
	for _, cmd := range defender {
		err := runCommand("reg", cmd...)
		if err != nil {
			output += "[X] Failed to disable Defender\n"
		} else {
			output += "[✓] Defender disabled\n"
		}
	}

	// Disable Firewall
	firewall := [][]string{
		{"advfirewall", "set", "allprofiles", "state", "off"},
	}
	for _, cmd := range firewall {
		err := runCommand("netsh", cmd...)
		if err != nil {
			output += "[X] Failed to disable Firewall\n"
		} else {
			output += "[✓] Firewall disabled\n"
		}
	}

	return output
}

func (n *NaagInstaller) InstallTools() string {
	tools := map[string]string{
		"PEStudio":     "https://www.winitor.com/tools/pestudio.zip",
		"Sysinternals": "https://download.sysinternals.com/files/SysinternalsSuite.zip",
		"Wireshark":    "https://1.na.dl.wireshark.org/win64/Wireshark-win64.exe",
	}

	currentUser, _ := user.Current()
	desktop := filepath.Join(currentUser.HomeDir, "Desktop")
	naagFolder := filepath.Join(desktop, "naag-tools")

	os.MkdirAll(naagFolder, os.ModePerm)

	output := ""

	for name, url := range tools {
		output += fmt.Sprintf("[*] Downloading %s...\n", name)

		filename := filepath.Join(naagFolder, name+filepath.Ext(url))
		err := downloadFile(url, filename)
		if err != nil {
			output += fmt.Sprintf("[X] Failed to download %s\n", name)
			continue
		}

		if strings.HasSuffix(filename, ".zip") {
			output += fmt.Sprintf("[~] Extracting %s...\n", name)
			err := unzip(filename, naagFolder)
			if err != nil {
				output += fmt.Sprintf("[X] Failed to extract %s\n", name)
			} else {
				output += fmt.Sprintf("[✓] Extracted %s\n", name)
			}
		} else {
			output += fmt.Sprintf("[✓] Downloaded installer for %s\n", name)
		}
	}

	return output
}

func downloadFile(url string, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		path := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, os.ModePerm)
			continue
		}

		os.MkdirAll(filepath.Dir(path), os.ModePerm)
		outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		_, err = io.Copy(outFile, rc)

		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	return nil
}
