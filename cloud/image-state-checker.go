package cloud

import (
	"runtime"
	"sync"
	"time"

	"github.com/jackytck/alti-cli/db"
	"github.com/jackytck/alti-cli/errors"
	"github.com/jackytck/alti-cli/gql"
)

// ImageStateChecker check the image states of all images within timeout.
type ImageStateChecker struct {
	Images  <-chan db.Image
	Done    <-chan struct{}
	Result  chan<- db.Image
	Timeout time.Duration
}

// Digest checks state of each image from Images and send back the
// result to Result until either Images or Done is closed.
func (isc *ImageStateChecker) Digest() {
	for img := range isc.Images {
		select {
		case isc.Result <- isc.checkState(img):
		case <-isc.Done:
			return
		}
	}
}

// Run starts n number of goroutines to digest each image.
// If n is not positive, it will be set to number of CPU cores.
// Return n.
func (isc *ImageStateChecker) Run(n int) int {
	if n <= 0 {
		n = runtime.NumCPU()
	}
	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func() {
			isc.Digest()
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(isc.Result)
	}()

	return n
}

// checkState checks the db image state via api, until state is changed to
// 'Ready' or 'Invalid', or timeout in this client.
func (isc *ImageStateChecker) checkState(img db.Image) db.Image {
	imgCh := make(chan db.Image)
	timer := time.NewTimer(isc.Timeout)

	go func() {
		defer close(imgCh)
		i := img
		qImg, err := gql.ProjectImage(img.PID, img.IID)
		for {
			if err != nil {
				i.Error = err.Error()
				imgCh <- i
				return
			}
			i.State = qImg.State
			if qImg.State == "Ready" || qImg.State == "Invalid" {
				imgCh <- i
				return
			}
			time.Sleep(time.Second)
		}
	}()

	ret := img
	select {
	case <-timer.C:
		ret.Error = errors.ErrClientTimeout.Error()
	case ret = <-imgCh:
	}

	return ret
}
