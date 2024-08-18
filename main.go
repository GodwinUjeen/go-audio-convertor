package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/hajimehoshi/go-mp3"
)

func main() {
	// Define the input and output directories
	inputDir := "./data"
	outputDir := "./converted"

	// Create the output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		fmt.Println("Error creating output directory:", err)
		return
	}

	var wg sync.WaitGroup

	// Loop through all files in the input directory
	err := filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Process only .mp3 files
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".mp3") {
			fmt.Printf("Processing file: %s\n", path)

			// Increment the WaitGroup counter
			wg.Add(1)

			// Run the conversion in a goroutine
			go func(mp3Path string) {
				defer wg.Done() // Decrement the counter when the goroutine completes
				err := convertMP3ToWAV(mp3Path, filepath.Join(outputDir, strings.TrimSuffix(info.Name(), ".mp3")+".wav"))
				if err != nil {
					fmt.Printf("Error converting file %s: %v\n", mp3Path, err)
				}
			}(path)
		}
		return nil
	})

	if err != nil {
		fmt.Println("Error walking through the directory:", err)
	}

	fmt.Println()
	// Wait for all goroutines to finish
	wg.Wait()

	fmt.Println("\n**** All files have been processed. ****\n")
}

// Function to convert an MP3 file to WAV
func convertMP3ToWAV(mp3Path, wavPath string) error {
	// Open the MP3 file
	mp3File, err := os.Open(mp3Path)
	if err != nil {
		return fmt.Errorf("error opening MP3 file: %w", err)
	}
	defer mp3File.Close()

	// Decode the MP3 file
	decoder, err := mp3.NewDecoder(mp3File)
	if err != nil {
		return fmt.Errorf("error decoding MP3 file: %w", err)
	}

	// Create the WAV file
	wavFile, err := os.Create(wavPath)
	if err != nil {
		return fmt.Errorf("error creating WAV file: %w", err)
	}
	defer wavFile.Close()

	// Create WAV encoder
	enc := wav.NewEncoder(wavFile, decoder.SampleRate(), 16, 2, 1)

	// Create a buffer to hold the audio data
	buf := make([]byte, 1024)

	// Read and write data
	for {
		n, err := decoder.Read(buf)
		if err != nil && err != io.EOF {
			return fmt.Errorf("error reading MP3 file: %w", err)
		}
		if n == 0 {
			break
		}

		// Convert the []byte buffer to []int
		audioBuf := &audio.IntBuffer{
			Data:           make([]int, n/2),
			Format:         &audio.Format{SampleRate: decoder.SampleRate(), NumChannels: 2},
			SourceBitDepth: 16,
		}
		for i := 0; i < n/2; i++ {
			audioBuf.Data[i] = int(int16(buf[i*2]) | int16(buf[i*2+1])<<8)
		}

		if err := enc.Write(audioBuf); err != nil {
			return fmt.Errorf("error writing WAV file: %w", err)
		}
	}

	// Close the encoder
	if err := enc.Close(); err != nil {
		return fmt.Errorf("error closing WAV file: %w", err)
	}

	fmt.Printf("Successfully converted %s to WAV\n", mp3Path)
	return nil
}
