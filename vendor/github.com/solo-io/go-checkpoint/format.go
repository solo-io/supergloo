package checkpoint

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// HomeDir returns the current users home directory irrespecitve of the OS
func HomeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

// ConfigDir returns the config directory for solo.io
func ConfigDir() (string, error) {
	d := filepath.Join(HomeDir(), ".soloio")
	_, err := os.Stat(d)
	if err == nil {
		return d, nil
	}
	if os.IsNotExist(err) {
		err = os.Mkdir(d, 0755)
		if err != nil {
			return "", err
		}
		return d, nil
	}

	return d, err
}

func getSigfile() string {
	sigfile := filepath.Join(HomeDir(), ".soloio.sig")
	configDir, err := ConfigDir()
	if err == nil {
		sigfile = filepath.Join(configDir, "soloio.sig")
	}
	return sigfile
}

// CallReport calls a basic version check
func CallReport(product string, version string, t time.Time) {
	sigfile := getSigfile()
	ctx := context.Background()
	report := &ReportParams{
		Product:       product,
		Version:       version,
		StartTime:     t,
		EndTime:       time.Now(),
		SignatureFile: sigfile,
		Type:          "r1",
	}
	Report(ctx, report)
}

// CallCheck calls a basic version check at an interval
func CallCheck(product string, version string, t time.Time) {
	signature, err := checkSignature(getSigfile())
	if err != nil {
		signature, err = generateSignature()
		if err != nil {
			signature = "siggenerror"
		}
	}
	params := &CheckParams{
		Product:   product,
		Version:   version,
		Signature: signature,
		Type:      "c1",
	}
	cb := func(resp *CheckResponse, err error) {
		if err != nil {
			return
		}
		if resp.Outdated && resp.CurrentVersion != "" && resp.CurrentVersion != version {
			fmt.Printf("A new version of %v is available. Please visit %v.\n", product, resp.CurrentDownloadURL)
		}
		return
	}
	CheckInterval(params, VersionCheckInterval, cb)
}
