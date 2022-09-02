package main

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/slingr-exercise/transcripts"
	"github.com/slingr-exercise/wer"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

func main() {
	_, _ = Transcribe("harbor.ops.veritone.com/challenges/deepspeech", "audio1.wav")
}

func Transcribe(image string, audioName string) (string, error) {
	startTime := time.Now()

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		fmt.Println("Unable to create docker client")
		panic(err)
	}

	PullOrUpdateImage(cli, image)

	cont, err := cli.ContainerCreate(
		context.Background(),
		&container.Config{
			Image: image,
			Cmd:   []string{"--audio", audioName},
		},
		nil, nil, nil, "")
	if err != nil {
		fmt.Println("Unable to create docker container")
		panic(err)
	}

	if err := cli.ContainerStart(context.Background(), cont.ID, types.ContainerStartOptions{}); err != nil {
		fmt.Println("Unable to start docker container")
		panic(err)
	}

	statusCh, errCh := cli.ContainerWait(context.Background(), cont.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case <-statusCh:
	}

	out, err := cli.ContainerLogs(context.Background(), cont.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		panic(err)
	}

	result := getStringResult(out)

	audio1Transcript := transcripts.GetAudio1Transcript()

	reference := strings.Split(strings.ToLower(audio1Transcript), " ")
	candidate := strings.Split(strings.ToLower(result), " ")

	wordErrorRate, wordAccuracy, sub, del, ins := wer.WER(reference, candidate)
	fmt.Printf("Word Error Rate: %f\n", wordErrorRate*float64(100))
	fmt.Printf("Word Accuracy: %f\n", wordAccuracy*float64(100))
	fmt.Printf("Subs:%d  - Ins:%d - Del:%d \n", sub, del, ins)

	if wordAccuracy > float64(0.8) {
		fmt.Println("Transcription Passed")
	} else {
		fmt.Println("Transcription not passed, Word accuracy below 80%")
	}

	endTime := time.Now()
	totalTime := endTime.Sub(startTime)
	fmt.Println("Elapsed transition time: " + totalTime.String())

	if err := stopAndRemoveContainer(cli, cont.ID); err != nil {
		panic(err)
	}

	return cont.ID, nil
}

func strip(s string) string {
	var result strings.Builder
	for i := 0; i < len(s); i++ {
		b := s[i]
		if ('a' <= b && b <= 'z') ||
			('A' <= b && b <= 'Z') ||
			('0' <= b && b <= '9') ||
			b == ' ' {
			result.WriteByte(b)
		}
	}
	return result.String()
}

func PullOrUpdateImage(cli *client.Client, image string) {
	reader, err := cli.ImagePull(context.Background(), image, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}

	defer reader.Close()
	io.Copy(os.Stdout, reader)
}

func stopAndRemoveContainer(client *client.Client, id string) error {
	ctx := context.Background()

	if err := client.ContainerStop(ctx, id, nil); err != nil {
		log.Printf("Unable to stop container %s: %s", id, err)
	}

	removeOptions := types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}

	if err := client.ContainerRemove(ctx, id, removeOptions); err != nil {
		log.Printf("Unable to remove container: %s", err)
		return err
	}

	return nil
}

func getStringResult(reader io.Reader) string {
	buf := new(strings.Builder)
	_, _ = io.Copy(buf, reader)
	result := strip(buf.String())
	fmt.Print(result + "\n")

	return result
}
