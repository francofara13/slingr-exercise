package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/slingr-exercise/models"
	"github.com/slingr-exercise/transcripts"
	"github.com/slingr-exercise/utils"
	"github.com/slingr-exercise/wer"
)

var (
	totalAccuracy    float64
	totalTimeElapsed time.Duration
	totalCPUUsage    float64
	numberOfRuns     float64
)

func main() {
	f, err := os.Create("report.txt")
	if err != nil {
		panic(err)
	}

	defer f.Close()

	f.WriteString("Transcriptions Report\n")

	Transcribe("harbor.ops.veritone.com/challenges/deepspeech", "audio1.wav", f)
	Transcribe("harbor.ops.veritone.com/challenges/deepspeech", "audio2.wav", f)
	Transcribe("harbor.ops.veritone.com/challenges/deepspeech", "audio3.wav", f)
	Transcribe("harbor.ops.veritone.com/challenges/deepspeech", "audio4.wav", f)
	Transcribe("harbor.ops.veritone.com/challenges/deepspeech", "audio5.wav", f)

	f.WriteString(fmt.Sprintf("\nAverage Accuracy: %%%f \n", (totalAccuracy/numberOfRuns)*float64(100)))
	runs := time.Duration(numberOfRuns)
	f.WriteString(fmt.Sprintf("Average Time Elapsed: %s \n", (totalTimeElapsed / runs).String()))
	f.WriteString(fmt.Sprintf("Average CPU Usage: %%%f", totalCPUUsage/numberOfRuns))
}

func Transcribe(image string, audioName string, reportFile *os.File) {
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

	result := utils.GetStringResult(out, false)

	audio1Transcript := utils.Strip(transcripts.GetAudioTranscript(audioName))

	reference := strings.Split(strings.ToLower(audio1Transcript), " ")
	candidate := strings.Split(strings.ToLower(result), " ")

	wordErrorRate, wordAccuracy, sub, del, ins := wer.WER(reference, candidate)

	numberOfRuns += 1
	reportFile.WriteString(fmt.Sprintf("\nTranscription Number %d - Audio: %s \n\n", int64(numberOfRuns), audioName))
	reportFile.WriteString(fmt.Sprintf("Expected: \n%s\n\n", audio1Transcript))
	reportFile.WriteString(fmt.Sprintf("Actual: \n%s\n\n", result))

	if wordAccuracy > float64(0.8) {
		reportFile.WriteString("Transcription Passed\n")
	} else {
		reportFile.WriteString("Transcription not passed, Word accuracy below 80%\n")
	}

	reportFile.WriteString(fmt.Sprintf("Word Error Rate: %%%f\n", wordErrorRate*float64(100)))
	reportFile.WriteString(fmt.Sprintf("Word Accuracy: %%%f\n", wordAccuracy*float64(100)))
	reportFile.WriteString(fmt.Sprintf("Substitutions:%d - Insertions:%d - Deletions:%d \n", sub, del, ins))

	endTime := time.Now()
	totalTime := endTime.Sub(startTime)
	reportFile.WriteString("Elapsed transition time: " + totalTime.String() + "\n")
	reportFile.WriteString(fmt.Sprintf("CPU average Usage: %%%f \n", *cpuUsage))
	reportFile.WriteString(fmt.Sprintf("Input Audio File Size: %s \n", *audioSize))

	if err := stopAndRemoveContainer(cli, cont.ID); err != nil {
		panic(err)
	}

	totalAccuracy += wordAccuracy
	totalTimeElapsed += totalTime
	totalCPUUsage += *cpuUsage
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

	result := utils.GetStringResult(out, true)
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
