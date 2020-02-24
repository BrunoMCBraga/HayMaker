package haymakerengines

import (
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ecr"
)

var ecrInstance *ecr.ECR
var repoName string
var imageTAG string

func createRepositoryStub(repositoryName *string) (*ecr.CreateRepositoryOutput, error) {

	mutability := ecr.ImageTagMutabilityImmutable

	createRepositoryInputObject := &ecr.CreateRepositoryInput{
		ImageScanningConfiguration: &ecr.ImageScanningConfiguration{
			ScanOnPush: aws.Bool(true),
		},
		ImageTagMutability: &mutability,
		RepositoryName:     repositoryName,
	}

	return ecrInstance.CreateRepository(createRepositoryInputObject)
}

func deleteRepositoryStub(repositoryName *string) (*ecr.DeleteRepositoryOutput, error) {

	deleteRepositoryInputObject := &ecr.DeleteRepositoryInput{
		Force:          aws.Bool(true),
		RepositoryName: repositoryName,
	}

	return ecrInstance.DeleteRepository(deleteRepositoryInputObject)
}

func describeRepositoriesStub(repositoryNames []*string) (*ecr.DescribeRepositoriesOutput, error) {

	describeRepositoryInputObject := &ecr.DescribeRepositoriesInput{
		RepositoryNames: repositoryNames,
	}

	return ecrInstance.DescribeRepositories(describeRepositoryInputObject)
}

func getAuthorizationTokenStub() (*ecr.GetAuthorizationTokenOutput, error) {

	describeRepositoryInputObject := &ecr.GetAuthorizationTokenInput{}

	return ecrInstance.GetAuthorizationToken(describeRepositoryInputObject)
}

func batchDeleteImageStub(repositoryName *string, imageIDs []*ecr.ImageIdentifier) (*ecr.BatchDeleteImageOutput, error) {

	batchDeleteImageInputObject := &ecr.BatchDeleteImageInput{
		RepositoryName: repositoryName,
		ImageIds:       imageIDs,
	}

	return ecrInstance.BatchDeleteImage(batchDeleteImageInputObject)
}

func SpinupContainerRepository() error {

	fmt.Println("Spining Up ECR Repository")
	createRepositoryResults, createRepositoryError := createRepositoryStub(&repoName)

	if createRepositoryError != nil {
		return errors.New("|" + "HayMaker->haymakerengines->ecr_engine->SpinupContainerRepository->createRepositoryStub:" + createRepositoryError.Error() + "|")
	}

	fmt.Println(*createRepositoryResults.Repository.RepositoryUri)
	return nil
}

func DestroyECRRepository() error {

	fmt.Println("Destroying ECR Repository")

	_, deleteRepositoryError := deleteRepositoryStub(&repoName)
	if aerr, ok := deleteRepositoryError.(awserr.Error); ok {
		switch aerr.Code() {
		case ecr.ErrCodeRepositoryNotFoundException:
			return nil
		default:
			return errors.New("|" + "HayMaker->haymakerengines->ecr_engine->DeleteContainerRepository->deleteRepositoryStub:" + deleteRepositoryError.Error() + "|")
		}
	}

	return nil
}

func DestroyDockerImageOnECR() error {

	fmt.Println("Destroying Docker Image On ECR")

	imageIDs := []*ecr.ImageIdentifier{}
	imageIDs = append(imageIDs, &ecr.ImageIdentifier{
		ImageTag: &imageTAG,
	})

	_, batchDeleteImageError := batchDeleteImageStub(&repoName, imageIDs)

	if batchDeleteImageError != nil {
		return errors.New("|" + "HayMaker->haymakerengines->ecr_engine->DeleteImageFromECR->batchDeleteImageStub:" + batchDeleteImageError.Error() + "|")
	}

	return nil
}

func DescribeRepositories() error {

	fmt.Println("Getting Information for ECR Repository")

	repoNames := []*string{&repoName}

	describeRepositoriesStubResult, describeRepositoriesStubError := describeRepositoriesStub(repoNames)

	if describeRepositoriesStubError != nil {
		return errors.New("|" + "HayMaker->haymakerengines->ecr_engine->DescribeRepositories->describeRepositoriesStub:" + describeRepositoriesStubError.Error() + "|")
	}

	fmt.Println("Repositories:")
	for _, repository := range describeRepositoriesStubResult.Repositories {
		fmt.Println(fmt.Sprintf("Name:%s URI:%s", repository.RepositoryName, repository.RepositoryUri))
	}

	return nil

}

func GetAuthorizationToken() (map[string]*string, error) {

	fmt.Println("Obtaining ECR Authorization Token")
	getAuthorizationTokenStubResult, getAuthorizationTokenStubError := getAuthorizationTokenStub()

	if getAuthorizationTokenStubError != nil {
		return nil, errors.New("|" + "HayMaker->haymakerengines->ecr_engine->GetAuthorizationToken->getAuthorizationTokenStub:" + getAuthorizationTokenStubError.Error() + "|")
	}

	authorizationTokenStruct := make(map[string]*string, 0)

	if len(getAuthorizationTokenStubResult.AuthorizationData) > 0 {
		authorizationTokenStruct["token"] = getAuthorizationTokenStubResult.AuthorizationData[0].AuthorizationToken
		protocolStrippedEndpoing := strings.TrimLeft(*getAuthorizationTokenStubResult.AuthorizationData[0].ProxyEndpoint, "https://")
		authorizationTokenStruct["endpoint"] = &protocolStrippedEndpoing
		return authorizationTokenStruct, nil
	} else {
		return nil, errors.New("|" + "HayMaker->haymakerengines->ecr_engine->GetAuthorizationToken: no repository found.")
	}

	return nil, nil

}

func GetRepositoryURI() (*string, error) {

	repoNames := []*string{&repoName}

	describeRepositoriesStubResult, describeRepositoriesStubError := describeRepositoriesStub(repoNames)

	if describeRepositoriesStubError != nil {
		return nil, errors.New("|" + "HayMaker->haymakerengines->ecr_engine->GetRepositoryURI->describeRepositoriesStub:" + describeRepositoriesStubError.Error() + "|")
	}

	if len(describeRepositoriesStubResult.Repositories) > 0 {
		return describeRepositoriesStubResult.Repositories[0].RepositoryUri, nil
	} else {
		return nil, errors.New("|" + "HayMaker->haymakerengines->ecr_engine->GetRepositoryURI: no repository found.")
	}

	return nil, nil

}

func InitECREngine(ecrInst *ecr.ECR, ecrConfig interface{}, dockerConfig interface{}) {
	ecrInstance = ecrInst

	repoName = ecrConfig.(map[string]interface{})["repo_name"].(string)
	imageTAG = dockerConfig.(map[string]interface{})["tag"].(string)

}
