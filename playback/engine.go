package playback

import (
	"errors"

	"gopkg.in/yaml.v3"
)

// PlaybackEngineType represents the type of playback engine used by the application.
type PlaybackEngineType string

// Available playback engines.
const (
	FFPlay PlaybackEngineType = "ffplay"
	MPV    PlaybackEngineType = "mpv"
)

func (p *PlaybackEngineType) UnmarshalYAML(value *yaml.Node) error {
	var val string
	if err := value.Decode(&val); err != nil {
		return err
	}

	switch PlaybackEngineType(val) {
	case FFPlay, MPV:
		*p = PlaybackEngineType(val)
		return nil
	default:
		return errors.New("invalid playbackEngine value: " + val)
	}
}
