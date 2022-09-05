package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/slingr-exercise/models"
	"github.com/slingr-exercise/transcripts"
	"github.com/slingr-exercise/utils"
	"github.com/slingr-exercise/wer"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

var (
	totalAccuracy    float64
	totalTimeElapsed time.Duration
	totalCPUUsage    float64
	numberOfRuns     float64
)

func main() {
	Transcribe("harbor.ops.veritone.com/challenges/deepspeech", "audio1.wav")
	Transcribe("harbor.ops.veritone.com/challenges/deepspeech", "audio2.wav")

	fmt.Printf("Average Accuracy: %%%f \n", (totalAccuracy/numberOfRuns)*float64(100))
	runs := time.Duration(numberOfRuns)
	fmt.Printf("Average Time Elapsed: %s \n", (totalTimeElapsed / runs).String())
	fmt.Printf("Average CPU Usage: %%%f", totalCPUUsage/numberOfRuns)
}

func Transcribe(image string, audioName string) {
	startTime := time.Now()

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		fmt.Println("Unable to create docker client")
		panic(err)
	}

	PullOrUpdateImage(cli, image)

	containerConfig := &container.Config{
		Image: image,
		Cmd:   []string{"--audio", audioName},
	}
	cont, err := cli.ContainerCreate(context.Background(), containerConfig, nil, nil, nil, "")
	if err != nil {
		fmt.Println("Unable to create docker container")
		panic(err)
	}

	if err := cli.ContainerStart(context.Background(), cont.ID, types.ContainerStartOptions{}); err != nil {
		fmt.Println("Unable to start docker container")
		panic(err)
	}

	cpuUsage, err := getCPUUsage(cli, cont.ID)
	if err != nil {
		fmt.Println("Unable to get docker stats")
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

	audioSize, err := getAudioSize(cli, cont.ID)
	if err != nil {
		fmt.Println("Unable to get audio size")
		panic(err)
	}

	out, err := cli.ContainerLogs(context.Background(), cont.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		panic(err)
	}

	result := utils.GetStringResult(out)

	audio1Transcript := transcripts.GetAudioTranscript(audioName)

	reference := strings.Split(strings.ToLower(audio1Transcript), " ")
	candidate := strings.Split(strings.ToLower(result), " ")

	wordErrorRate, wordAccuracy, sub, del, ins := wer.WER(reference, candidate)
	fmt.Printf("Word Error Rate: %%%f\n", wordErrorRate*float64(100))
	fmt.Printf("Word Accuracy: %%%f\n", wordAccuracy*float64(100))
	fmt.Printf("Subs:%d - Ins:%d - Del:%d \n", sub, del, ins)

	if wordAccuracy > float64(0.8) {
		fmt.Println("Transcription Passed")
	} else {
		fmt.Println("Transcription not passed, Word accuracy below 80%")
	}

	endTime := time.Now()
	totalTime := endTime.Sub(startTime)
	fmt.Println("Elapsed transition time: " + totalTime.String())
	fmt.Printf("CPU average Usage: %%%f \n", *cpuUsage)
	fmt.Printf("Input Audio File Size: %s \n", *audioSize)

	if err := stopAndRemoveContainer(cli, cont.ID); err != nil {
		panic(err)
	}

	totalAccuracy += wordAccuracy
	totalTimeElapsed += totalTime
	totalCPUUsage += *cpuUsage
	numberOfRuns += 1
}

func PullOrUpdateImage(cli *client.Client, image string) {
	reader, err := cli.ImagePull(context.Background(), image, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}

	defer reader.Close()
	io.Copy(os.Stdout, reader)
}

func getCPUUsage(cli *client.Client, containerID string) (*float64, error) {
	statsResponse, _ := cli.ContainerStats(context.Background(), containerID, false)

	var stats models.Stats
	if err := json.NewDecoder(statsResponse.Body).Decode(&stats); err != nil {
		return nil, err
	}

	cpuDelta := stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage
	systemCpuDelta := stats.CPUStats.SystemCPUUsage - stats.PreCPUStats.SystemCPUUsage
	cpuUsage := (float64(cpuDelta) / float64(systemCpuDelta)) * float64(100.0)

	return &cpuUsage, nil
}

func getAudioSize(cli *client.Client, containerID string) (*string, error) {
	out, err := cli.ContainerLogs(context.Background(), containerID, types.ContainerLogsOptions{ShowStderr: true, Tail: "2"})
	if err != nil {
		return nil, err
	}

	result := utils.GetStringResult(out)
	audioSize := strings.Split(strings.ToLower(result), " ")[4]

	return &audioSize, err
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
