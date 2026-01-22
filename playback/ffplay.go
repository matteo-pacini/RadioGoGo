package playback

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/zi0p4tch0/radiogogo/common"
)

// FFPlayPlaybackManager represents a playback manager for FFPlay.
type FFPlayPlaybackManager struct {
	nowPlaying     *exec.Cmd
	currentStation common.Station
}

func NewFFPlaybackManager() PlaybackManagerService {
	return &FFPlayPlaybackManager{}
}

func (d FFPlayPlaybackManager) Name() string {
	return "ffplay"
}

func (d FFPlayPlaybackManager) IsPlaying() bool {
	return d.nowPlaying != nil
}

func (d FFPlayPlaybackManager) IsAvailable() bool {
	_, err := exec.LookPath("ffplay")
	return err == nil
}

func (d FFPlayPlaybackManager) NotAvailableErrorString() string {
	return `RadioGoGo requires "ffplay" (part of "ffmpeg") to be installed and available in your PATH.`
}

func (d *FFPlayPlaybackManager) PlayStation(station common.Station, volume int) error {
	err := d.StopStation()
	if err != nil {
		return err
	}
	cmd := exec.Command("ffplay", "-nodisp", "-volume", fmt.Sprintf("%d", volume), station.Url.URL.String())
	err = cmd.Start()
	if err != nil {
		return err
	}
	d.nowPlaying = cmd
	d.currentStation = station
	return nil
}

func (d *FFPlayPlaybackManager) StopStation() error {
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
		d.currentStation = common.Station{}
	}
	return nil
}

func (d FFPlayPlaybackManager) VolumeMin() int {
	return 0
}

func (d FFPlayPlaybackManager) VolumeDefault() int {
	return 80
}

func (d FFPlayPlaybackManager) VolumeMax() int {
	return 100
}

func (d FFPlayPlaybackManager) VolumeIsPercentage() bool {
	return false
}

func (d FFPlayPlaybackManager) CurrentStation() common.Station {
	return d.currentStation
}
