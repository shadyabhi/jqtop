package main

import (
	"os"

	"github.com/hpcloud/tail"
)

func tailThis(location string) (*tail.Tail, error) {
	tailConfig := tail.Config{
		Location: &tail.SeekInfo{
			Offset: 0, Whence: os.SEEK_END,
		},
		ReOpen: true,
		Follow: true,
	}
	t, err := tail.TailFile(location, tailConfig)
	if err != nil {
		return nil, err
	}
	return t, nil
}
