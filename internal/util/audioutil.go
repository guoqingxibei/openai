package util

import (
	"github.com/tcolgate/mp3"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
)

func GetAudioDuration(inputFile string) (duration float64, err error) {
	r, err := os.Open(inputFile)
	if err != nil {
		return
	}

	d := mp3.NewDecoder(r)
	var f mp3.Frame
	skipped := 0
	for {
		if err = d.Decode(&f, &skipped); err != nil {
			if err == io.EOF {
				break
			}
			return
		}
		duration = duration + f.Duration().Seconds()
	}
	return duration, nil
}

func SplitAudioByDuration(inputFile string, segmentDuration string) (files []string, err error) {
	outDir := inputFile[:len(inputFile)-4] + "-split-files"
	err = os.Mkdir(outDir, 0755)
	if err != nil {
		return
	}

	cmd := exec.Command("ffmpeg", "-i", inputFile, "-f", "segment", "-segment_time", segmentDuration, outDir+"/%03d.mp3")
	_, err = cmd.CombinedOutput()
	if err != nil {
		return
	}

	err = filepath.Walk(outDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return
	}

	sort.Strings(files)
	return
}
