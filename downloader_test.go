package httpDownloader

import (
	"bytes"
	"io"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"testing"
)

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
	log.Println("callback Update, fileSize:", fileSize, ", index:", packet.Index, ", len:", len(packet.Data), ", range:", packet.RangeStart, "-", packet.RangeEnd)
	atomic.AddInt32(countOfCompleted, 1)
	atomic.AddInt32(byteOfCompleted, int32(len(packet.Data)))

	fileMutex.Lock()
	defer fileMutex.Unlock()
	dataMap[packet.Index] = packet
}

var dataMap map[int]Packet
var wg sync.WaitGroup
var countOfCompleted *int32 //完成幾次下載
var byteOfCompleted *int32  //總共已經下載多少
var file *os.File
var fileMutex sync.Mutex
var getNumOfGoroutine = 5 //goroutine數量
var lengthOfPacket = 4096 //封包大小

const fileName = "ZlCp4Dl.jpg"                    //"ZlCp4Dl.jpg"
const fileURL = "https://i.imgur.com/" + fileName //要被下載的檔案與路徑

func Test_donwloader_txt(t *testing.T) {
	log.Println(fileURL)

	dataMap = make(map[int]Packet)

	defer func() {
		data := getData()
		log.Println("getdata:", len(data))

		file, err := os.Create(fileName)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		_, err = io.Copy(file, bytes.NewReader(data))
		if err != nil {
			log.Fatal(err)
		}
	}()

	value := int32(0)
	countOfCompleted = &value

	value2 := int32(0)
	byteOfCompleted = &value2

	wg.Add(1)

	targetBuilder, err := TargetBuilder(fileURL)
	if err != nil {
		t.Fatal("TargetBuilder, ", err)
	}
	target := targetBuilder.SetLengthOfPacket(lengthOfPacket).SetNumOfGoroutine(getNumOfGoroutine).Build()

	if target.GetNumOfGoroutine() != getNumOfGoroutine {
		t.Fatal("target.GetNumOfGoroutine!=1024, ", target.GetNumOfGoroutine())
	}

	if target.GetLengthOfPacket() != lengthOfPacket {
		t.Fatal("target.GetNumOfGoroutine!=1024, ", target.GetNumOfGoroutine())
	}

	httpProgressListener := httpProgressListener{}
	downloader := Downloader{}
	downloader.setProgressListener(&httpProgressListener)
	successed := downloader.Start(*target)
	if !successed {
		t.Fatal("downloader.Start, ")
	}

	wg.Wait()

	t.Log("countOfCompleted:", atomic.LoadInt32(countOfCompleted))
	t.Log("byteOfCompleted:", atomic.LoadInt32(byteOfCompleted))
}

func getData() []byte {
	var data []byte
	for _, value := range dataMap {
		data = bytesCombine(data, value.Data)
	}
	return data
}

func bytesCombine(pBytes ...[]byte) []byte {
	return bytes.Join(pBytes, []byte(""))
}
