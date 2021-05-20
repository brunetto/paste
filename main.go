package main

import (
	"bytes"
	"flag"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/brunetto/paste"
	"github.com/pkg/errors"
)

func main() {
	// before exit succesfully print execution duration
	t0 := time.Now()
	defer func() { log.Println("done in", time.Since(t0)) }()

	// input args
	var fn string
	flag.StringVar(&fn, "in", "", "input file")
	flag.Parse()

	if fn == "" {
		log.Println("empty input file name")

		flag.PrintDefaults()
		os.Exit(1)
	}

	// allow SDK to read credentials
	err := os.Setenv("AWS_SDK_LOAD_CONFIG", "1")
	dieIf(errors.Wrap(err, "can't set AWS_SDK_LOAD_CONFIG=1 env var"))

	s, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-west-1"), MaxRetries: aws.Int(1),
		CredentialsChainVerboseErrors: aws.Bool(true),
	})
	dieIf(errors.Wrap(err, "can't init aws session"))

	// set-up get function: it could be an object as well, but for now
	// just a function is enough
	rpl := paste.NewReplacer(s)

	// open input file
	in, err := os.ReadFile(fn)
	dieIf(errors.Wrapf(err, "can't open file %v", fn))

	buf := &bytes.Buffer{}

	// replace all params placeholders found in text with values from param store
	err = paste.ReplaceAll(rpl, bytes.NewBuffer(in), buf)
	dieIf(errors.Wrap(err, "can't get and replace params"))

	// truncate file
	out, err := os.Create(fn)
	dieIf(errors.Wrapf(err, "can't recreate file '%v'", fn))

	defer out.Close()                 // close on exit
	defer func() { _ = out.Sync() }() // be super-sure everything has been written to disk

	// dump replaced data to file
	_, err = buf.WriteTo(out)
	dieIf(errors.Wrapf(err, "can't write buffer to file '%v'", fn))
}

func dieIf(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
