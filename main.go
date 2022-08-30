package main

import (
	"fmt"
	"io"
	"log"
	"os"

	astideepspeech "github.com/asticode/go-astideepspeech"
	"github.com/cryptix/wav"
)

func main() {

	var model = "C:\\Users\\franc\\Documents\\deepspeech\\deepspeech-0.9.0-models.pbmm"
	var scorer = "C:\\Users\\franc\\Documents\\deepspeech\\deepspeech-0.9.0-models.scorer"
	var audio = "C:\\Users\\franc\\Documents\\deepspeech\\audio\\._2830-3980-0043.wav"

	// Initialize DeepSpeech
	m, err := astideepspeech.New(model)
	if err != nil {
		log.Fatal("Failed initializing model: ", err)
	}
	defer m.Close()

	if err := m.EnableExternalScorer(scorer); err != nil {
		log.Fatal("Failed enabling external scorer: ", err)
	}

	// Stat audio
	i, err := os.Stat(audio)
	if err != nil {
		log.Fatal(fmt.Errorf("stating %s failed: %w", audio, err))
	}

	// Open audio
	f, err := os.Open(audio)
	if err != nil {
		log.Fatal(fmt.Errorf("opening %s failed: %w", audio, err))
	}

	// Create reader
	r, err := wav.NewReader(f, i.Size())
	if err != nil {
		log.Fatal(fmt.Errorf("creating new reader failed: %w", err))
	}

	// Read
	var d []int16
	for {
		// Read sample
		s, err := r.ReadSample()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(fmt.Errorf("reading sample failed: %w", err))
		}

		// Append
		d = append(d, int16(s))
	}

	// Speech to text
	var results []string
	res, err := m.SpeechToText(d)
	if err != nil {
		log.Fatal("Failed converting speech to text: ", err)
	}
	results = []string{res}
	for _, res := range results {
		fmt.Println("Text:", res)
	}
}
