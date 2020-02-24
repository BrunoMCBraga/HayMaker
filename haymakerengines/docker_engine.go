package haymakerengines

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/jhoonb/archivex"
	"golang.org/x/net/context"
)

var dockerHost string = "unix:///var/run/docker.sock"
var dockerVersion string = "1.40"
var dockerFilePath string
var dockerTag string
var dockerImageName string
var remoteDockerImageName string

var authorizationToken string
var registryAddress string

var tarTempPath string = "/tmp"
var tarName string = "docker.tar"

func createTarArchiveForDocker(tarPath *string, tarName *string, folderToAdd *string) error {

	tar := new(archivex.TarFile)
	createError := tar.Create(fmt.Sprintf("%s/%s", *tarPath, *tarName))
	if createError != nil {
		return errors.New("|" + "HayMaker->haymakerengines->docker_engine->createTarArchiveForDocker->tar.Create:" + createError.Error() + "|")
	}

	addAllError := tar.AddAll(*folderToAdd, false)
	if addAllError != nil {
		return errors.New("|" + "HayMaker->haymakerengines->docker_engine->createTarArchiveForDocker->tar.AddAll:" + addAllError.Error() + "|")
	}
	defer tar.Close()

	return nil
}

func generateAuthString(base64Token *string, regAddress *string) (*string, error) {

	decodeStringResult, decodeStringError := base64.URLEncoding.DecodeString(*base64Token)
	if decodeStringError != nil {
		return nil, errors.New("|" + "HayMaker->haymakerengines->docker_engine->generateAuthString->base64.URLEncoding.DecodeString:" + decodeStringError.Error() + "|")
	}
	usernamAndPassword := strings.Split(string(decodeStringResult), ":")

	authConfig := types.AuthConfig{
		Username:      usernamAndPassword[0],
		Password:      usernamAndPassword[1],
		ServerAddress: fmt.Sprintf("https://%s", *regAddress),
	}
	marshallResult, marshallError := json.Marshal(authConfig)
	if marshallError != nil {
		return nil, errors.New("|" + "HayMaker->haymakerengines->docker_engine->generateAuthString->json.Marshal:" + marshallError.Error() + "|")
	}

	authStr := base64.URLEncoding.EncodeToString(marshallResult)

	return &authStr, nil
}

func writeToLog(reader io.ReadCloser) error {
	defer reader.Close()
	rd := bufio.NewReader(reader)
	for {
		n, _, err := rd.ReadLine()
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		fmt.Println(string(n))
	}
	return nil
}

func buildImageFromDockerfile(dockHost *string, dockVersion *string, dockFilePath *string, tags []string, tarPath *string) error {

	fmt.Println("Building Docker Image From Dockerfile")
	ctx := context.Background()
	newClientWithOptsResult, newClientWithOptsError := client.NewClient(*dockHost, *dockVersion, nil, nil)
	if newClientWithOptsError != nil {
		return errors.New("|" + "HayMaker->haymakerengines->docker_engine->buildImageFromDockerfile->client.NewClient:" + newClientWithOptsError.Error() + "|")

	}

	openResult, openError := os.Open(*tarPath)
	if openError != nil {
		return errors.New("|" + "HayMaker->haymakerengines->docker_engine->buildImageFromDockerfile->os.Open:" + openError.Error() + "|")

	}
	defer openResult.Close()

	imageBuildResult, imageBuildError := newClientWithOptsResult.ImageBuild(ctx, openResult, types.ImageBuildOptions{
		Tags:        tags,
		ForceRemove: true,
		Remove:      true,
	})
	if imageBuildError != nil {
		return errors.New("|" + "HayMaker->haymakerengines->docker_engine->buildImageFromDockerfile->ImageBuild:" + imageBuildError.Error() + "|")
	}

	io.Copy(ioutil.Discard, imageBuildResult.Body)
	//writeToLog(imageBuildResult.Body)
	defer imageBuildResult.Body.Close()

	localNameAndTag := fmt.Sprintf("%s:%s", dockerImageName, dockerTag)
	imageTagError := newClientWithOptsResult.ImageTag(ctx, localNameAndTag, remoteDockerImageName)
	if imageTagError != nil {
		return errors.New("|" + "HayMaker->haymakerengines->docker_engine->buildImageFromDockerfile->ImageTag:" + imageTagError.Error() + "|")
	}

	return nil
}

func pushDockerImageToRemoteRepository(dockHost *string, dockVersion *string, imagName *string, authToken *string, regAddress *string) error {

	fmt.Println("Pushing Docker Image To ECR")
	ctx := context.Background()
	newClientWithOptsResult, newClientWithOptsError := client.NewClient(*dockHost, *dockVersion, nil, nil)
	if newClientWithOptsError != nil {
		return errors.New("|" + "HayMaker->haymakerengines->docker_engine->pushDockerImageToRemoteRepository->client.NewClient:" + newClientWithOptsError.Error() + "|")
	}

	generateAuthStringResult, generateAuthStringError := generateAuthString(authToken, regAddress)

	if generateAuthStringError != nil {
		return errors.New("|" + "HayMaker->haymakerengines->docker_engine->pushDockerImageToRemoteRepository->generateAuthString:" + generateAuthStringError.Error() + "|")
	}

	imagePushResult, imagePushError := newClientWithOptsResult.ImagePush(ctx, *imagName, types.ImagePushOptions{
		All:          true,
		RegistryAuth: *generateAuthStringResult,
	})

	if imagePushError != nil {
		return errors.New("|" + "HayMaker->haymakerengines->docker_engine->pushDockerImageToRemoteRepository->ImagePush:" + imagePushError.Error() + "|")
	}

	defer imagePushResult.Close()

	io.Copy(ioutil.Discard, imagePushResult)
	/*
		buf := new(bytes.Buffer)
		buf.ReadFrom(imagePushResult)
		newStr := buf.String()

		fmt.Printf(newStr)*/
	return nil
}

func destroyLocalImagesStub(dockHost *string, dockVersion *string, imagName *string, dockerImageName *string, dockerTag *string) error {

	fmt.Println("Destroying Local Images")
	ctx := context.Background()
	newClientWithOptsResult, newClientWithOptsError := client.NewClient(*dockHost, *dockVersion, nil, nil)
	if newClientWithOptsError != nil {
		return errors.New("|" + "HayMaker->haymakerengines->docker_engine->destroyLocalImagesStub->NewClient:" + newClientWithOptsError.Error() + "|")
	}

	newClientWithOptsResult.ImageRemove(ctx, *imagName, types.ImageRemoveOptions{
		Force: true,
	})
	/*
		if imageRemoveError != nil {
			return errors.New("|" + "HayMaker->haymakerengines->docker_engine->destroyLocalImagesStub->ImageRemove:" + imageRemoveError.Error())
		}*/

	localNameAndTag := fmt.Sprintf("%s:%s", *dockerImageName, *dockerTag)
	newClientWithOptsResult.ImageRemove(ctx, localNameAndTag, types.ImageRemoveOptions{
		Force: true,
	})
	/*
		if imageRemoveError != nil {
			return errors.New("|" + "HayMaker->haymakerengines->docker_engine->destroyLocalImagesStub->ImageRemove:" + imageRemoveError.Error())
		}*/

	return nil
}

func BuildImageAndPush() error {

	createTarArchiveForDockerError := createTarArchiveForDocker(&tarTempPath, &tarName, &dockerFilePath)
	if createTarArchiveForDockerError != nil {
		return errors.New("|" + "HayMaker->haymakerengines->docker_engine->BuildImageAndPush->createTarArchiveForDocker:" + createTarArchiveForDockerError.Error() + "|")
	}

	dockerTarPath := fmt.Sprintf("%s/%s", tarTempPath, tarName)

	//localNameAndTag := fmt.Sprintf("%s:%s", dockerImageName, dockerTag)
	tempTags := []string{dockerImageName}

	buildContainerFromDockerfileError := buildImageFromDockerfile(&dockerHost, &dockerVersion, &dockerFilePath, tempTags, &dockerTarPath)
	if buildContainerFromDockerfileError != nil {
		return errors.New("|" + "HayMaker->haymakerengines->docker_engine->BuildImageAndPush->buildImageFromDockerfile:" + buildContainerFromDockerfileError.Error() + "|")
	}

	pushDockerImageToRemoteRepositoryError := pushDockerImageToRemoteRepository(&dockerHost, &dockerVersion, &remoteDockerImageName, &authorizationToken, &registryAddress)
	if pushDockerImageToRemoteRepositoryError != nil {
		return errors.New("|" + "HayMaker->haymakerengines->docker_engine->BuildImageAndPush->pushDockerImageToRemoteRepository:" + pushDockerImageToRemoteRepositoryError.Error() + "|")
	}

	removeError := os.Remove(dockerTarPath)
	if removeError != nil {
		return errors.New("|" + "HayMaker->haymakerengines->docker_engine->BuildImageAndPush->os.Remove:" + removeError.Error() + "|")
	}

	return nil
}

func DeleteLocalImages() error {

	createTarArchiveForDockerError := destroyLocalImagesStub(&dockerHost, &dockerVersion, &remoteDockerImageName, &dockerImageName, &dockerTag)
	if createTarArchiveForDockerError != nil {
		return errors.New("|" + "HayMaker->haymakerengines->docker_engine->DeleteLocalImages->deleteLocalImagesStub:" + createTarArchiveForDockerError.Error() + "|")
	}

	return nil
}

func InitDockerEngine(dockerConfig interface{}) {

	dockerTag = dockerConfig.(map[string]interface{})["tag"].(string)

	awsConfig := dockerConfig.(map[string]interface{})["aws"].(map[string]interface{})
	authorizationToken = awsConfig["authorization_token"].(string)

	registryAddress = awsConfig["ecr_repo_host"].(string)

	dockerImageName = dockerConfig.(map[string]interface{})["image_name"].(string)

	remoteDockerImageName = fmt.Sprintf("%s/%s:%s", registryAddress, dockerImageName, dockerTag)

	dockerFilePath = dockerConfig.(map[string]interface{})["dockerfile_path"].(string)

}

/*
describe image on amazon...get path
delete image from there...
*/
