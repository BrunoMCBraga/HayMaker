package haymakerutil

import (
	"fmt"
)

//Put some config checks...
func BuildEKSConfiguration(configStruct interface{}) interface{} {

	cast := configStruct.(map[string]interface{})

	if eks, ok := cast["eks"]; ok {

		return eks
	}

	return nil

}

func BuildVPCConfiguration(configStruct interface{}) interface{} {

	cast := configStruct.(map[string]interface{})

	if vpc, ok := cast["vpc"]; ok {

		return vpc
	}

	return nil

}

func BuildECRConfiguration(configStruct interface{}) interface{} {

	cast := configStruct.(map[string]interface{})

	if ecr, ok := cast["ecr"]; ok {

		return ecr
	}

	return nil

}

func BuildKubernetesConfiguration(configStruct interface{}) interface{} {

	cast := configStruct.(map[string]interface{})

	if kubernetes, ok := cast["kubernetes"]; ok {

		return kubernetes
	}

	return nil

}

func BuildSessionRegionConfiguration(configStruct interface{}) *string {

	cast := configStruct.(map[string]interface{})

	if sessionRegion, ok := cast["session_region"]; ok {

		regionString := sessionRegion.(string)
		return &regionString
	}

	return nil

}

func BuildDockerConfiguration(configStruct interface{}) interface{} {

	cast := configStruct.(map[string]interface{})

	if docker, ok := cast["docker"]; ok {

		return docker
	}

	return nil

}

func UpdateECRHostOnConfiguration(configStruct interface{}, authorizationToken map[string]*string) interface{} {

	cast := configStruct.(map[string]interface{})

	if kubernetes, ok := cast["kubernetes"]; ok {
		if docker, ok := cast["docker"]; ok {
			if ecr, ok := cast["ecr"]; ok {

				ecrRepoName := ecr.(map[string]interface{})["repo_name"].(string)

				dockerTag := docker.(map[string]interface{})["tag"].(string)
				kubernetes.(map[string]interface{})["deployment"].(map[string]interface{})["image_name"] = fmt.Sprintf("%s/%s:%s", *authorizationToken["endpoint"], ecrRepoName, dockerTag)

				docker.(map[string]interface{})["aws"].(map[string]interface{})["authorization_token"] = *authorizationToken["token"]
				docker.(map[string]interface{})["aws"].(map[string]interface{})["ecr_repo_host"] = *authorizationToken["endpoint"]

			}
		}

	}

	return configStruct

}
