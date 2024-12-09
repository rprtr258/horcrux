package main

import (
	"fmt"

	"github.com/pkg/errors"
)

func enewf(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}

func enew(format string) error {
	return enewf("%s", format)
}

func ewrapf(err error, format string, args ...interface{}) error {
	return errors.Wrapf(err, format, args...)
}

func ewrap(err error, format string) error {
	return ewrapf(err, "%s", format)
}
