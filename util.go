package main

import "os"

func isExecutable(mode os.FileMode) bool {
	return mode&0111 != 0
}
