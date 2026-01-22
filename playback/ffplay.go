package playback

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/zi0p4tch0/radiogogo/common"
)

// FFPlayPlaybackManager represents a playback manager for FFPlay.
type FFPlayPlaybackManager struct {
	nowPlaying     *exec.Cmd
	currentStation common.Station
	nowRecording   *exec.Cmd
	recordingPath  string
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
	// Stop recording first if active
	if _, err := d.StopRecording(); err != nil {
		return err
	}

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

func (d FFPlayPlaybackManager) IsRecordingAvailable() bool {
	_, err := exec.LookPath("ffmpeg")
	return err == nil
}

func (d FFPlayPlaybackManager) RecordingNotAvailableErrorString() string {
	return `Recording requires "ffmpeg" to be installed and available in your PATH.`
}

func (d FFPlayPlaybackManager) IsRecording() bool {
	return d.nowRecording != nil
}

func (d *FFPlayPlaybackManager) StartRecording(outputPath string) error {
	if !d.IsPlaying() {
		return fmt.Errorf("cannot start recording: no station is playing")
	}

	// Stop any existing recording first
	if _, err := d.StopRecording(); err != nil {
		return err
	}

	// Start ffmpeg recording: ffmpeg -i <stream_url> -c copy output.ext
	// Use -y to overwrite existing files without prompting
	cmd := exec.Command("ffmpeg", "-y", "-i", d.currentStation.Url.URL.String(), "-c", "copy", outputPath)

	// Suppress ffmpeg's stderr output (it's verbose)
	cmd.Stderr = nil
	cmd.Stdout = nil

	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start recording: %w", err)
	}

	d.nowRecording = cmd
	d.recordingPath = outputPath
	return nil
}

func (d *FFPlayPlaybackManager) StopRecording() (string, error) {
	if d.nowRecording == nil {
		return "", nil
	}

	filePath := d.recordingPath

	if runtime.GOOS == "windows" {
		killCmd := exec.Command("taskkill", "/T", "/F", "/PID", fmt.Sprintf("%d", d.nowRecording.Process.Pid))
		if err := killCmd.Run(); err != nil {
			return "", err
		}
	} else {
		// Send SIGINT (ctrl+c) for graceful shutdown - ffmpeg will finalize the file
		if err := d.nowRecording.Process.Signal(os.Interrupt); err != nil {
			// Fallback to kill if interrupt fails
			if err := d.nowRecording.Process.Kill(); err != nil {
				return "", err
			}
		}
	}

	// Wait for process to complete
	// Ignore error as process may have already exited
	_, _ = d.nowRecording.Process.Wait()

	d.nowRecording = nil
	d.recordingPath = ""

	return filePath, nil
}

func (d FFPlayPlaybackManager) CurrentRecordingPath() string {
	return d.recordingPath
}
