package hybridlog

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"sync"
	"testing"
	"time"
)

func TestHybridLog(t *testing.T) {
	l := mustOpen()
	nData := 1024 // 1kb
	data := make([]byte, nData)
	for i := 0; i < nData; i++ {
		data[i] = byte(i % 256)
	}
	var wg sync.WaitGroup
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()
			if _, err := l.Write(data); err != nil {
				t.Fatal(err)
			}
		}()
	}
	wg.Wait()
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()
			readB := make([]byte, 128)
			if n, err := l.ReadAt(readB, 128); err != nil {
				t.Fatal(err)
			} else {
				assert.Equal(t, n, len(readB))
				assert.Equal(t, 128, int(readB[0]))
			}
		}()
	}
	wg.Wait()
	all := make([]byte, nData*100)
	n, err := l.ReadAt(all, 100)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, nData*100-100, n)
	cleanup()
}

func TestCompactHybridLog(t *testing.T) {
	clog := mustOpen()
	nData := 1024 // 1kb
	data := make([]byte, nData)
	for i := 0; i < nData; i++ {
		data[i] = byte(i % 256)
	}
	var wg sync.WaitGroup
	wg.Add(10000)
	for i := 0; i < 10000; i++ {
		go func() {
			defer wg.Done()
			if _, err := clog.Write(data); err != nil {
				t.Fatal(err)
			}
		}()
	}
	wg.Wait()
	time.Sleep((fragmentationCheckInterval + 5) * time.Second)
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()
			readB := make([]byte, 128)
			if n, err := clog.ReadAt(readB, 128); err != nil {
				t.Fatal(err)
			} else {
				assert.Equal(t, n, len(readB))
				assert.Equal(t, 128, int(readB[0]))
			}
		}()
	}
	wg.Wait()
	clog.Close()
	cleanup()
}

func BenchmarkHybridLog_Write_512b(b *testing.B) {
	l := mustOpen()
	defer l.Close()
	defer cleanup()
	benchWrite(b, 512, l)
}

func BenchmarkHybridLog_Write_1KB(b *testing.B) {
	l := mustOpen()
	defer l.Close()
	defer cleanup()
	benchWrite(b, 1024, l)
}

func BenchmarkHybridLog_Write_4KB(b *testing.B) {
	l := mustOpen()
	defer l.Close()
	defer cleanup()
	benchWrite(b, 1024*4, l)
}

func BenchmarkHybridLog_Write_128KB(b *testing.B) {
	l := mustOpen()
	defer l.Close()
	defer cleanup()
	benchWrite(b, 1024*128, l)
}

func BenchmarkHybridLog_Write_1MB(b *testing.B) {
	l := mustOpen()
	defer l.Close()
	defer cleanup()
	benchWrite(b, 1024*1024, l)
}

func BenchmarkHybridLog_Read_512b(b *testing.B) {
	l := mustOpen()
	defer l.Close()
	defer cleanup()
	benchRead(b, 512, l, l)
}

func BenchmarkHybridLog_Read_1KB(b *testing.B) {
	l := mustOpen()
	defer l.Close()
	defer cleanup()
	benchRead(b, 1024, l, l)
}

func BenchmarkHybridLog_Read_4KB(b *testing.B) {
	l := mustOpen()
	defer l.Close()
	defer cleanup()
	benchRead(b, 1024*4, l, l)
}

func BenchmarkHybridLog_Read_128KB(b *testing.B) {
	l := mustOpen()
	defer l.Close()
	defer cleanup()
	benchRead(b, 1024*128, l, l)
}

func BenchmarkHybridLog_Read_1MB(b *testing.B) {
	l := mustOpen()
	defer l.Close()
	defer cleanup()
	benchRead(b, 1024*1024, l, l)
}

func BenchmarkFile_Write_512b(b *testing.B) {
	l := mustOpen()
	defer l.Close()
	defer cleanup()
	benchWrite(b, 512, l)
}

func BenchmarkFile_Write_1KB(b *testing.B) {
	l := mustOpen()
	defer l.Close()
	defer cleanup()
	benchWrite(b, 1024, l)
}

func BenchmarkFile_Write_4KB(b *testing.B) {
	l := mustOpenFile()
	defer l.Close()
	defer cleanup()
	benchWrite(b, 1024*4, l)
}

func BenchmarkFile_Write_128KB(b *testing.B) {
	l := mustOpenFile()
	defer l.Close()
	defer cleanup()
	benchWrite(b, 1024*128, l)
}

func BenchmarkFile_Write_1MB(b *testing.B) {
	l := mustOpenFile()
	defer l.Close()
	defer cleanup()
	benchWrite(b, 1024*1024, l)
}

func BenchmarkFile_Read_512b(b *testing.B) {
	l := mustOpenFile()
	defer l.Close()
	defer cleanup()
	benchRead(b, 512, l, l)
}

func BenchmarkFile_Read_1KB(b *testing.B) {
	l := mustOpenFile()
	defer l.Close()
	defer cleanup()
	benchRead(b, 1024, l, l)
}

func BenchmarkFile_Read_4KB(b *testing.B) {
	l := mustOpenFile()
	defer l.Close()
	defer cleanup()
	benchRead(b, 1024*4, l, l)
}

func BenchmarkFile_Read_128KB(b *testing.B) {
	l := mustOpenFile()
	defer l.Close()
	defer cleanup()
	benchRead(b, 1024*128, l, l)
}

func BenchmarkFile_Read_1MB(b *testing.B) {
	l := mustOpenFile()
	defer l.Close()
	defer cleanup()
	benchRead(b, 1024*1024, l, l)
}

func benchWrite(b *testing.B, dataSize int, writer io.Writer) {
	data := make([]byte, dataSize)
	for i := 0; i < dataSize; i++ {
		data[i] = byte(i % 256)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		writer.Write(data)
	}
	b.StopTimer()
}

func benchRead(b *testing.B, dataSize int, writer io.Writer, reader io.ReaderAt) {
	data := make([]byte, dataSize)
	for i := 0; i < dataSize; i++ {
		data[i] = byte(i % 256)
	}
	for i := 0; i < 1000; i++ {
		writer.Write(data)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader.ReadAt(data, 0)
	}
	b.StopTimer()
}

func mustOpen() HybridLog {
	l, err := open(Config{
		Path:          "./test.log",
		HighWaterMark: 30,
	})
	if err != nil {
		fmt.Printf("%+v", err)
		panic(err)
	}
	return l
}

func mustOpenFile() *os.File {
	l, err := os.OpenFile("./test.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Printf("%+v", err)
		panic(err)
	}
	return l
}

func cleanup() {
	os.Remove("./test.log")
}
