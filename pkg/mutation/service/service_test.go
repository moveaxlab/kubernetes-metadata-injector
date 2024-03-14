package service

import (
	"io/ioutil"

	"github.com/sirupsen/logrus"
)

func logger() *logrus.Entry {
	mute := logrus.StandardLogger()
	mute.Out = ioutil.Discard
	return mute.WithField("logger", "test")
}
