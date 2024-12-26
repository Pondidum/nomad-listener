package main

var GitCommit string

func version() string {
	if GitCommit == "" {
		return "local"
	}

	return GitCommit[:7]
}
