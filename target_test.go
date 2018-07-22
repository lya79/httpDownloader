package httpDownloader

import (
	"testing"
)

func Test_TargetBuilder(t *testing.T) {
	getNumOfGoroutine := 5
	lengthOfPacket := 1024

	targetBuilder, err := TargetBuilder("https://www.google.com.tw")
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
}

func Test_TargetBuilder2(t *testing.T) {
	getNumOfGoroutine := 0
	lengthOfPacket := 0

	targetBuilder, err := TargetBuilder("https://www.google.com.tw")
	if err != nil {
		t.Fatal("TargetBuilder, ", err)
	}
	target := targetBuilder.SetLengthOfPacket(lengthOfPacket).SetNumOfGoroutine(getNumOfGoroutine).Build()

	if target.GetNumOfGoroutine() != 1 {
		t.Fatal("target.GetNumOfGoroutine!=1024, ", target.GetNumOfGoroutine())
	}

	if target.GetLengthOfPacket() != 1 {
		t.Fatal("target.GetNumOfGoroutine!=1024, ", target.GetNumOfGoroutine())
	}
}

func Test_TargetBuilder3(t *testing.T) {
	getNumOfGoroutine := -1
	lengthOfPacket := -1

	targetBuilder, err := TargetBuilder("https://www.google.com.tw")
	if err != nil {
		t.Fatal("TargetBuilder, ", err)
	}
	target := targetBuilder.SetLengthOfPacket(lengthOfPacket).SetNumOfGoroutine(getNumOfGoroutine).Build()

	if target.GetNumOfGoroutine() != 1 {
		t.Fatal("target.GetNumOfGoroutine!=1024, ", target.GetNumOfGoroutine())
	}

	if target.GetLengthOfPacket() != 1 {
		t.Fatal("target.GetNumOfGoroutine!=1024, ", target.GetNumOfGoroutine())
	}
}
