package k8s

import (
	"io"
)

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
func WritePodLogs(streamer PodLogStreamer, w io.Writer) (int64, error) {
	rc, err := streamer.Stream()
	if err != nil {
		return 0, err
	}
	defer rc.Close()
	return io.Copy(w, rc)
}
