# httpDownloader
適合下載較大的 http資源

## Features
- 檔案分割成多個區塊進行下載
- 可設定多條 Goroutine讓下載速度加快
- 可設定暫存檔案大小

## Sample
```golang
package httpDownloader

import (
	"fmt"
	"log"
	"os"
	"sync"
	"testing"
)

const fileName = "ZlCp4Dl.jpg"                    // filename
const fileURL = "https://i.imgur.com/" + fileName // url
const getNumOfGoroutine = 5                       // number of goroutine
const lengthOfPacket = 4096                       // temp file size

var dataMap map[int]Packet
var wg sync.WaitGroup
var file *os.File
var fileMutex sync.Mutex

type httpProgressListener struct {
}

func (this *httpProgressListener) Successed() {
	log.Println("callback Successed")
	wg.Done()
}

func (this *httpProgressListener) Failed() {
	log.Println("callback Failed")
	wg.Done()
}

func (this *httpProgressListener) Update(fileSize int, packet Packet) {
	log.Println("callback Update, fileSize:", fileSize, ", index:", packet.Index, ", range:", packet.RangeStart, "-", packet.RangeEnd, ", len:", packet.LenOfPacket, ", temp filename:", packet.TmpFilename)

	fileMutex.Lock()
	defer fileMutex.Unlock()
	dataMap[packet.Index] = packet
}

func Test_donwloader(t *testing.T) {
	log.Println(fileURL)

	dataMap = make(map[int]Packet)

	wg.Add(1)

	targetBuilder, err := TargetBuilder(fileURL)
	target := targetBuilder.SetLengthOfPacket(lengthOfPacket).SetNumOfGoroutine(getNumOfGoroutine).Build()
	httpProgressListener := httpProgressListener{}
	downloader := Downloader{}
	downloader.setProgressListener(&httpProgressListener)
	successed := downloader.Start(*target)
	if !successed {
		t.Fatal("downloader.Start, ")
	}

	wg.Wait()

	file, err := os.Create(fileName)
	if err != nil {
		fmt.Println("os.Create", err)
		return
	}
	defer file.Close()

	numOfTempFile := len(dataMap)
	for index := 0; index < numOfTempFile; index++ {
		filename2 := dataMap[index].TmpFilename
		mergeFile(filename2, file)
		os.Remove(filename2)
	}
}

func mergeFile(rfilename string, wfile *os.File) error {
	rfile, err := os.OpenFile(rfilename, os.O_RDWR, 0666)
	if err != nil {
		log.Println("os.OpenFile, err:", err)
		return err
	}
	defer rfile.Close()

	stat, err := rfile.Stat()
	if err != nil {
		return err
	}

	num := stat.Size()

	buf := make([]byte, 1024*1024)
	for i := 0; int64(i) < num; {
		length, err := rfile.Read(buf)
		if err != nil {
			log.Println("rfile.Read, err:", err)
			return err
		}
		i += length

		wfile.Write(buf[:length])
	}
	return nil
}
```

## Copyright and license
- [The Unlicense](https://github.com/lya79/httpDownloader/blob/master/LICENSE)
