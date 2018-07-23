package httpDownloader

import (
	"errors"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

type Downloader struct {
	target                 *Target
	running                bool
	runningMutex           sync.RWMutex
	progressListener       ProgressListener
	progressListenerMutext sync.Mutex
	countOfCompleted       *int32 //計數目前下載完成幾個封包
	countOfGorouint        int    //計數目前運作中的 Goroutine有幾個
	countOfGorouintMutex   sync.Mutex
	countOfGorouintCond    *sync.Cond
}

type Packet struct {
	Index       int
	RangeStart  int // 起始位置
	RangeEnd    int // 終點位置
	LenOfPacket int
	TmpFilename string
}

type ProgressListener interface {
	Successed()
	Failed()
	Update(fileSize int, packet Packet) 
}

func (this *Downloader) setProgressListener(progressListener ProgressListener) bool {
	this.runningMutex.RLock()
	defer this.runningMutex.RUnlock()
	if this.running {
		return false
	}
	if progressListener == nil {
		return false
	}
	this.progressListener = progressListener
	return true
}

func (this *Downloader) IsRunning() bool {
	this.runningMutex.RLock()
	defer this.runningMutex.RUnlock()
	return this.running
}

func (this *Downloader) Stop() {
	this.runningMutex.Lock()
	defer this.runningMutex.Unlock()
	this.running = false
}

func (this *Downloader) Start(target Target) bool {
	this.runningMutex.Lock()
	defer this.runningMutex.Unlock()
	if this.running {
		return false
	}
	if this.progressListener == nil {
		return false
	}

	value := int32(0)
	this.countOfCompleted = &value

	this.countOfGorouintMutex.Lock()
	if this.countOfGorouintCond == nil {
		this.countOfGorouintCond = sync.NewCond(&this.countOfGorouintMutex)
	}
	for {
		if this.countOfGorouint > 0 {
			this.countOfGorouintCond.Wait()
			continue
		}
		break
	}
	this.countOfGorouintCond.Broadcast()
	this.countOfGorouintMutex.Unlock()

	this.running = true
	this.target = target.Clone()
	go this.exec()
	return true
}

func (this *Downloader) getRandomString(len int) string {
	if len < 1 {
		len = 1
	}
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		bytes[i] = byte(65 + rand.Intn(25)) //A=65 and Z = 65+25
	}
	return string(bytes)

}

func (this *Downloader) exec() {
	fileSize, err := this.getFileSize()
	if err != nil {
		this.progressListener.Failed()
		return
	} else if fileSize <= 0 {
		this.progressListener.Failed()
		return
	}

	numOfPacket, lenOfLastPacket, err := this.getNumOfPacket(fileSize)
	if err != nil {
		this.progressListener.Failed()
		return
	} else if numOfPacket <= 0 {
		this.progressListener.Failed()
		return
	} else if lenOfLastPacket <= 0 {
		this.progressListener.Failed()
		return
	}

	prefix := this.getRandomString(5) 

	for index := 0; index < numOfPacket; index++ {
		var lenOfPacket int
		if index == numOfPacket-1 {
			lenOfPacket = lenOfLastPacket
		} else {
			lenOfPacket = this.target.GetLengthOfPacket()
		}
		go this.sendRequest(prefix, fileSize, numOfPacket, index, lenOfPacket)
	}
}

func (this *Downloader) getFileSize() (int, error) {
	url := this.target.GetURL()

	resp, err := http.Head(url)
	if err != nil {
		return 0, err
	}

	if resp.StatusCode != http.StatusOK {
		return 0, errors.New("resp.StatusCode:" + strconv.Itoa(resp.StatusCode))
	}

	size, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		return 0, err
	}

	return size, nil
}

func (this *Downloader) getNumOfPacket(fileSize int) (int, int, error) {
	if fileSize <= 0 {
		return 0, 0, errors.New("fileSize <= 0, " + strconv.Itoa(fileSize))
	}

	numOfPacket := 0
	lenOfLastPacket := 0
	if fileSize <= this.target.GetLengthOfPacket() {
		numOfPacket = 1
		lenOfLastPacket = this.target.GetLengthOfPacket()
	} else {
		numOfPacket = fileSize / this.target.GetLengthOfPacket()
		if (fileSize % this.target.GetLengthOfPacket()) > 0 {
			lenOfLastPacket = fileSize % this.target.GetLengthOfPacket()
			numOfPacket++
		} else {
			lenOfLastPacket = this.target.GetLengthOfPacket()
		}
	}

	return numOfPacket, lenOfLastPacket, nil
}

func (this *Downloader) sendRequest(prefix string, fileSize, numOfPacket, index, lenOfPacket int) {
	this.countOfGorouintMutex.Lock()
	for {
		if this.countOfGorouint >= this.target.GetNumOfGoroutine() {
			this.countOfGorouintCond.Wait()
			continue
		}
		break
	}
	this.countOfGorouint++
	this.countOfGorouintCond.Broadcast()
	this.countOfGorouintMutex.Unlock()

	rangeStart := this.target.GetLengthOfPacket() * index
	rangeEnd := rangeStart + lenOfPacket - 1
	rangeHeader := "bytes=" + strconv.Itoa(rangeStart) + "-" + strconv.Itoa(rangeEnd)

	req, _ := http.NewRequest("GET", this.target.GetURL(), nil)
	req.Header.Add("Range", rangeHeader)

	var resp *http.Response
	var err error
	for {
		this.runningMutex.RLock()
		if this.running == false {
			this.runningMutex.RUnlock()
			this.countOfGorouintMutex.Lock()
			this.countOfGorouint--
			this.countOfGorouintCond.Broadcast()
			this.countOfGorouintMutex.Unlock()
			return
		}
		this.runningMutex.RUnlock()
		client := &http.Client{}
		resp, err = client.Do(req)
		if err != nil {
			log.Println("fail client.Do, err:", err)
			continue
		}
		break
	}
	defer resp.Body.Close()

	url := this.target.GetURL()
	subStringsSlice := strings.Split(url, "/")
	fileName := subStringsSlice[len(subStringsSlice)-1]
	tmpFilename := fileName + "-" + prefix + "-" + strconv.Itoa(index)
	tmpFilename = strings.Replace(tmpFilename, ".", "", -1)

	out, err := os.Create(tmpFilename)
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Println("fio.Copy, err:", err)
	}

	data, _ := ioutil.ReadAll(resp.Body)

	this.countOfGorouintMutex.Lock()
	this.countOfGorouint--
	this.countOfGorouintCond.Broadcast()
	this.countOfGorouintMutex.Unlock()

	packet := Packet{}
	packet.Index = index
	packet.RangeStart = rangeStart
	packet.RangeEnd = rangeEnd
	packet.LenOfPacket = len(data)
	packet.TmpFilename = tmpFilename

	this.progressListenerMutext.Lock()
	this.progressListener.Update(fileSize, packet)
	this.progressListenerMutext.Unlock()

	atomic.AddInt32(this.countOfCompleted, 1)

	this.progressListenerMutext.Lock()
	if atomic.LoadInt32(this.countOfCompleted) == int32(numOfPacket) {
		this.progressListener.Successed()
	}
	this.progressListenerMutext.Unlock()
}
