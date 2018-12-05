// +build !windows

// Copyright 2013, Örjan Persson. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package logging

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"time"
	"os"
	"sync"
)

type color int

const (
	ColorBlack = iota + 30
	ColorRed
	ColorGreen
	ColorYellow
	ColorBlue
	ColorMagenta
	ColorCyan
	ColorWhite
)

var (
	colors = []string{
		CRITICAL: ColorSeq(ColorMagenta),
		ERROR:    ColorSeq(ColorRed),
		WARNING:  ColorSeq(ColorYellow),
		NOTICE:   ColorSeq(ColorGreen),
		DEBUG:    ColorSeq(ColorCyan),
	}
	boldcolors = []string{
		CRITICAL: ColorSeqBold(ColorMagenta),
		ERROR:    ColorSeqBold(ColorRed),
		WARNING:  ColorSeqBold(ColorYellow),
		NOTICE:   ColorSeqBold(ColorGreen),
		DEBUG:    ColorSeqBold(ColorCyan),
	}
)

// LogBackend utilizes the standard log module.
type LogBackend struct {
	Logger      *log.Logger
	Color       bool
	ColorConfig []string
}

// NewLogBackend creates a new LogBackend.
func NewLogBackend(out io.Writer, prefix string, flag int) *LogBackend {
	return &LogBackend{Logger: log.New(out, prefix, flag)}
}

// Log implements the Backend interface.
func (b *LogBackend) Log(level Level, calldepth int, rec *Record) error {
	if b.Color {
		col := colors[level]
		if len(b.ColorConfig) > int(level) && b.ColorConfig[level] != "" {
			col = b.ColorConfig[level]
		}

		buf := &bytes.Buffer{}
		buf.Write([]byte(col))
		buf.Write([]byte(rec.Formatted(calldepth + 1)))
		buf.Write([]byte("\033[0m"))
		// For some reason, the Go logger arbitrarily decided "2" was the correct
		// call depth...
		return b.Logger.Output(calldepth+2, buf.String())
	}

	return b.Logger.Output(calldepth+2, rec.Formatted(calldepth+1))
}

// ConvertColors takes a list of ints representing colors for log levels and
// converts them into strings for ANSI color formatting
func ConvertColors(colors []int, bold bool) []string {
	converted := []string{}
	for _, i := range colors {
		if bold {
			converted = append(converted, ColorSeqBold(color(i)))
		} else {
			converted = append(converted, ColorSeq(color(i)))
		}
	}

	return converted
}

func ColorSeq(color color) string {
	return fmt.Sprintf("\033[%dm", int(color))
}

func ColorSeqBold(color color) string {
	return fmt.Sprintf("\033[%d;1m", int(color))
}

func doFmtVerbLevelColor(layout string, level Level, output io.Writer) {
	if layout == "bold" {
		output.Write([]byte(boldcolors[level]))
	} else if layout == "reset" {
		output.Write([]byte("\033[0m"))
	} else {
		output.Write([]byte(colors[level]))
	}
}

//自己搞的对象，日志文件
type WxLogBackend struct {
	Logger      *log.Logger
	Color       bool
	ColorConfig []string
	fileFd      *os.File
	fileName    string
	fileDate    string
	mu          sync.Mutex
}

func NewWxLogBackend(out *os.File, prefix string, flag int, filename string, filedate string) *WxLogBackend {
	return &WxLogBackend{
		Logger: log.New(out, prefix, flag),
		fileFd: out,
		fileName: filename,
		fileDate: filedate,
	}
}

func (b *WxLogBackend) Log(level Level, calldepth int, rec *Record) error {
	//文件日期修改，需要重新初始化日志文件
	tempFileDate := time.Now().Local().Format("2006-01-02")
	if tempFileDate != b.fileDate {
		//获取锁
		b.mu.Lock()
		defer b.mu.Unlock()
		if tempFileDate != b.fileDate {
			newFileName := b.fileName + "." + tempFileDate
			f, err := os.OpenFile(newFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
			if err == nil {
				b.Logger = log.New(f, "", 0)
				b.fileFd.Close()
				b.fileDate = tempFileDate
				b.fileFd = f
			}
		}
	}
	
	if b.Color {
		col := colors[level]
		if len(b.ColorConfig) > int(level) && b.ColorConfig[level] != "" {
			col = b.ColorConfig[level]
		}
		
		buf := &bytes.Buffer{}
		buf.Write([]byte(col))
		buf.Write([]byte(rec.Formatted(calldepth + 1)))
		buf.Write([]byte("\033[0m"))
		// For some reason, the Go logger arbitrarily decided "2" was the correct
		// call depth...
		return b.Logger.Output(calldepth+2, buf.String())
	}
	
	return b.Logger.Output(calldepth+2, rec.Formatted(calldepth+1))
}
