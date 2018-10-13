package jqtop

import (
	"io"

	"github.com/hpcloud/tail"
)

// TailFile tails the file that exists at path "location"
func TailFile(location string) (*tail.Tail, error) {
	tailConfig := tail.Config{
		Location: &tail.SeekInfo{
			Offset: 0, Whence: io.SeekEnd,
		},
		ReOpen:    true,
		Follow:    true,
		MustExist: true, //As we're reading stats, file should exist?
	}
	t, err := tail.TailFile(location, tailConfig)
	if err != nil {
		return nil, err
	}
	return t, nil
}
