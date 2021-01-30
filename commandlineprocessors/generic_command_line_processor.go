package commandlineprocessors

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/BrunoMCBraga/HayMaker/globalstringsproviders"
	"github.com/BrunoMCBraga/HayMaker/haymakerengines"
	"github.com/BrunoMCBraga/HayMaker/haymakerutil"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/elb"
)

const defaultKubeconfigPathWithinHome string = ".kube/config"

var defaultKubeconfigFile string

func readConfigFile(configFilePath *string) (interface{}, error) {

	configBytes, readfileError := ioutil.ReadFile(*configFilePath)
	if readfileError != nil {
		return nil, errors.New("|" + "|HayMaker->commandlineprocessors->generic_command_line_processor->readConfigFile->ioutil.ReadFile:" + readfileError.Error() + "|")
	}
	configurationString := string(configBytes)

	configStruct, convertStringToJSONStructError := haymakerutil.ConvertStringToJSONStruct(&configurationString)
	if convertStringToJSONStructError != nil {
		return nil, errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->readConfigFile->haymakerutil.ConvertStringToJSONStruct:" + convertStringToJSONStructError.Error() + "|")
	}

	return configStruct, nil
}

func writeConfigFile(configFilePath *string, config interface{}) error {

	convertStructToJSONStringResult, convertStructToJSONStringError := haymakerutil.ConvertStructToJSONString(config)
	if convertStructToJSONStringError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->writeConfigFile->haymakerutil.ConvertStructToJSONString:" + convertStructToJSONStringError.Error() + "|")
	}

	writeFileError := ioutil.WriteFile(*configFilePath, []byte(*convertStructToJSONStringResult), 0644)
	if writeFileError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->writeConfigFile->ioutil.WriteFile:" + writeFileError.Error() + "|")
	}

	return nil
}

func spinupAWSResources(configFilePath *string) error {

	awsSession, newSessionErr := session.NewSession()
	if newSessionErr != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->spinupAWSResources->session.NewSession:" + newSessionErr.Error() + "|")
	}

	configStruct, readConfigFileError := readConfigFile(configFilePath)
	if readConfigFileError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->spinupAWSResources->readConfigFile:" + readConfigFileError.Error() + "|")
	}

	eksConfig := haymakerutil.BuildEKSConfiguration(configStruct)
	regionConfig := haymakerutil.BuildSessionRegionConfiguration(configStruct)

	eksSession := eks.New(awsSession, aws.NewConfig().WithRegion(*regionConfig))
	haymakerengines.InitEKSEngine(eksSession, eksConfig)

	deleteClusterError := haymakerengines.DeleteClusterAndNodeGroups()
	if deleteClusterError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->spinupAWSResources->haymakerengines.DeleteClusterAndNodeGroups:" + deleteClusterError.Error() + "|")
	}

	ec2Session := ec2.New(awsSession, aws.NewConfig().WithRegion(*regionConfig))
	vpcConfig := haymakerutil.BuildVPCConfiguration(configStruct)
	haymakerengines.InitVPCEngine(ec2Session, vpcConfig)

	elbSession := elb.New(awsSession, aws.NewConfig().WithRegion(*regionConfig))
	haymakerengines.InitELBEngine(elbSession)

	performNetworkConfigCleanupError := haymakerengines.DestroyNetworkResources()
	if performNetworkConfigCleanupError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->spinupAWSResources->haymakerengines.PerformNetworkConfigCleanup:" + performNetworkConfigCleanupError.Error() + "|")
	}

	initializeVPCComponentsError := haymakerengines.CreateNetworkResources(true)
	if initializeVPCComponentsError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->spinupAWSResources->haymakerengines.InitializeVPCComponents:" + initializeVPCComponentsError.Error() + "|")
	}

	haymakerengines.SetSubnetIdsForWorkerNodes(haymakerengines.GetSubnetIdsForEKSWorkerNodes())
	spinupEKSClusterError := haymakerengines.SpinupEKSCluster()
	if spinupEKSClusterError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->spinupAWSResources->haymakerengines.SpinupEKSCluster:" + spinupEKSClusterError.Error() + "|")
	}

	return nil
}

func writeFile(stringToWrite *string, filePath *string) error {

	fileHandle, createError := os.Create(*filePath)
	if createError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->writeFile->os.Create:" + createError.Error() + "|")
	}

	defer fileHandle.Close()

	_, writeError := fileHandle.Write([]byte(*stringToWrite))
	if writeError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->writeFile->file.Write:" + writeError.Error() + "|")
	}

	return nil
}

func generateKubectlFile(configFilePath *string, kubeconfigPath *string) error {

	awsSession, newSessionErr := session.NewSession()
	if newSessionErr != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->generateKubectlFile->session.NewSession:" + newSessionErr.Error() + "|")
	}

	configStruct, readConfigFileError := readConfigFile(configFilePath)
	if readConfigFileError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->generateKubectlFile->readConfigFile:" + readConfigFileError.Error() + "|")
	}

	eksConfig := haymakerutil.BuildEKSConfiguration(configStruct)
	regionConfig := haymakerutil.BuildSessionRegionConfiguration(configStruct)

	eksSession := eks.New(awsSession, aws.NewConfig().WithRegion(*regionConfig))
	haymakerengines.InitEKSEngine(eksSession, eksConfig)

	ec2Session := ec2.New(awsSession, aws.NewConfig().WithRegion(*regionConfig))

	vpcConfig := haymakerutil.BuildVPCConfiguration(configStruct)
	haymakerengines.InitVPCEngine(ec2Session, vpcConfig)

	elbSession := elb.New(awsSession, aws.NewConfig().WithRegion(*regionConfig))
	haymakerengines.InitELBEngine(elbSession)

	getClusterParametersForConfigFileResult, getClusterParametersForConfigFileError := haymakerengines.GetClusterParametersForConfigFile()
	if getClusterParametersForConfigFileError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->generateKubectlFile-> haymakerengines.GetClusterParametersForConfigFile:" + getClusterParametersForConfigFileError.Error() + "|")
	}

	fileWriteError := writeFile(haymakerutil.GenerateKubectlConfFileFromStruct(getClusterParametersForConfigFileResult), kubeconfigPath)
	if fileWriteError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->generateKubectlFile->writeFile:" + fileWriteError.Error() + "|")
	}

	return nil

}

//#is conf necessary?
func destroyAWSResources(configFilePath *string) error {
	awsSession, newSessionErr := session.NewSession()
	if newSessionErr != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->destroyAWSResources->session.NewSession:" + newSessionErr.Error() + "|")
	}

	configStruct, readConfigFileError := readConfigFile(configFilePath)
	if readConfigFileError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->destroyAWSResources->readConfigFile:" + readConfigFileError.Error() + "|")
	}

	eksConfig := haymakerutil.BuildEKSConfiguration(configStruct)
	regionConfig := haymakerutil.BuildSessionRegionConfiguration(configStruct)

	eksSession := eks.New(awsSession, aws.NewConfig().WithRegion(*regionConfig))
	haymakerengines.InitEKSEngine(eksSession, eksConfig)

	deleteClusterError := haymakerengines.DeleteClusterAndNodeGroups()
	if deleteClusterError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->destroyAWSResources->DeleteClusterAndNodeGroups:" + deleteClusterError.Error() + "|")
	}

	ec2Session := ec2.New(awsSession, aws.NewConfig().WithRegion(*regionConfig))
	vpcConfig := haymakerutil.BuildVPCConfiguration(configStruct)
	haymakerengines.InitVPCEngine(ec2Session, vpcConfig)

	elbSession := elb.New(awsSession, aws.NewConfig().WithRegion(*regionConfig))
	haymakerengines.InitELBEngine(elbSession)

	performNetworkConfigCleanupError := haymakerengines.DestroyNetworkResources()
	if performNetworkConfigCleanupError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->destroyAWSResources->PerformNetworkConfigCleanup:" + performNetworkConfigCleanupError.Error() + "|")
	}

	return nil
}

func spinUpService(configFilePath *string) error {

	configStruct, readConfigFileError := readConfigFile(configFilePath)
	if readConfigFileError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->spinUpService->readConfigFile:" + readConfigFileError.Error() + "|")
	}

	kubernetesConfig := haymakerutil.BuildKubernetesConfiguration(configStruct)
	haymakerengines.InitKubernetesEngine(kubernetesConfig)
	haymakerengines.LoadKubeConfig()

	spinupContainersError := haymakerengines.SpinupContainers()

	if spinupContainersError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->spinUpService->haymakerengines.SpinupContainers:" + spinupContainersError.Error() + "|")
	}

	exposeServiceError := haymakerengines.CreateService()
	if exposeServiceError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->spinUpService->haymakerengines.ExposeService:" + exposeServiceError.Error() + "|")
	}

	return nil
}

func describeService(configFilePath *string) error {

	configStruct, readConfigFileError := readConfigFile(configFilePath)
	if readConfigFileError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->describeService->readConfigFile:" + readConfigFileError.Error() + "|")
	}

	kubernetesConfig := haymakerutil.BuildKubernetesConfiguration(configStruct)
	haymakerengines.InitKubernetesEngine(kubernetesConfig)
	haymakerengines.LoadKubeConfig()

	describeServiceError := haymakerengines.DescribeService()
	if describeServiceError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->describeService->haymakerengines.DescribeService:" + describeServiceError.Error() + "|")
	}

	return nil
}

func deleteService(configFilePath *string) error {

	configStruct, readConfigFileError := readConfigFile(configFilePath)
	if readConfigFileError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->deleteService->readConfigFile:" + readConfigFileError.Error() + "|")
	}

	kubernetesConfig := haymakerutil.BuildKubernetesConfiguration(configStruct)
	haymakerengines.InitKubernetesEngine(kubernetesConfig)
	haymakerengines.LoadKubeConfig()

	deleteServiceError := haymakerengines.DeleteService()
	if deleteServiceError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->deleteService->haymakerengines.DeleteService:" + deleteServiceError.Error() + "|")
	}

	deleteContainersError := haymakerengines.DeleteContainers()
	if deleteContainersError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->deleteService->haymakerengines.DeleteContainers:" + deleteContainersError.Error() + "|")
	}

	return nil

}

func spinupECR(configFilePath *string) error {

	awsSession, newSessionErr := session.NewSession()
	if newSessionErr != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->spinupECR->session.NewSession:" + newSessionErr.Error() + "|")
	}

	configStruct, readConfigFileError := readConfigFile(configFilePath)
	if readConfigFileError != nil {
		return errors.New("|" + "HayMaker->main->spinupHaymaker->spinupECR->readConfigFile:" + readConfigFileError.Error() + "|")
	}

	regionConfig := haymakerutil.BuildSessionRegionConfiguration(configStruct)

	ecrConfig := haymakerutil.BuildECRConfiguration(configStruct)
	ecrSession := ecr.New(awsSession, aws.NewConfig().WithRegion(*regionConfig))
	dockerConfig := haymakerutil.BuildDockerConfiguration(configStruct)

	haymakerengines.InitECREngine(ecrSession, ecrConfig, dockerConfig)

	deleteClusterError := haymakerengines.DestroyECRRepository()
	if deleteClusterError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->spinupECR->haymakerengines.DeleteContainerRepository:" + deleteClusterError.Error() + "|")
	}

	spinupContainerRepositoryError := haymakerengines.SpinupContainerRepository()
	if spinupContainerRepositoryError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->spinupECR->haymakerengines.SpinupContainerRepository:" + spinupContainerRepositoryError.Error() + "|")
	}

	return nil
}

func deleteECR(configFilePath *string) error {

	awsSession, newSessionErr := session.NewSession()
	if newSessionErr != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->deleteECR->session.NewSession:" + newSessionErr.Error() + "|")
	}
	configStruct, readConfigFileError := readConfigFile(configFilePath)
	if readConfigFileError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->deleteECR->spinupHaymaker->readConfigFile:" + readConfigFileError.Error() + "|")
	}

	regionConfig := haymakerutil.BuildSessionRegionConfiguration(configStruct)
	ecrConfig := haymakerutil.BuildECRConfiguration(configStruct)
	dockerConfig := haymakerutil.BuildDockerConfiguration(configStruct)

	ecrSession := ecr.New(awsSession, aws.NewConfig().WithRegion(*regionConfig))
	haymakerengines.InitECREngine(ecrSession, ecrConfig, dockerConfig)

	deleteClusterError := haymakerengines.DestroyECRRepository()
	if deleteClusterError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->deleteECR->haymakerengines.DeleteContainerRepository:" + deleteClusterError.Error() + "|")
	}

	return nil
}

func deleteImageFromECRAndLocalDocker(configFilePath *string) error {

	awsSession, newSessionErr := session.NewSession()
	if newSessionErr != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->deleteImageFromECRAndLocalDocker->session.NewSession:" + newSessionErr.Error() + "|")
	}
	configStruct, readConfigFileError := readConfigFile(configFilePath)
	if readConfigFileError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->deleteImageFromECRAndLocalDocker->readConfigFile:" + readConfigFileError.Error() + "|")
	}

	regionConfig := haymakerutil.BuildSessionRegionConfiguration(configStruct)
	ecrConfig := haymakerutil.BuildECRConfiguration(configStruct)
	dockerConfig := haymakerutil.BuildDockerConfiguration(configStruct)

	ecrSession := ecr.New(awsSession, aws.NewConfig().WithRegion(*regionConfig))
	haymakerengines.InitECREngine(ecrSession, ecrConfig, dockerConfig)

	deleteImageFromECRError := haymakerengines.DestroyDockerImageOnECR()
	if deleteImageFromECRError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->deleteImageFromECRAndLocalDocker->haymakerengines.DeleteImageFromECR:" + deleteImageFromECRError.Error() + "|")
	}

	deleteImageFromLocalDockerError := deleteImageFromLocalDocker(configFilePath)
	if deleteImageFromLocalDockerError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->deleteImageFromECRAndLocalDocker->deleteImageFromLocalDocker:" + deleteImageFromLocalDockerError.Error() + "|")
	}

	return nil
}

func deleteImageFromLocalDocker(configFilePath *string) error {

	configStruct, readConfigFileError := readConfigFile(configFilePath)
	if readConfigFileError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->deleteImageFromLocalDocker->readConfigFile:" + readConfigFileError.Error() + "|")
	}

	dockerConfig := haymakerutil.BuildDockerConfiguration(configStruct)

	haymakerengines.InitDockerEngine(dockerConfig)

	deleteLocalImagesError := haymakerengines.DeleteLocalImages()
	if deleteLocalImagesError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->deleteImageFromLocalDocker->haymakerengines.DeleteLocalImages:" + deleteLocalImagesError.Error() + "|")
	}

	return nil
}

func updateLocalConfiguration(configFilePath *string) error {

	configStruct, readConfigFileError := readConfigFile(configFilePath)
	if readConfigFileError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->updateLocalConfiguration->readConfigFile:" + readConfigFileError.Error() + "|")
	}

	getAuthorizationTokenResult, getAuthorizationTokenError := haymakerengines.GetAuthorizationToken()
	if getAuthorizationTokenError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->updateLocalConfiguration->haymakerengines.GetAuthorizationToken:" + getAuthorizationTokenError.Error() + "|")
	}

	updateECRHostOnConfigurationResult := haymakerutil.UpdateECRHostOnConfiguration(configStruct, getAuthorizationTokenResult)

	writeConfigFileError := writeConfigFile(configFilePath, updateECRHostOnConfigurationResult)
	if writeConfigFileError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->updateLocalConfiguration->writeConfigFile:" + writeConfigFileError.Error() + "|")
	}

	return nil
}

func pushImageToRegistry(configFilePath *string) error {

	awsSession, newSessionErr := session.NewSession()
	if newSessionErr != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->pushImageToRegistry->session.NewSession:" + newSessionErr.Error() + "|")
	}

	configStruct, readConfigFileError := readConfigFile(configFilePath)
	if readConfigFileError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->pushImageToRegistry->readConfigFile:" + readConfigFileError.Error() + "|")
	}
	///////////////

	/*
		This is duplicated code but the reason is the following: i need to update the disk configuration with the latest creds.
		For that i need to use ECR and so i need to initialize the Engine. I parse docker conf because the configuration checks should be made
		on the Build[SERVICE]Configuration. The Init functions assume some fields. Once i update the config, i then proceed to redo all these
		steps for consistency.
	*/
	regionConfig := haymakerutil.BuildSessionRegionConfiguration(configStruct)
	ecrConfig := haymakerutil.BuildECRConfiguration(configStruct)
	ecrSession := ecr.New(awsSession, aws.NewConfig().WithRegion(*regionConfig))
	dockerConfig := haymakerutil.BuildDockerConfiguration(configStruct)

	haymakerengines.InitECREngine(ecrSession, ecrConfig, dockerConfig)

	updateLocalConfigurationError := updateLocalConfiguration(configFilePath)
	if updateLocalConfigurationError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->pushImageToRegistry->updateLocalConfiguration:" + updateLocalConfigurationError.Error() + "|")
	}

	///////////////

	configStruct, readConfigFileError = readConfigFile(configFilePath)
	if readConfigFileError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->pushImageToRegistry->readConfigFile:" + readConfigFileError.Error() + "|")
	}

	regionConfig = haymakerutil.BuildSessionRegionConfiguration(configStruct)
	ecrConfig = haymakerutil.BuildECRConfiguration(configStruct)
	ecrSession = ecr.New(awsSession, aws.NewConfig().WithRegion(*regionConfig))
	dockerConfig = haymakerutil.BuildDockerConfiguration(configStruct)

	haymakerengines.InitECREngine(ecrSession, ecrConfig, dockerConfig)

	haymakerengines.InitDockerEngine(dockerConfig)

	buildImageAndPushError := haymakerengines.BuildImageAndPush()
	if buildImageAndPushError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->pushImageToRegistry->haymakerengines.BuildImageAndPush:" + buildImageAndPushError.Error() + "|")
	}

	return nil
}

func deployHaymakerFull(configFilePath *string, kubeconfigPath *string) error {

	spinupHaymakerError := spinupAWSResources(configFilePath)
	if spinupHaymakerError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->deployHaymakerFull->spinupAWSResources:" + spinupHaymakerError.Error() + "|")
	}

	generateKubectlFileError := generateKubectlFile(configFilePath, kubeconfigPath)
	if generateKubectlFileError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->deployHaymakerFull->generateKubectlFile:" + generateKubectlFileError.Error() + "|")
	}

	deleteECRrror := deleteECR(configFilePath)
	if deleteECRrror != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->deployHaymakerFull->deleteECR:" + deleteECRrror.Error() + "|")
	}

	spinupECRError := spinupECR(configFilePath)
	if spinupECRError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->deployHaymakerFull->spinupECR:" + spinupECRError.Error() + "|")
	}

	updateLocalConfigurationError := updateLocalConfiguration(configFilePath)
	if updateLocalConfigurationError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->deployHaymakerFull->updateLocalConfiguration:" + updateLocalConfigurationError.Error() + "|")
	}

	deleteImageFromLocalDockerError := deleteImageFromLocalDocker(configFilePath)
	if deleteImageFromLocalDockerError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->deployHaymakerFull->deleteImageFromLocalDocker:" + deleteImageFromLocalDockerError.Error() + "|")
	}

	pushImageToRegistryerror := pushImageToRegistry(configFilePath)
	if pushImageToRegistryerror != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->deployHaymakerFull->pushImageToRegistry:" + pushImageToRegistryerror.Error() + "|")
	}

	spinUpServiceError := spinUpService(configFilePath)
	if spinUpServiceError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->deployHaymakerFull->spinUpService:" + spinUpServiceError.Error() + "|")
	}

	return nil

}
func destroyHaymakerFull(configFilePath *string, kubeconfigPath *string) error {

	shutdownAWSResourcesError := destroyAWSResources(configFilePath)
	if shutdownAWSResourcesError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->destroyHaymakerFull->destroyAWSResources:" + shutdownAWSResourcesError.Error() + "|")
	}

	deleteECRrror := deleteECR(configFilePath)
	if deleteECRrror != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->destroyHaymakerFull->deleteECR:" + deleteECRrror.Error() + "|")
	}

	deleteImageFromLocalDockerError := deleteImageFromLocalDocker(configFilePath)
	if deleteImageFromLocalDockerError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->destroyHaymakerFull->deleteImageFromLocalDocker:" + deleteImageFromLocalDockerError.Error() + "|")
	}

	return nil
}

func describeRepository(configFilePath *string) error {

	awsSession, newSessionErr := session.NewSession()
	if newSessionErr != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->describeRepository->session.NewSession:" + newSessionErr.Error() + "|")
	}

	configStruct, readConfigFileError := readConfigFile(configFilePath)
	if readConfigFileError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->describeRepository->readConfigFile:" + readConfigFileError.Error() + "|")
	}

	regionConfig := haymakerutil.BuildSessionRegionConfiguration(configStruct)

	ecrConfig := haymakerutil.BuildECRConfiguration(configStruct)
	ecrSession := ecr.New(awsSession, aws.NewConfig().WithRegion(*regionConfig))
	dockerConfig := haymakerutil.BuildDockerConfiguration(configStruct)

	haymakerengines.InitECREngine(ecrSession, ecrConfig, dockerConfig)
	/*
		kubernetesConfig := haymakerutil.BuildKubernetesConfiguration(configStruct)

		haymakerengines.InitKubernetesEngine(kubernetesConfig)

		haymakerengines.LoadKubeConfig()
	*/
	describeRepositoriesError := haymakerengines.DescribeRepositories()
	if describeRepositoriesError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->describeRepository->haymakerengines.DescribeRepositories:" + describeRepositoriesError.Error() + "|")
	}

	return nil
}

func ProcessCommandLine(commandLineMap map[string]*string) error {

	userHomeDirResult, userHomeDirError := os.UserHomeDir()
	if userHomeDirError != nil {
		return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->ProcessCommandLine->os.UserHomeDir:" + userHomeDirError.Error() + "|")
	}
	defaultKubeconfigFile = fmt.Sprintf("%s/%s", userHomeDirResult, defaultKubeconfigPathWithinHome)

	if opt, ok := commandLineMap["option"]; ok {
		if confFile, ok := commandLineMap["config_file"]; ok {
			switch *opt {
			case "sh":
				fmt.Println("Spinup Haymaker Resources")
				spinupHaymakerError := spinupAWSResources(confFile)
				if spinupHaymakerError != nil {
					return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->ProcessCommandLine->spinupAWSResources:" + spinupHaymakerError.Error() + "|")
				}
			case "dh":
				fmt.Println("Destroy Haymaker Resources")
				shutdownAWSResourcesError := destroyAWSResources(confFile)
				if shutdownAWSResourcesError != nil {
					return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->ProcessCommandLine->destroyAWSResources:" + shutdownAWSResourcesError.Error() + "|")
				}
			case "gk":
				fmt.Println("Generate Kubeconfig File")
				var kubeConfigTempString string
				if kubeConfig, ok := commandLineMap["kubeconfig_file"]; ok && *kubeConfig != "" {
					kubeConfigTempString = *kubeConfig
				} else {
					kubeConfigTempString = defaultKubeconfigFile
				}
				generateKubectlFileError := generateKubectlFile(confFile, &kubeConfigTempString)
				if generateKubectlFileError != nil {
					return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->ProcessCommandLine->generateKubectlFile:" + generateKubectlFileError.Error() + "|")
				}
			case "sr":
				fmt.Println("Spinup Repository")
				spinupECRError := spinupECR(confFile)
				if spinupECRError != nil {
					return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->ProcessCommandLine->spinupECR:" + spinupECRError.Error() + "|")
				}
			case "dr":
				fmt.Println("Destroy Repository")
				deleteECRError := deleteECR(confFile)
				if deleteECRError != nil {
					return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->ProcessCommandLine->deleteECR:" + deleteECRError.Error() + "|")
				}
			case "pr":
				fmt.Println("Print Repository Information")
				describeRepositoryError := describeRepository(confFile)
				if describeRepositoryError != nil {
					return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->ProcessCommandLine->describeRepository:" + describeRepositoryError.Error() + "|")
				}
			case "sc":
				fmt.Println("Deploy Container And Create Service")
				spinUpServiceError := spinUpService(confFile)
				if spinUpServiceError != nil {
					return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->ProcessCommandLine->spinUpService:" + spinUpServiceError.Error() + "|")
				}
			case "dc":
				fmt.Println("Delete Container And Service")
				deleteServiceError := deleteService(confFile)
				if deleteServiceError != nil {
					return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->ProcessCommandLine->deleteService:" + deleteServiceError.Error() + "|")
				}
			case "is":
				fmt.Println("Print Information About Service")
				describeServiceError := describeService(confFile)
				if describeServiceError != nil {
					return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->ProcessCommandLine->describeService:" + describeServiceError.Error() + "|")
				}
			case "pi":
				fmt.Println("Build and Push Docker Image To Registry")
				pushImageToRegistryError := pushImageToRegistry(confFile)
				if pushImageToRegistryError != nil {
					return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->ProcessCommandLine->pushImageToRegistry:" + pushImageToRegistryError.Error() + "|")
				}
			case "di":
				fmt.Println("Delete Local and ECR Docker Image")
				deleteImageFromECRError := deleteImageFromECRAndLocalDocker(confFile)
				if deleteImageFromECRError != nil {
					return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->ProcessCommandLine->deleteImageFromECRAndLocalDocker:" + deleteImageFromECRError.Error() + "|")
				}
			case "dli":
				fmt.Println("Delete Local Docker image")
				deleteImageFromLocalDockerError := deleteImageFromLocalDocker(confFile)
				if deleteImageFromLocalDockerError != nil {
					return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->ProcessCommandLine->deleteImageFromLocalDocker:" + deleteImageFromLocalDockerError.Error() + "|")
				}
			case "fs":
				fmt.Println("Full Haymaker Spinup")
				var kubeConfigTempString string
				if kubeConfig, ok := commandLineMap["kubeconfig_file"]; ok && *kubeConfig != "" {
					kubeConfigTempString = *kubeConfig
				} else {
					kubeConfigTempString = defaultKubeconfigFile
				}
				deployHaymakerFullError := deployHaymakerFull(confFile, &kubeConfigTempString)
				if deployHaymakerFullError != nil {
					return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->ProcessCommandLine->deployHaymakerFull:" + deployHaymakerFullError.Error() + "|")
				}
			case "fd":
				fmt.Println("Full Haymaker Destroy")
				var kubeConfigTempString string
				if kubeConfig, ok := commandLineMap["kubeconfig_file"]; ok && *kubeConfig != "" {
					kubeConfigTempString = *kubeConfig
				} else {
					kubeConfigTempString = defaultKubeconfigFile
				}
				destroyHaymakerFullError := destroyHaymakerFull(confFile, &kubeConfigTempString)
				if destroyHaymakerFullError != nil {
					return errors.New("|" + "HayMaker->commandlineprocessors->generic_command_line_processor->ProcessCommandLine->destroyHaymakerFull:" + destroyHaymakerFullError.Error() + "|")
				}
			default:
				fmt.Println("Invalid option")
				fmt.Println(globalstringsproviders.GetMenuPictureStringWithOptions())
			}
		}
	}

	return nil

}
