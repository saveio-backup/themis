package poc

import ()

type NonceData struct {
	View                uint32
	Deadline            uint64
	PlotName            string
	Nonce               uint64
	NumNonce            uint64
	ReaderTaskProcessed bool
}

// txEmptyBuffers should not be needed since GC
func WorkTask(rxReadReplies <-chan *ReadReply, txEmptyBuffers chan<- struct{}, txNonceData chan<- *NonceData) {
	for {
		select {
		case readReply := <-rxReadReplies:
			deadline, offset := findBestDeadline(readReply.Buffer, readReply.GenSig[:])
			txNonceData <- &NonceData{
				View:                readReply.View,
				Deadline:            deadline,
				PlotName:            readReply.PlotName,
				Nonce:               offset + readReply.StartNonce,
				NumNonce:            readReply.NumNonce,
				ReaderTaskProcessed: readReply.Finished,
			}
			bufferPool.Put(readReply.Buffer)
		}
	}
}
