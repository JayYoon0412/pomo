package audio

import (
	"fmt"
	"sort"
	"strings"

	"github.com/JayYoon0412/pomo/assets"
)

var soundMap = map[string]string{
	"fire": "sounds/fireplace.wav",
	"rain": "sounds/lofi_rain.wav",
}

// ResolveSound validates name and returns the embedded FS path for it.
// Returns a formatted error listing all available sounds if name is unknown.
func ResolveSound(name string) (string, error) {
	if _, ok := soundMap[name]; !ok {
		names := make([]string, 0, len(soundMap))
		for k := range soundMap {
			names = append(names, k)
		}
		sort.Strings(names)
		return "", fmt.Errorf("unknown sound %q. available sounds: %s", name, strings.Join(names, ", "))
	}
	return soundMap[name], nil
}

// soundData reads and returns the raw bytes for the given embedded FS path.
func soundData(path string) ([]byte, error) {
	data, err := assets.SoundsFS.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("audio: could not read sound %q: %w", path, err)
	}
	return data, nil
}
