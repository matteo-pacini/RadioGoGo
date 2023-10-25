package playback

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/zi0p4tch0/radiogogo/common"
)

// MPVPlaybackManager represents a playback manager for MPV.
type MPVPlaybackManager struct {
	nowPlaying *exec.Cmd
}

func NewMPVbackManager() PlaybackManagerService {
	return &MPVPlaybackManager{}
}

func (d MPVPlaybackManager) Name() string {
	return "mpv"
}

func (d MPVPlaybackManager) IsPlaying() bool {
	return d.nowPlaying != nil
}

func (d MPVPlaybackManager) IsAvailable() bool {
	_, err := exec.LookPath("mpv")
	return err == nil
}

func (d MPVPlaybackManager) NotAvailableErrorString() string {
	return `RadioGoGo requires "mpv" to be installed and available in your PATH.`
}

func (d *MPVPlaybackManager) PlayStation(station common.Station, volume int) error {
	err := d.StopStation()
	if err != nil {
		return err
	}
	cmd := exec.Command("mpv", "--no-video", fmt.Sprintf("--volume=%d", volume), station.Url.URL.String())
	err = cmd.Start()
	if err != nil {
		return err
	}
	d.nowPlaying = cmd
	return nil
}

func (d *MPVPlaybackManager) StopStation() error {
	if d.nowPlaying != nil {
		if runtime.GOOS == "windows" {
			// On Windows, use taskkill to ensure all child processes are also killed.
			killCmd := exec.Command("taskkill", "/T", "/F", "/PID", fmt.Sprintf("%d", d.nowPlaying.Process.Pid))
			if err := killCmd.Run(); err != nil {
				return err
			}
		} else {
			// On other platforms, just use the normal Kill method.
			if err := d.nowPlaying.Process.Kill(); err != nil {
				return err
			}
		}

		_, err := d.nowPlaying.Process.Wait()
		if err != nil {
			return err
		}
		d.nowPlaying = nil
	}
	return nil
}

func (d MPVPlaybackManager) VolumeMin() int {
	return 0
}

func (d MPVPlaybackManager) VolumeDefault() int {
	return 100
}

func (d MPVPlaybackManager) VolumeMax() int {
	return 200
}

func (d MPVPlaybackManager) VolumeIsPercentage() bool {
	return true
}
