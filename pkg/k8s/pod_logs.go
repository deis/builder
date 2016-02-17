package k8s

import (
	"io"
)

// streamPodLogs streams pod logs from streamer to logsCh, then closes succCh. On any error, errCh will receive once and and succCh wil not be closed. This func may block if you don't receive on all 3 channels.
//
// Example usage:
//
//  pls := StreamerFromKubeClient(myKubeClient, namespace, podName)
//	logsCh := make(chan string)
//	errCh := make(chan error)
// 	succCh := make(chan struct{})
//	go StreamPodLogs(pls, succCh, logsCh, errCh)
//	for {
//    select{
//    case b := <- logsCh:
//      fmt.Println(string(b))
//		case <-succCh:
//			break
//    case err := <-errCh:
//			// error streaming logs, we're done
//			fmt.Println(err)
//      return
//    }
//  }
func streamPodLogs(streamer PodLogStreamer, succCh chan<- struct{}, logsCh chan<- []byte, errCh chan<- error) {
	rc, err := streamer.Stream()
	if err != nil {
		errCh <- err
		return
	}
	defer rc.Close()
	w := chanWriter{ch: logsCh}
	if _, err := io.Copy(w, rc); err != nil {
		errCh <- err
		return
	}
	close(succCh)
}

// WritePodLogs writes all logs streamed from streamer to w. It returns the number of bytes written to w on success, and an appropriate error on failure.
//
// Example usage:
//
//	// get a log streamer to print all logs since the beginning of time for 'myPod' in 'myNamespace', even if the pod is not running anymore
//	s := LogStreamerFromKubeClient(myKubeClient, "myNamespace", "myPod", true, time.Time{})
//	n, err := WritePodLogs(myLogsStreamer, os.Stdout)
//	if err != nil {
//		log.Fatal(err)
//	}
//	log.Printf("Total size of logs: %d", n)
func WritePodLogs(streamer PodLogStreamer, w io.Writer) (int, error) {
	succCh := make(chan struct{})
	logsCh := make(chan []byte)
	errCh := make(chan error)
	go streamPodLogs(streamer, succCh, logsCh, errCh)
	logsLen := 0
	for {
		select {
		case <-succCh:
			return logsLen, nil
		case b := <-logsCh:
			n, err := w.Write(b)
			if err != nil {
				return 0, err
			}
			logsLen += n
		case err := <-errCh:
			return 0, err
		}
	}
}
