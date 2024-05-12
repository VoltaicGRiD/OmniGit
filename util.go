package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/rivo/tview"
)

// ---------------
// STRUCTS
// ---------------

type CurrentTime struct {
	time   string
	rwlock sync.RWMutex
}

// ---------------
// TIME FUNCTIONS
// ---------------

func (ct *CurrentTime) Set(value string) {
	ct.rwlock.Lock()
	defer ct.rwlock.Unlock()
	ct.time = value
}

func (ct *CurrentTime) Get() string {
	ct.rwlock.RLock()
	defer ct.rwlock.RUnlock()
	return ct.time
}

func getTime(currentTime *CurrentTime) {
	now := time.Now()
	currentTime.Set(fmt.Sprintf("%d-%d-%d %d:%d:%d\n",
		now.Year(),
		now.Month(),
		now.Day(),
		now.Hour(),
		now.Minute(),
		now.Second()))
}

// ---------------
// LOGGER FUNCTIONS 
// ---------------

// Writes a message to the logger
// This message is not formatted before being output
func WriteLog(text string) *tview.TextView {
	w := log.BatchWriter()
	defer w.Close()
	fmt.Fprintln(w, text)

	return log
}

// Writes an error to the logger
// Message is written-out with red text
func WriteErr(text string) *tview.TextView {
	err := fmt.Sprintf("[red]%v[-]", text)
	w := log.BatchWriter()
	defer w.Close()
	fmt.Fprintln(w, err)

	return log
}

// Clears the logger and returns it for writting
func ClearLog() *tview.TextView {
	log.Clear()
	return log
}
