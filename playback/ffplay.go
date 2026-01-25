package playback

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/zi0p4tch0/radiogogo/common"
	"github.com/zi0p4tch0/radiogogo/i18n"
)

// CommandExecutor defines an interface for executing commands.
// This allows for dependency injection and easier testing.
type CommandExecutor interface {
	// Command creates a new Cmd with the given name and args.
	Command(name string, args ...string) Cmd
	// LookPath searches for an executable in PATH.
	LookPath(file string) (string, error)
}

// Cmd represents an executable command.
type Cmd interface {
	// Start starts the command but doesn't wait for it to complete.
	Start() error
	// Run runs the command and waits for it to complete.
	Run() error
	// Process returns the underlying process once started.
	Process() Process
	// SetStderr sets the stderr writer.
	SetStderr(w *os.File)
	// SetStdout sets the stdout writer.
	SetStdout(w *os.File)
}

// Process represents a running process.
type Process interface {
	// Kill causes the process to exit immediately.
	Kill() error
	// Signal sends a signal to the process.
	Signal(sig os.Signal) error
	// Wait waits for the process to exit.
	Wait() (*os.ProcessState, error)
	// Pid returns the process ID.
	Pid() int
}

// realCommandExecutor is the production implementation using os/exec.
type realCommandExecutor struct{}

func (e *realCommandExecutor) Command(name string, args ...string) Cmd {
	return &realCmd{cmd: exec.Command(name, args...)}
}

func (e *realCommandExecutor) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

// realCmd wraps exec.Cmd to implement the Cmd interface.
type realCmd struct {
	cmd *exec.Cmd
}

func (c *realCmd) Start() error        { return c.cmd.Start() }
func (c *realCmd) Run() error          { return c.cmd.Run() }
func (c *realCmd) Process() Process    { return &realProcess{proc: c.cmd.Process} }
func (c *realCmd) SetStderr(w *os.File) { c.cmd.Stderr = w }
func (c *realCmd) SetStdout(w *os.File) { c.cmd.Stdout = w }

// realProcess wraps os.Process to implement the Process interface.
type realProcess struct {
	proc *os.Process
}

func (p *realProcess) Kill() error                       { return p.proc.Kill() }
func (p *realProcess) Signal(sig os.Signal) error        { return p.proc.Signal(sig) }
func (p *realProcess) Wait() (*os.ProcessState, error)   { return p.proc.Wait() }
func (p *realProcess) Pid() int                          { return p.proc.Pid }

// FFPlayPlaybackManager represents a playback manager for FFPlay.
type FFPlayPlaybackManager struct {
	nowPlaying     Cmd
	currentStation common.Station
	nowRecording   Cmd
	recordingPath  string
	executor       CommandExecutor
	defaultVolume  int // Configured default volume (0-100)
}

// NewFFPlaybackManager creates a new FFPlayPlaybackManager with the default command executor
// and the specified default volume. The volume should be in the range 0-100.
func NewFFPlaybackManager(defaultVolume int) PlaybackManagerService {
	// Clamp volume to valid range
	if defaultVolume < 0 {
		defaultVolume = 0
	} else if defaultVolume > 100 {
		defaultVolume = 100
	}
	return &FFPlayPlaybackManager{
		executor:      &realCommandExecutor{},
		defaultVolume: defaultVolume,
	}
}

// NewFFPlaybackManagerWithExecutor creates a new FFPlayPlaybackManager with a custom command executor.
// This is primarily useful for testing. Uses default volume of 80.
func NewFFPlaybackManagerWithExecutor(executor CommandExecutor) *FFPlayPlaybackManager {
	return &FFPlayPlaybackManager{
		executor:      executor,
		defaultVolume: 80,
	}
}

func (d FFPlayPlaybackManager) Name() string {
	return "ffplay"
}

func (d FFPlayPlaybackManager) IsPlaying() bool {
	return d.nowPlaying != nil
}

func (d FFPlayPlaybackManager) IsAvailable() bool {
	_, err := d.executor.LookPath("ffplay")
	return err == nil
}

func (d FFPlayPlaybackManager) NotAvailableErrorString() string {
	return i18n.T("error_ffplay_required")
}

func (d *FFPlayPlaybackManager) PlayStation(station common.Station, volume int) error {
	err := d.StopStation()
	if err != nil {
		return err
	}
	cmd := d.executor.Command("ffplay", "-nodisp", "-volume", fmt.Sprintf("%d", volume), station.Url.URL.String())
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
			killCmd := d.executor.Command("taskkill", "/T", "/F", "/PID", fmt.Sprintf("%d", d.nowPlaying.Process().Pid()))
			if err := killCmd.Run(); err != nil {
				return err
			}
		} else {
			// Unix/macOS: SIGKILL is sufficient for single-process termination
			if err := d.nowPlaying.Process().Kill(); err != nil {
				return err
			}
		}

		// Wait for process to be reaped to avoid zombie processes
		_, err := d.nowPlaying.Process().Wait()
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
	return d.defaultVolume
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
	_, err := d.executor.LookPath("ffmpeg")
	return err == nil
}

func (d FFPlayPlaybackManager) RecordingNotAvailableErrorString() string {
	return i18n.T("error_ffmpeg_required")
}

func (d FFPlayPlaybackManager) IsRecording() bool {
	return d.nowRecording != nil
}

func (d *FFPlayPlaybackManager) StartRecording(outputPath string) error {
	if !d.IsPlaying() {
		return errors.New(i18n.T("error_no_station_playing"))
	}

	// Stop any existing recording first
	if _, err := d.StopRecording(); err != nil {
		return err
	}

	// Start ffmpeg recording: ffmpeg -i <stream_url> -c copy output.ext
	// Use -y to overwrite existing files without prompting
	cmd := d.executor.Command("ffmpeg", "-y", "-i", d.currentStation.Url.URL.String(), "-c", "copy", outputPath)

	// Suppress ffmpeg's stderr output (it's verbose)
	cmd.SetStderr(nil)
	cmd.SetStdout(nil)

	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.T("error_start_recording"), err)
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
		killCmd := d.executor.Command("taskkill", "/T", "/F", "/PID", fmt.Sprintf("%d", d.nowRecording.Process().Pid()))
		if err := killCmd.Run(); err != nil {
			return "", err
		}
	} else {
		// Unix/macOS: SIGINT allows ffmpeg to finalize the output file properly
		if err := d.nowRecording.Process().Signal(os.Interrupt); err != nil {
			// Fallback to SIGKILL if SIGINT fails (process may be unresponsive)
			if err := d.nowRecording.Process().Kill(); err != nil {
				return "", err
			}
		}
	}

	// Wait for process to be reaped (ignore errors as process may have already exited)
	_, _ = d.nowRecording.Process().Wait()

	d.nowRecording = nil
	d.recordingPath = ""

	return filePath, nil
}

func (d FFPlayPlaybackManager) CurrentRecordingPath() string {
	return d.recordingPath
}
