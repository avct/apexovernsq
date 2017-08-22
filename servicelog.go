package apexovernsq

import (
	"os"
	"path"

	"github.com/apex/log"
)

func processName() string {
	return path.Base(os.Args[0])
}

func NewServiceLogContext() *log.Entry {

	hostname, err := os.Hostname()
	if err != nil {
		log.WithError(err).Warn("Unable to get hostname for service logging")
	}

	return log.WithFields(
		log.Fields{
			"service":  processName(),
			"hostname": hostname,
			"pid":      os.Getpid(),
		})
}
