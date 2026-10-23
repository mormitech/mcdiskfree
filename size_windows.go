package main

import "os"

func diskSize(info os.FileInfo) int64 { return info.Size() }

func deviceOf(info os.FileInfo) uint64 { return 0 }
