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

// StopStation stops the currently playing station and any active recording.
// Platform-specific behavior:
//   - Windows: Uses taskkill with /T (tree kill) and /F (force) flags to kill
//     the ffplay process and all its child processes. This is necessary because
//     Windows doesn't propagate signals to child processes like Unix does.
//   - Unix/macOS: Uses SIGKILL via Process.Kill() which immediately terminates
//     the process. This is sufficient as ffplay doesn't spawn child processes
//     on these platforms.
func (d *FFPlayPlaybackManager) StopStation() error {
	// Stop recording first if active
	if _, err := d.StopRecording(); err != nil {
		return err
	}

	if d.nowPlaying != nil {
		if runtime.GOOS == "windows" {
			// Windows: taskkill /T kills entire process tree, /F forces termination
			killCmd := exec.Command("taskkill", "/T", "/F", "/PID", fmt.Sprintf("%d", d.nowPlaying.Process.Pid))
			if err := killCmd.Run(); err != nil {
				return err
			}
		} else {
			// Unix/macOS: SIGKILL is sufficient for single-process termination
			if err := d.nowPlaying.Process.Kill(); err != nil {
				return err
			}
		}

		// Wait for process to be reaped to avoid zombie processes
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

// StopRecording stops the current recording and returns the output file path.
// Platform-specific behavior:
//   - Windows: Uses taskkill to force-terminate ffmpeg. This may result in
//     slightly corrupted file endings, but is the most reliable cross-platform
//     approach on Windows.
//   - Unix/macOS: Sends SIGINT (Ctrl+C) to ffmpeg, allowing it to gracefully
//     finalize the output file (write proper headers/trailers). Falls back to
//     SIGKILL if SIGINT fails.
func (d *FFPlayPlaybackManager) StopRecording() (string, error) {
	if d.nowRecording == nil {
		return "", nil
	}

	filePath := d.recordingPath

	if runtime.GOOS == "windows" {
		// Windows: Force kill - ffmpeg doesn't handle signals well on Windows
		killCmd := exec.Command("taskkill", "/T", "/F", "/PID", fmt.Sprintf("%d", d.nowRecording.Process.Pid))
		if err := killCmd.Run(); err != nil {
			return "", err
		}
	} else {
		// Unix/macOS: SIGINT allows ffmpeg to finalize the output file properly
		if err := d.nowRecording.Process.Signal(os.Interrupt); err != nil {
			// Fallback to SIGKILL if SIGINT fails (process may be unresponsive)
			if err := d.nowRecording.Process.Kill(); err != nil {
				return "", err
			}
		}
	}

	// Wait for process to be reaped (ignore errors as process may have already exited)
	_, _ = d.nowRecording.Process.Wait()

	d.nowRecording = nil
	d.recordingPath = ""

	return filePath, nil
}

func (d FFPlayPlaybackManager) CurrentRecordingPath() string {
	return d.recordingPath
}
