package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"node-service/rabbit"
	"node-service/utils"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
)

type DeadLetterPair struct {
	Topic   string
	Message string
}

type SendDeadLettersPayload struct {
	DeadLetters []DeadLetterPair
}

type SendMessagePayload struct {
	Message string
}

type CreateDatabasePayload struct {
	Name     string
	Type     string
	Version  string
	Password string
	User     string
	UUID     uuid.UUID
}

type CreateDatabaseResponse struct {
	Status   string
	NodeIP   string
	NodePort string
}

type StatusLogsPair struct {
	Log    string
	Status string
}

var statusLogs = []StatusLogsPair{
	{
		Log:    "CREATE DATABASE",
		Status: fmt.Sprintf("%s:%s", "Database created successfully...", "70%"),
	},
	{
		Log:    "Starting PostgreSQL",
		Status: fmt.Sprintf("%s:%s", "Deployment finished successfully...", "80%"),
	},
	{
		Log:    "database system is ready to accept connections",
		Status: fmt.Sprintf("%s:%s", "Deployment finished successfully...", "100%"),
	},
}

func (r *RPCServer) CreateDatabase(payload CreateDatabasePayload, reply *CreateDatabaseResponse) error {

	go r.trackDeploymentStatus(payload.UUID.String())

	dbPort, exporterPort, err := r.createDatabase(payload.Name, payload.Password, payload.User, payload.Type, payload.Version, payload.UUID.String())
	if err != nil {
		(*reply).Status = "ERROR"
		return nil
	}

	RabbitPayload := rabbit.Job{
		DashboardName: payload.UUID.String(),
		DbType:        payload.Type,
		NodeIP:        utils.URL.MyIP,
		NodePort:      exporterPort,
		Datname:       payload.Name,
	}
	body, _ := json.Marshal(RabbitPayload)

	err = Publisher.Push(body)
	if err != nil {
		log.Println("Error sending message to monitoring queue")
	}

	(*reply).Status = "CREATED"
	(*reply).NodeIP = app.MyIP
	(*reply).NodePort = dbPort

	r.portMtx.Lock()
	defer r.portMtx.Unlock()

	portNumber, _ := strconv.Atoi(dbPort)
	delete(r.portReserved, portNumber)

	portNumber, _ = strconv.Atoi(exporterPort)
	delete(r.portReserved, portNumber)

	return nil
}

func (r *RPCServer) trackDeploymentStatus(deploymentUUID string) {

	statusCmd := r.redisClient.Set(deploymentUUID, fmt.Sprintf("%s:%s", "Preparing terraform file for deployment...", "20%"), 0)
	if err := statusCmd.Err(); err != nil {
		log.Println(err)
	}

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Println("Can't create docker client")
		return
	}

	ctx := context.Background()

	for {

		_, err = dockerClient.ContainerInspect(ctx, deploymentUUID)
		if err != nil && client.IsErrNotFound(err) {
			r.redisClient.Set(deploymentUUID, fmt.Sprintf("%s:%s", "Preparing terraform file for deployment...", "20%"), 0)
			continue
		}

		if err != nil {
			log.Println("Error inspecting container")
			return
		}

		r.redisClient.Set(deploymentUUID, fmt.Sprintf("%s:%s", "Deployment of container started...", "40%"), 0)
		break
	}

	r.redisClient.Set(deploymentUUID, fmt.Sprintf("%s:%s", "Preparing database...", "60%"), 0)
	lastPercentage := 0

	for {

		_, err = dockerClient.ContainerInspect(ctx, deploymentUUID)
		if err != nil {
			log.Println("Error inspecting container")
			return
		}

		out, err := dockerClient.ContainerLogs(ctx, deploymentUUID, container.LogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     false,
			Tail:       "10",
		})
		if err != nil {
			log.Printf("Error getting container logs: %v", err)
		}
		defer out.Close()

		var buffer bytes.Buffer

		scanner := bufio.NewScanner(out)
		for scanner.Scan() {
			buffer.WriteString(scanner.Text())
		}

		logsDump := buffer.String()

		for _, sl := range statusLogs {

			if strings.Contains(logsDump, sl.Log) {
				currPercentageSplit := strings.Split(sl.Status, ":")[1]
				currPercentageString := strings.Split(currPercentageSplit, "%")[0]
				currPercentage, _ := strconv.Atoi(currPercentageString)

				if currPercentage <= lastPercentage {
					continue
				}
				lastPercentage = currPercentage

				r.redisClient.Set(deploymentUUID, sl.Status, 0)

				if sl.Log == statusLogs[len(statusLogs)-1].Log {
					log.Println("Deployment completed.")
					return
				}
			}
		}
	}
}

func (r *RPCServer) SendMessage(payload SendMessagePayload, reply *string) error {
	r.processMessage(payload.Message)
	*reply = "OK"
	return nil
}

func (r *RPCServer) SendDeadLetters(payload SendDeadLettersPayload, reply *string) error {
	deadLetters := payload.DeadLetters

	for _, dlPair := range deadLetters {
		r.processMessage(dlPair.Message)
	}

	(*reply) = "OK"
	return nil
}

func (r *RPCServer) processMessage(message string) {

	messageInfo := strings.Split(message, "/")

	url := utils.URL.FileServiceUrl + "/regions/" + messageInfo[0] + "/types/" + messageInfo[1] + "/versions/" + messageInfo[2]
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Println("Error getting file from file service")
		return
	}
	defer response.Body.Close()

	directoryPath := messageInfo[1] + "/" + messageInfo[2]
	err = os.MkdirAll(directoryPath, 0750)
	if err != nil {
		log.Println("Error creating directories for file")
		return
	}

	filePath := filepath.Join(directoryPath, "main.tf")

	newFile, err := os.Create(filePath)
	if err != nil {
		log.Println("Error creating the file")
	}
	defer newFile.Close()

	_, err = io.Copy(newFile, response.Body)
	if err != nil {
		log.Println("Error creating the file from response")
		return
	}
}

func (r *RPCServer) createDatabase(dbName, dbPassword, dbUser, dbType, version, directoryUUID string) (string, string, error) {

	scriptLocation := dbType + "/" + version + "/main.tf"

	err := os.MkdirAll(directoryUUID, 0750)
	if err != nil {
		log.Println("Error creating directory for user uuid")
		return "", "", err
	}

	err = copyFile(scriptLocation, directoryUUID+"/main.tf")
	if err != nil {
		log.Println("Error when copying terraform file from source to user directory")
		return "", "", err
	}

	err = r.terraformInit(directoryUUID)
	if err != nil {
		log.Println("Failed to initialize Terraform:", err)
		return "", "", err
	}

	dbPort := r.getAvailablePort()
	exporterPort := r.getAvailablePort()

	cmd := exec.Command("terraform", "apply", "-auto-approve",
		"-var", fmt.Sprintf("db_name=%s", dbName),
		"-var", fmt.Sprintf("db_password=%s", dbPassword),
		"-var", fmt.Sprintf("db_user=%s", dbUser),
		"-var", fmt.Sprintf("db_port=%v", dbPort),
		"-var", fmt.Sprintf("db_container_name=%s", directoryUUID),
		"-var", fmt.Sprintf("exporter_port=%v", exporterPort),
		"-var", fmt.Sprintf("exporter_container_name=%s", directoryUUID+"exporter"),
		"-var", fmt.Sprintf("node_ip=%s", utils.URL.MyIP))

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = directoryUUID
	return strconv.Itoa(dbPort), strconv.Itoa(exporterPort), cmd.Run()
}

func copyFile(src, dst string) error {

	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return err
	}

	err = destinationFile.Sync()
	if err != nil {
		return err
	}

	return nil
}

func (r *RPCServer) terraformInit(workingDir string) error {
	cmd := exec.Command("terraform", "init")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = workingDir
	return cmd.Run()
}

func (r *RPCServer) getAvailablePort() int {
	r.portMtx.Lock()
	defer r.portMtx.Unlock()

	fromPort := 3000
	toPort := 65535
	mp := make(map[int]string)

	cmd := exec.Command("netstat", "-ano")

	var cmdOutput bytes.Buffer
	cmd.Stdout = &cmdOutput

	if err := cmd.Run(); err != nil {
		log.Println("Error when getting available port")
		return -1
	}

	output := cmdOutput.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	lines = lines[3:]

	re := regexp.MustCompile(`\s+`)

	for _, line := range lines {

		newLine := re.ReplaceAllString(line, " ")
		localAddress := strings.Split(newLine, " ")[2]
		data := strings.Split(localAddress, ":")
		port := data[len(data)-1]
		num, _ := strconv.Atoi(port)

		mp[num] = port
	}

	for i := fromPort; i <= toPort; i++ {

		if _, exists := mp[i]; exists {
			continue
		}

		if _, exists := r.portReserved[i]; !exists {
			r.portReserved[i] = i
			return i
		}
	}

	log.Println("No ports available")
	return -1
}
