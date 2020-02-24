package haymakerengines

import (
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/eks"
)

var eksInstance *eks.EKS

var eksClusterName *string
var eksArn *string
var nodeRoleArn *string
var workerNodeGroupName *string
var instanceType *string
var amiType *string
var scallingDesiredSize int64
var scallingMaxSize int64
var scallingMinSize int64
var diskSize int64

var subnetIdsForWorkerNodes []*string
var apiServerEndpoint *string

func createClusterStub(subnetIdsForWorkerNodesL []*string, eksClusterNameL *string, eksArnL *string) (*eks.CreateClusterOutput, error) {

	vpcConfigRequestObject := &eks.VpcConfigRequest{
		EndpointPrivateAccess: aws.Bool(true),
		EndpointPublicAccess:  aws.Bool(true),
		SubnetIds:             subnetIdsForWorkerNodesL,
	}

	createCacheClusterInputConfigObject := &eks.CreateClusterInput{
		Name:               aws.String(*eksClusterNameL),
		RoleArn:            aws.String(*eksArnL),
		ResourcesVpcConfig: vpcConfigRequestObject,
	}

	return eksInstance.CreateCluster(createCacheClusterInputConfigObject)
}

func createNodeGroupStub(subnetIdsForWorkerNodesL []*string, workerNodeGroupNameL *string, eksClusterNameL *string, amiTypeL *string, instanceTypeL *string, scallingDesiredSizeL int64, nodeRoleArnL *string, scallingMaxSizeL int64, scallingMinSizeL int64, diskSizeL int64) (*eks.CreateNodegroupOutput, error) {

	instanceTypes := []*string{instanceTypeL}

	nodegroupScalingConfigObject := &eks.NodegroupScalingConfig{
		DesiredSize: aws.Int64(scallingDesiredSizeL),
		MaxSize:     aws.Int64(scallingMaxSizeL),
		MinSize:     aws.Int64(scallingMinSizeL),
	}

	createNodeGroupInputObject := &eks.CreateNodegroupInput{
		AmiType:       amiTypeL,
		ClusterName:   eksClusterNameL,
		DiskSize:      aws.Int64(diskSizeL),
		InstanceTypes: instanceTypes,
		NodeRole:      nodeRoleArnL,
		NodegroupName: workerNodeGroupNameL,
		ScalingConfig: nodegroupScalingConfigObject,
		Subnets:       subnetIdsForWorkerNodesL,
	}

	return eksInstance.CreateNodegroup(createNodeGroupInputObject)
}

func deleteClusterStub(eksClusterNameL *string) (*eks.DeleteClusterOutput, error) {

	deleteClusterInputConfigObject := &eks.DeleteClusterInput{
		Name: aws.String(*eksClusterNameL),
	}

	return eksInstance.DeleteCluster(deleteClusterInputConfigObject)
}

func deleteNodeGroupStub(eksClusterNameL *string, workerNodeGroupNameL *string) (*eks.DeleteNodegroupOutput, error) {

	deleteNodegroupInputObject := &eks.DeleteNodegroupInput{
		ClusterName:   aws.String(*eksClusterNameL),
		NodegroupName: aws.String(*workerNodeGroupNameL),
	}

	return eksInstance.DeleteNodegroup(deleteNodegroupInputObject)
}

func describeClusterStub(eksClusterNameL *string) (*eks.DescribeClusterOutput, error) {

	describeClusterObject := &eks.DescribeClusterInput{
		Name: aws.String(*eksClusterNameL),
	}

	return eksInstance.DescribeCluster(describeClusterObject)
}

func describeNodeGroupStub(eksClusterNameL *string, workerNodeGroupNameL *string) (*eks.DescribeNodegroupOutput, error) {

	describeNodeGroupObject := &eks.DescribeNodegroupInput{
		ClusterName:   aws.String(*eksClusterNameL),
		NodegroupName: aws.String(*workerNodeGroupNameL),
	}

	return eksInstance.DescribeNodegroup(describeNodeGroupObject)
}

func SetSubnetIdsForWorkerNodes(subnetIds []*string) {

	subnetIdsForWorkerNodes = subnetIds

}

func destroyCluster() error {

	fmt.Println("Destroying EKS Cluster")
	describeClusterResult, describeCacheClustersErr := describeClusterStub(eksClusterName)

	if describeCacheClustersErr != nil {
		if aerr, ok := describeCacheClustersErr.(awserr.Error); ok {
			switch aerr.Code() {
			case eks.ErrCodeResourceNotFoundException:
				//It means the cluster does not exist
				return nil
			}
		}
	}

	//We check if it is running. We may catch it in the middle of deployment...
	for true {
		if (describeClusterResult != nil) && ((describeClusterResult.Cluster) != nil) && *describeClusterResult.Cluster.Status == eks.ClusterStatusActive {
			break
		}
		//It seems that it takes a bit to propagate and if i try to delete the VPCs right after i get dependency error.
		time.Sleep(20 * time.Second)
		describeClusterResult, describeCacheClustersErr = describeClusterStub(eksClusterName)

	}

	_, deleteClusterError := deleteClusterStub(eksClusterName)
	if deleteClusterError != nil {
		return errors.New("|" + "HayMaker->haymakerengines->eks_engine->destroyCluster->deleteClusterStub:" + deleteClusterError.Error() + "|")
	}

	_, describeCacheClustersErr = describeClusterStub(eksClusterName)

	for true {

		if describeCacheClustersErr != nil {
			if aerr, ok := describeCacheClustersErr.(awserr.Error); ok {
				if aerr.Code() == eks.ErrCodeResourceNotFoundException {
					break
				}
			}
		}

		time.Sleep(20 * time.Second)
		_, describeCacheClustersErr = describeClusterStub(eksClusterName)

	}

	return nil
}

func destroyNodeGroup() error {

	fmt.Println("Destroying EKS Nodes")

	describeNodeGroupResult, describeNodeGroupErr := describeNodeGroupStub(eksClusterName, workerNodeGroupName)
	if describeNodeGroupErr != nil {
		if aerr, ok := describeNodeGroupErr.(awserr.Error); ok {
			switch aerr.Code() {
			case eks.ErrCodeResourceNotFoundException:
				//It means the cluster does not exist
				return nil
			}
		}
	}

	//We check if it is running. We may catch it in the middle of deployment...
	for true {
		if (describeNodeGroupResult != nil) && ((describeNodeGroupResult.Nodegroup) != nil) && *describeNodeGroupResult.Nodegroup.Status == eks.NodegroupStatusActive {
			break
		}
		//It seems that it takes a bit to propagate and if i try to delete the VPCs right after i get dependency error.
		time.Sleep(20 * time.Second)
		describeNodeGroupResult, describeNodeGroupErr = describeNodeGroupStub(eksClusterName, workerNodeGroupName)

	}

	_, deleteNodeError := deleteNodeGroupStub(eksClusterName, workerNodeGroupName)
	if deleteNodeError != nil {
		return errors.New("|" + "HayMaker->haymakerengines->eks_engine->destroyNodeGroup->deleteNodeGroupStub:" + deleteNodeError.Error() + "|")
	}

	_, describeNodeGroupErr = describeNodeGroupStub(eksClusterName, workerNodeGroupName)

	for true {

		if describeNodeGroupErr != nil {
			if aerr, ok := describeNodeGroupErr.(awserr.Error); ok {
				if aerr.Code() == eks.ErrCodeResourceNotFoundException {
					break
				}
			}
		}

		time.Sleep(20 * time.Second)
		_, describeNodeGroupErr = describeNodeGroupStub(eksClusterName, workerNodeGroupName)

	}

	return nil
}

func DeleteClusterAndNodeGroups() error {

	deleteNodeGroupError := destroyNodeGroup()
	if deleteNodeGroupError != nil {
		return errors.New("|" + "HayMaker->haymakerengines->eks_engine->DeleteClusterAndNodeGroups->deleteNodeGroup:" + deleteNodeGroupError.Error() + "|")
	}

	deleteClusterError := destroyCluster()
	if deleteClusterError != nil {
		return errors.New("|" + "HayMaker->haymakerengines->eks_engine->DeleteClusterAndNodeGroups->deleteCluster:" + deleteClusterError.Error() + "|")
	}

	return nil
}
func waitUntilEKSClusterAvailable() error {

	fmt.Println(fmt.Sprintf("Waiting For Cluster: %s to become available.", *eksClusterName))
	for true {
		describeCacheClusterResult, describeCacheClusterErr := describeClusterStub(eksClusterName)
		if describeCacheClusterErr != nil {
			return errors.New("|" + "HayMaker->haymakerengines->eks_engine->waitUntilEKSClusterAvailable->describeClusterStub:" + describeCacheClusterErr.Error() + "|")
		} else if *describeCacheClusterResult.Cluster.Status != eks.ClusterStatusActive {
			time.Sleep(20 * time.Second)
		} else {
			break
		}
	}
	return nil
}

func waitUntilNodesAvailable() error {

	fmt.Println(fmt.Sprintf("Waiting For Nodes: %s to become available.", *eksClusterName))
	for true {

		describeNodeGroupResult, describeNodeGroupErr := describeNodeGroupStub(eksClusterName, workerNodeGroupName)
		if describeNodeGroupErr != nil {
			return errors.New("|" + "HayMaker->haymakerengines->eks_engine->waitUntilNodesAvailable->describeNodeGroupStub:" + describeNodeGroupErr.Error() + "|")
		} else if *describeNodeGroupResult.Nodegroup.Status != eks.NodegroupStatusActive {
			time.Sleep(20 * time.Second)
		} else {
			break
		}
	}

	return nil
}

func GetClusterParametersForConfigFile() (map[string]interface{}, error) {

	var parameters map[string]interface{} = make(map[string]interface{}, 0)

	describeClusterResult, describeClusterErr := describeClusterStub(eksClusterName)
	if describeClusterErr != nil {
		return parameters, errors.New("|" + "HayMaker->haymakerengines->eks_engine->GetClusterParametersForConfigFile->describeClusterStub:" + describeClusterErr.Error() + "|")
	}

	parameters["server"] = describeClusterResult.Cluster.Endpoint
	parameters["cert"] = describeClusterResult.Cluster.CertificateAuthority.Data
	parameters["name"] = describeClusterResult.Cluster.Name

	return parameters, nil
}

func SpinupEKSCluster() error {

	_, createClusterError := createClusterStub(subnetIdsForWorkerNodes, eksClusterName, eksArn)
	if createClusterError != nil {
		return errors.New("|" + "HayMaker->haymakerengines->eks_engine->SpinupEKSCluster->createClusterStub:" + createClusterError.Error() + "|")
	}
	waitUntilEKSClusterAvailable()

	//For some silly reason the Endpoint returned by CreateCluster contains a null pointer for endpoint...
	describeClusterResult, describeClusterErr := describeClusterStub(eksClusterName)
	if createClusterError != nil {
		return errors.New("|" + "HayMaker->haymakerengines->eks_engine->SpinupEKSCluster->describeClusterStub:" + describeClusterErr.Error() + "|")
	}

	apiServerEndpoint = describeClusterResult.Cluster.Endpoint
	fmt.Println(fmt.Sprintf("Creating Node Group..."))
	_, createNodeGroupErr := createNodeGroupStub(subnetIdsForWorkerNodes, workerNodeGroupName, eksClusterName, amiType, instanceType, scallingDesiredSize, nodeRoleArn, scallingMaxSize, scallingMinSize, diskSize)
	if createNodeGroupErr != nil {
		return errors.New("|" + "HayMaker->haymakerengines->eks_engine->SpinupEKSCluster->createNodeGroupStub:" + createNodeGroupErr.Error() + "|")
	}

	waitUntilNodesAvailable()

	return nil
}

func InitEKSEngine(eksIns *eks.EKS, eksConfStruct interface{}) {

	eksInstance = eksIns

	eksClusterNameTemp := (((eksConfStruct.(map[string]interface{}))["eks_cluster_name"]).(string))
	eksClusterName = &eksClusterNameTemp

	eksArnTemp := (((eksConfStruct.(map[string]interface{}))["eks_role_arn"]).(string))
	eksArn = &eksArnTemp

	nodeRoleArnTemp := (((eksConfStruct.(map[string]interface{}))["worker_node_role_arn"]).(string))
	nodeRoleArn = &nodeRoleArnTemp

	workerNodeGroupNameTemp := (((eksConfStruct.(map[string]interface{}))["worker_node_group_name"]).(string))
	workerNodeGroupName = &workerNodeGroupNameTemp

	instanceTypeTemp := (((eksConfStruct.(map[string]interface{}))["instance_type"]).(string))
	instanceType = &instanceTypeTemp

	amiTypeTemp := (((eksConfStruct.(map[string]interface{}))["ami_type"]).(string))
	amiType = &amiTypeTemp

	diskSize = int64((eksConfStruct.(map[string]interface{})["disk_size"]).(float64))

	scallingDesiredSize = int64((eksConfStruct.(map[string]interface{})["desired_size"]).(float64))
	scallingMaxSize = int64((eksConfStruct.(map[string]interface{})["max_size"]).(float64))
	scallingMinSize = int64((eksConfStruct.(map[string]interface{})["min_size"]).(float64))

}
