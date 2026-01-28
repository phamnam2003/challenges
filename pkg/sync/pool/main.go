package main

import (
	"bytes"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

var bufPool = sync.Pool{
	New: func() any {
		log.Println("create new buffer")
		// The Pool's New function should generally only return pointer
		// types, since a pointer can be put into the return interface
		// value without an allocation:
		return new(bytes.Buffer)
	},
}

// timeNow is a fake version of time.Now for tests.
func timeNow() time.Time {
	return time.Unix(1136214245, 0)
}

func Log(w io.Writer, key, val string) {
	b := bufPool.Get().(*bytes.Buffer)
	b.Reset()
	// Replace this with time.Now() in a real logger.
	b.WriteString(timeNow().UTC().Format(time.RFC3339))
	b.WriteByte(' ')
	b.WriteString(key)
	b.WriteByte('=')
	b.WriteString(val)
	b.WriteString("\n")
	w.Write(b.Bytes())
	bufPool.Put(b)
}

func main() {
	for range 10 {
		Log(os.Stdout, "path", "/search?q=flowers")
	}
}
