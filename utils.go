package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/pkg/errors"
)

func filesExist(paths ...string) bool {
	for _, path := range paths {
		if _, err := os.Stat(path); err != nil {
			return false
		}
	}
	return true
}

func md5sumFile(filePath string) ([]byte, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, errors.Wrap(err, "fail to make md5sum")
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, errors.Wrap(err, "fail to make md5sum")
	}

	return h.Sum(nil), nil
}

func notifyErrToTelegram(err error) {
	msgFmt := "## ERR on ingress ##\n%v"

	if err2 := sendMsgToTelegram(fmt.Sprintf(msgFmt, err)); err2 != nil {
		log.Printf("ERR: %s", err2)
	}
}
