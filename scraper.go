package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"sync"
)

type Img struct {
	URL string
}

func downloadImg(url string, wg *sync.WaitGroup) {
	res, err := http.Get(url)
	if err != nil {
		fmt.Printf("Failed to download %s: %v\n", url, err)
		wg.Done()
		return
	}

	if res.StatusCode != http.StatusOK {
		res.Body.Close()
		fmt.Printf("HTTP Error response for %s: %s\n", url, res.Status)
		wg.Done()
		return
	}

	filename := path.Base(url)
	file, err := os.Create(filename)
	if err != nil {
		res.Body.Close()
		fmt.Printf("Failed to create file %s: %v", url, err)
		wg.Done()
		return
	}

	_, err = io.Copy(file, res.Body)
	if err != nil {
		res.Body.Close()
		file.Close()
		fmt.Printf("Failed to save image %s: %v\n", url, err)
		wg.Done()
		return
	}

	/* On success, cleanup all. */
	res.Body.Close()
	file.Close()
	fmt.Printf("Download Complete: %s\n", filename)
	wg.Done()
}

func main() {
	urls := [5]string{
		"https://cdni.fancaps.net/file/fancaps-animeimages/22363046.jpg",
		"https://cdni.fancaps.net/file/fancaps-animeimages/22698291.jpg",
		"https://cdni.fancaps.net/file/fancaps-animeimages/23210821.jpg",
		"https://cdni.fancaps.net/file/fancaps-animeimages/23216561.jpg",
		"https://cdni.fancaps.net/file/fancaps-animeimages/23325557.jpg",
	}

	var wg sync.WaitGroup

	for _, url := range urls {
		wg.Add(1)
		go downloadImg(url, &wg)
	}

	wg.Wait()
	fmt.Println("All done!")
}
