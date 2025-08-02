package types

import (
	"sync"
)

var (
	imgsProcessed uint32 // Total number of processed images.
	imgsSkipped   uint32 // Total number of skipped images.
	imgsTotal     uint32 // Total number of images.
)

var (
	processedMu sync.RWMutex // Prevents bad writes to `imgsProcessed`, while allowing multiple readers.
	skippedMu   sync.RWMutex // Prevents bad writes to `imgsSkipped`, while allowing multiple readers.
	totalMu     sync.RWMutex // Prevents bad writes to `imgsTotal`, while allowing multiple readers.
)

/* Returns the total number of processed images across all titles. */
func ProcessedTotal() uint32 {
	processedMu.RLock()
	defer processedMu.RUnlock()

	return imgsProcessed
}

/* Returns the total number of skipped images. */
func SkippedTotal() uint32 {
	skippedMu.RLock()
	defer skippedMu.RUnlock()

	return imgsSkipped
}

/* Returns the total number of images. */
func ImgTotal() uint32 {
	totalMu.RLock()
	defer totalMu.RUnlock()

	return imgsTotal
}
