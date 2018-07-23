package httpDownloader

import (
	"net/url"
)

//Target not synchronized
type Target struct {
	resURL         string //檔案位置
	numOfGoroutine int    //使用幾個 Goroutine進行下載
	lengthOfPacket int    //檔案被分割後的美個封包大小
}

func (this *Target) Clone() *Target {
	clone := &Target{}
	clone.resURL = this.resURL
	clone.numOfGoroutine = this.numOfGoroutine
	clone.lengthOfPacket = this.lengthOfPacket
	return clone
}

func TargetBuilder(resURL string) (*Target, error) {
	_, err := url.ParseRequestURI(resURL)
	if err != nil {
		return nil, err
	}

	target := &Target{}
	target.resURL = resURL
	return target, nil
}

func (this *Target) GetURL() string {
	return this.resURL
}

func (this *Target) SetNumOfGoroutine(numOfGoroutine int) *Target {
	this.numOfGoroutine = numOfGoroutine
	return this
}

func (this *Target) GetNumOfGoroutine() int {
	if this.numOfGoroutine < 1 {
		this.numOfGoroutine = 1
	}
	return this.numOfGoroutine
}

func (this *Target) SetLengthOfPacket(lengthOfPacket int) *Target {
	this.lengthOfPacket = lengthOfPacket
	return this
}

func (this *Target) GetLengthOfPacket() int {
	if this.lengthOfPacket < 1 {
		this.lengthOfPacket = 1
	}
	return this.lengthOfPacket
}

func (this *Target) Build() *Target {
	return this.Clone()
}
