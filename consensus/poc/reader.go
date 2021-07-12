package poc

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/core/utils"
)

type ReadReply struct {
	Buffer     []byte
	Len        uint64
	View       uint32
	GenSig     []byte
	PlotName   string
	StartNonce uint64
	NumNonce   uint64
	Finished   bool
}

type Reader struct {
	DriveIdToPlots map[string]*PlotsDetail
	TxReadReplies  chan<- *ReadReply
	Interupts      chan struct{}
	BufferSize     uint64
	Wg             sync.WaitGroup
}

func NewReader(driveIdToPlots map[string]*PlotsDetail, txReadReplies chan<- *ReadReply, bufferSize uint64) (*Reader, error) {
	reader := &Reader{
		DriveIdToPlots: driveIdToPlots,
		TxReadReplies:  txReadReplies,
		Interupts:      make(chan struct{}, 10),
		BufferSize:     bufferSize,
	}
	return reader, nil
}

func (self *Reader) startReading(view uint32, scoop uint32, gensig []byte) {
	//terminate old task if any
	if self.Interupts != nil {
		for _, _ = range self.DriveIdToPlots {
			self.Interupts <- struct{}{}
		}
		self.Wg.Wait()
	}

	self.Interupts = make(chan struct{}, len(self.DriveIdToPlots))
	for _, detail := range self.DriveIdToPlots {
		log.Infof("startReading create reader task")
		go self.readerTask(detail, view, scoop, gensig)
	}

}

func (self *Reader) readerTask(detail *PlotsDetail, view uint32, scoop uint32, gensig []byte) {
	self.Wg.Add(1)
	defer self.Wg.Done()

	plots := detail.Plots
	for i, plot := range plots {
		var err error

		_, file := filepath.Split(plot.FilePath)
		if !plot.Removed {
			if plot.FileHandle != nil {
				plot.FileHandle.Close()
				plot.FileHandle = nil
			}
			plot.FileHandle, err = os.Open(plot.FilePath)
			if err == nil {
				_, err = plot.FileHandle.Seek(int64(scoop)*int64(plot.Nonces)*utils.SCOOP_SIZE, 0)
			}
		}

		if err != nil || plot.Removed {
			buffer := bufferPool.Get().([]byte)[:0]
			readReply := &ReadReply{
				Buffer:     buffer,
				View:       view,
				PlotName:   file,
				StartNonce: plot.StartNonce,
				NumNonce:   plot.Nonces,
				Finished:   i == (len(plots) - 1),
			}
			self.TxReadReplies <- readReply

			log.Debugf("readerTask skip plot file %s", plot.FilePath)
			continue
		}

		plot.ReadOffset = 0
		for plot.ReadOffset < utils.SCOOP_SIZE*plot.Nonces {

			select {
			case <-self.Interupts:
				log.Debugf("readerTask interupted")
				return
			default:
			}

			readOffset := plot.ReadOffset

			bufferCap := self.BufferSize
			bytesToRead := bufferCap

			log.Debugf("readerTask BufferSize %d, bufferCap %d", self.BufferSize, bufferCap)

			if plot.ReadOffset+bufferCap > utils.SCOOP_SIZE*plot.Nonces {
				bytesToRead = (utils.SCOOP_SIZE * plot.Nonces) - plot.ReadOffset
			}
			log.Debugf("readerTask readOffset %d, bytesToRead %d", readOffset, bytesToRead)
			log.Debugf("readerTask total need to read %d", utils.SCOOP_SIZE*plot.Nonces)

			buffer := bufferPool.Get().([]byte)[:bufferCap]
			buffer = buffer[:bytesToRead]

			nRead, err := plot.FileHandle.Read(buffer)
			needSkip := false
			if err != nil {
				log.Infof("readerTask read plot error %v", err)
				needSkip = true
			} else if uint64(nRead) != bytesToRead {
				log.Infof("readerTask read %d bytes, expected %d", nRead, bytesToRead)
				needSkip = true
			}
			if needSkip {
				readReply := &ReadReply{
					Buffer:     buffer[:0],
					View:       view,
					PlotName:   file,
					StartNonce: plot.StartNonce + readOffset/utils.SCOOP_SIZE,
					NumNonce:   bytesToRead / utils.SCOOP_SIZE,
					Finished:   i == (len(plots) - 1),
				}

				self.TxReadReplies <- readReply
				break
			}
			log.Debugf("readerTask read %d bytes", nRead)

			plot.ReadOffset += uint64(nRead)

			nextPlot := plot.ReadOffset >= (utils.SCOOP_SIZE * plot.Nonces)
			finished := i == (len(plots)-1) && nextPlot

			readReply := &ReadReply{
				Buffer:     buffer,
				Len:        bytesToRead,
				View:       view,
				GenSig:     gensig,
				PlotName:   file,
				StartNonce: plot.StartNonce + readOffset/utils.SCOOP_SIZE,
				NumNonce:   bytesToRead / utils.SCOOP_SIZE,
				Finished:   finished,
			}

			self.TxReadReplies <- readReply

			if !nextPlot {
				_, err := plot.FileHandle.Seek(int64(scoop)*int64(plot.Nonces)*utils.SCOOP_SIZE+int64(plot.ReadOffset), 0)
				if err != nil {
					break
				}
			}
		}
		if plot.FileHandle != nil {
			plot.FileHandle.Close()
			plot.FileHandle = nil
		}
	}
	log.Infof("reader task finish")
}
