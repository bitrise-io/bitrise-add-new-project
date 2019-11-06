package config

// Stacks contains the current stack IDs
// known by the BANP tool. The slice should
// be updated whenever the list of supported
// stacks are modified.
func Stacks() []string {
	return []string{
		"linux-docker-android-lts",
		"linux-docker-android",
		"osx-vs4mac-beta",
		"osx-vs4mac-previous-stable",
		"osx-vs4mac-stable",
		"osx-xcode-10.0.x",
		"osx-xcode-10.1.x",
		"osx-xcode-10.2.x",
		"osx-xcode-10.3.x",
		"osx-xcode-11.0.x",
		"osx-xcode-11.1.x",
		"osx-xcode-11.2.x",
		"osx-xcode-8.3.x",
		"osx-xcode-9.4.x",
		"osx-xcode-edge",
	}
}
