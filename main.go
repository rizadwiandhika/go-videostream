package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"regexp"
	"strconv"
)

var totalRead int

func homePage(w http.ResponseWriter, r *http.Request)  {
	fmt.Println("Endpoint Hit: homePage")
	t, err := template.ParseFiles("index.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err)
		return
	}

	t.Execute(w, nil)
}

func videoStream(w http.ResponseWriter, r *http.Request)  {
	if _, ok := r.Header["Range"]; !ok {
		fmt.Println("Range header is not set!")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Header range is not set!")
		return
	}

	re := regexp.MustCompile(`\d+`)
	start, err := strconv.Atoi(re.FindString(r.Header.Get("range")))
	if err != nil {
		fmt.Println("error in Atoi", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err)
		return
	}
	
	videoPath := "./roxanne.webm"
	video, err := os.Open(videoPath)
	if err != nil {
		fmt.Println("Error opening video file")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err)
		return
	}
	defer video.Close()

	const CHUNK = 1000000 // 1 MB
	videoStat, _ := video.Stat()
	videoSize := videoStat.Size()
	end := int(math.Min(float64(start + CHUNK), float64(videoSize - 1)))
	length := end - start + 1
	if totalRead >= int(videoSize) {
		totalRead = 0
	}
	
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, videoSize))
	w.Header().Set("Content-Lenght", strconv.Itoa(length))
	w.Header().Set("Content-Type", "video/webm")

	// tell <video> we as chunk so that it will request more data untill complete
	w.WriteHeader(206) 
	
	data := make([]byte, CHUNK)
	n, err := video.ReadAt(data, int64(start))
	totalRead += n

	fmt.Printf("%dKB read of %dKB\n", totalRead / 1000, videoSize / 1000)

	switch err {
	case nil:
	case io.EOF: fmt.Println("Finished reading!")
	default: 
		fmt.Println("Error", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err)
		return
	}

	// send to <video>
	w.Write(data)
}

func handleRequest()  {
	http.HandleFunc("/", homePage)
	http.HandleFunc("/video", videoStream)

	log.Fatalln(http.ListenAndServe(":8080", nil))
}


func main() {	
	handleRequest()	
}