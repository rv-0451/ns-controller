package utils

import (
	"flag"
	"net/http"
	"time"
)

func NewHttpClient() *http.Client {
	return &http.Client{
		Timeout: 1 * time.Second,
	}
}

func Overprovisioning() float64 {
	return flag.Lookup("overprovisioning").Value.(flag.Getter).Get().(float64)
}
