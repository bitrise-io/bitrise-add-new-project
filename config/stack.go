package config

// Stacks contains the current stack IDs
// known by the BANP tool. The slice should
// be updated whenever the list of supported
// stacks are modified.
func Stacks() []string {
	return []string{
		"linux-docker-android-20.04",
		"linux-docker-android-lts",
		"linux-docker-android",
		"osx-vs4mac-beta",
		"osx-vs4mac-previous-stable",
		"osx-vs4mac-stable",
		"osx-xcode-10.3.x",
		"osx-xcode-11.6.x",
		"osx-xcode-11.7.x",
		"osx-xcode-12.0.x",
		"osx-xcode-12.1.x",
		"osx-xcode-12.2.x",
		"osx-xcode-12.3.x",
		"osx-xcode-12.4.x",
		"osx-xcode-12.5.x",
		"osx-xcode-13.0.x",
		"osx-xcode-edge",
	}
}
