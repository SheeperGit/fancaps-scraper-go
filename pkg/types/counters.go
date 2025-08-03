package types

import (
	"sync"
)

var (
	imgsDownloaded uint32 // Total number of downloaded images.
	imgsSkipped    uint32 // Total number of skipped images.
	imgsTotal      uint32 // Total number of images.
)

var (
	downloadMu sync.RWMutex // Prevents bad writes to the global downloaded image counter, while allowing multiple readers.
	skipMu     sync.RWMutex // Prevents bad writes to the global skipped image counter, while allowing multiple readers.
	totalMu    sync.RWMutex // Prevents bad writes to the global total image counter, while allowing multiple readers.
)

/* Returns the total number of processed images across all titles. */
func GlobalDownloadedImages() uint32 {
	downloadMu.RLock()
	defer downloadMu.RUnlock()

	return imgsDownloaded
}

/* Returns the total number of skipped images. */
func GlobalSkippedImages() uint32 {
	skipMu.RLock()
	defer skipMu.RUnlock()

	return imgsSkipped
}

/* Returns the total number of images. */
func GlobalTotalImages() uint32 {
	totalMu.RLock()
	defer totalMu.RUnlock()

	return imgsTotal
}

/* Increments total downloaded image counter by 1. */
func IncrementGlobalDownloadedCount() {
	downloadMu.Lock()
	defer downloadMu.Unlock()

	imgsDownloaded++
}

/* Increments total skipped image counter by 1. */
func IncrementGlobalSkippedCount() {
	skipMu.Lock()
	defer skipMu.Unlock()

	imgsSkipped++
}

/* Increments total image counter by 1. */
func IncrementGlobalImageCount() {
	totalMu.Lock()
	defer totalMu.Unlock()

	imgsTotal++
}
