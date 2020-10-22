package main

import "io"

// Route to device
type Route interface {
	Reader() io.Reader
	writer() io.Writer
}
