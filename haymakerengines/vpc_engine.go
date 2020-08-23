package haymakerengines

import (
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/aws/aws-sdk-go/aws"
)

var vpcInstance *ec2.EC2

var vpcCIDR *string
var privateNetworkConfigStruct []interface{}
var publicNetworkConfigStruct []interface{}
var haymakerVPCTags map[string]interface{}
var securityGroupTags map[string]interface{}

const destinationCIDRForDefaultGateway = "0.0.0.0/0"

var vpcid *string
var internetGatewayID *string
var subnetIds map[string]*string = make(map[string]*string, 0)
var routeTableIDs map[string]*string = make(map[string]*string, 0)

var sleepDelay time.Duration = time.Duration(20 * time.Second)

func createTagsStub(resources []*string, tags *[]*ec2.Tag) (*ec2.CreateTagsOutput, error) {

	createTagsInputObject := &ec2.CreateTagsInput{
		Resources: resources,
		Tags:      *tags,
	}

	return vpcInstance.CreateTags(createTagsInputObject)

}

func createInternetGatewayStub(tags *[]*ec2.Tag) (*ec2.CreateInternetGatewayOutput, error) {

	createInternetGatewayObject := &ec2.CreateInternetGatewayInput{}

	createInternetGatewayOutput, createInternetGatewayErr := vpcInstance.CreateInternetGateway(createInternetGatewayObject)

	resources := []*string{createInternetGatewayOutput.InternetGateway.InternetGatewayId}

	_, createTagsErr := createTagsStub(resources, tags)
	if createTagsErr != nil {
		return createInternetGatewayOutput, errors.New("|" + "HayMaker->haymakerengines->vpc_engine->InitializeVPCComponents->CreateTagsStub:" + createTagsErr.Error() + "|")
	}

	return createInternetGatewayOutput, createInternetGatewayErr

}

func describeInternetGatewaysStub(filters []*ec2.Filter) (*ec2.DescribeInternetGatewaysOutput, error) {

	describeInternetGatewaysObject := &ec2.DescribeInternetGatewaysInput{
		Filters: filters,
	}

	return vpcInstance.DescribeInternetGateways(describeInternetGatewaysObject)

}

func deleteInternetGatewayStub(internetGatewayID *string) (*ec2.DeleteInternetGatewayOutput, error) {

	deleteInternetGatewayObject := &ec2.DeleteInternetGatewayInput{
		InternetGatewayId: internetGatewayID,
	}

	return vpcInstance.DeleteInternetGateway(deleteInternetGatewayObject)

}

func detachInternetGatewayStub(vpcID *string, internetGatewayID *string) (*ec2.DetachInternetGatewayOutput, error) {

	detachInternetGatewayObject := &ec2.DetachInternetGatewayInput{
		VpcId:             vpcID,
		InternetGatewayId: internetGatewayID,
	}

	return vpcInstance.DetachInternetGateway(detachInternetGatewayObject)

}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func createRouteTableStub(vpcID *string, tags *[]*ec2.Tag) (*ec2.CreateRouteTableOutput, error) {

	createRouteTableInputObject := &ec2.CreateRouteTableInput{
		VpcId: vpcID,
	}

	createRouteTableOutput, createRouteTableErr := vpcInstance.CreateRouteTable(createRouteTableInputObject)

	resources := []*string{createRouteTableOutput.RouteTable.RouteTableId}

	_, createTagsErr := createTagsStub(resources, tags)
	if createTagsErr != nil {
		return createRouteTableOutput, errors.New("|" + "HayMaker->haymakerengines->vpc_engine->InitializeVPCComponents->CreateTagsStub:" + createTagsErr.Error() + "|")
	}

	return createRouteTableOutput, createRouteTableErr

}

func deleteRouteTableStub(routeTableID *string) (*ec2.DeleteRouteTableOutput, error) {

	deleteRouteTableInputObject := &ec2.DeleteRouteTableInput{
		RouteTableId: routeTableID,
	}

	return vpcInstance.DeleteRouteTable(deleteRouteTableInputObject)

}

func describeRouteTableStub(filters []*ec2.Filter) (*ec2.DescribeRouteTablesOutput, error) {

	describeRouteTablesInputObject := &ec2.DescribeRouteTablesInput{
		Filters: filters,
	}

	return vpcInstance.DescribeRouteTables(describeRouteTablesInputObject)

}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func createRouteStub(desinationCIDRIP *string, internetGatewayID *string, routeTableID *string, tags *[]*ec2.Tag) (*ec2.CreateRouteOutput, error) {

	createRouteObject := &ec2.CreateRouteInput{
		DestinationCidrBlock: desinationCIDRIP,
		GatewayId:            internetGatewayID,
		RouteTableId:         routeTableID,
	}

	return vpcInstance.CreateRoute(createRouteObject)

}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func modifyVpcAttributeStub(VPCID *string, modifyVpcAttributeInput *ec2.ModifyVpcAttributeInput) (*ec2.ModifyVpcAttributeOutput, error) {

	modifyVpcAttributeInput.VpcId = VPCID

	return vpcInstance.ModifyVpcAttribute(modifyVpcAttributeInput)

}

func enableVPCDNS(VPCID *string) error {

	attributeBooleanValue := &ec2.AttributeBooleanValue{
		Value: aws.Bool(true),
	}

	modifyVpcAttributeInputObject := &ec2.ModifyVpcAttributeInput{
		EnableDnsHostnames: attributeBooleanValue,
	}

	_, modifyVpcAttributeError := modifyVpcAttributeStub(VPCID, modifyVpcAttributeInputObject)

	if modifyVpcAttributeError != nil {
		return errors.New("|" + "HayMaker->haymakerengines->vpc_engine->InitializeVPCComponents->enableVPCDNS:" + modifyVpcAttributeError.Error() + "|")
	}

	modifyVpcAttributeInputObject = &ec2.ModifyVpcAttributeInput{
		EnableDnsSupport: attributeBooleanValue,
	}

	_, modifyVpcAttributeError = modifyVpcAttributeStub(VPCID, modifyVpcAttributeInputObject)

	if modifyVpcAttributeError != nil {
		return errors.New("|" + "HayMaker->haymakerengines->vpc_engine->InitializeVPCComponents->enableVPCDNS:" + modifyVpcAttributeError.Error() + "|")
	}

	return nil
}

func associateRouteTableStub(routeTableID *string, subnetID *string) (*ec2.AssociateRouteTableOutput, error) {

	associateRouteTableObject := &ec2.AssociateRouteTableInput{
		RouteTableId: aws.String(*routeTableID),
		SubnetId:     aws.String(*subnetID),
	}

	return vpcInstance.AssociateRouteTable(associateRouteTableObject)

}

func attachInternetGatewayStub(internetGatewayID *string, VPCID *string) (*ec2.AttachInternetGatewayOutput, error) {

	attachInternetGatewayInputObject := &ec2.AttachInternetGatewayInput{
		InternetGatewayId: aws.String(*internetGatewayID),
		VpcId:             aws.String(*VPCID),
	}

	return vpcInstance.AttachInternetGateway(attachInternetGatewayInputObject)

}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func createVPCtub(IPCIDR string, tags *[]*ec2.Tag) (*ec2.CreateVpcOutput, error) {
	createPrivateVpcInputObject := &ec2.CreateVpcInput{
		CidrBlock: aws.String(IPCIDR),
	}

	createVpcOutput, createVpcErr := vpcInstance.CreateVpc(createPrivateVpcInputObject)
	if createVpcErr != nil {
		return createVpcOutput, errors.New("|" + "HayMaker->haymakerengines->vpc_engine->InitializeVPCComponents->CreateVpc:" + createVpcErr.Error() + "|")
	}

	resources := []*string{createVpcOutput.Vpc.VpcId}

	_, createTagsErr := createTagsStub(resources, tags)
	if createTagsErr != nil {
		return createVpcOutput, errors.New("|" + "HayMaker->haymakerengines->vpc_engine->InitializeVPCComponents->CreateTagsStub:" + createTagsErr.Error() + "|")
	}

	return createVpcOutput, createVpcErr
}

func describeVPCsStub(filters []*ec2.Filter) (*ec2.DescribeVpcsOutput, error) {

	createPrivateVpcInputObject := &ec2.DescribeVpcsInput{
		Filters: filters,
	}

	return vpcInstance.DescribeVpcs(createPrivateVpcInputObject)

}

func deleteVPCStub(VPCID *string) (*ec2.DeleteVpcOutput, error) {

	deleteVpcInputObject := &ec2.DeleteVpcInput{
		VpcId: VPCID,
	}

	return vpcInstance.DeleteVpc(deleteVpcInputObject)

}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func describeNetworkInterfacesStub(networkInterfaceIds *[]*ec2.Filter) (*ec2.DescribeNetworkInterfacesOutput, error) {

	describeNetworkInterfacesInputObject := &ec2.DescribeNetworkInterfacesInput{

		Filters: *networkInterfaceIds,
	}

	return vpcInstance.DescribeNetworkInterfaces(describeNetworkInterfacesInputObject)

}

func deleteNetworkInterfaceStub(networkInterfaceId *string) (*ec2.DeleteNetworkInterfaceOutput, error) {

	deleteNetworkInterfaceInputObject := &ec2.DeleteNetworkInterfaceInput{
		NetworkInterfaceId: networkInterfaceId,
	}

	return vpcInstance.DeleteNetworkInterface(deleteNetworkInterfaceInputObject)

}

func detachNetworkInterfacesStub(attachmentId *string) (*ec2.DetachNetworkInterfaceOutput, error) {

	detachNetworkInterfaceInputObject := &ec2.DetachNetworkInterfaceInput{
		AttachmentId: attachmentId,
	}

	return vpcInstance.DetachNetworkInterface(detachNetworkInterfaceInputObject)

}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func createSubnetStub(VPCID *string, IPCIDR *string, availabilityZone *string, tags *[]*ec2.Tag) (*ec2.CreateSubnetOutput, error) {

	createSubnetInputObject := &ec2.CreateSubnetInput{
		CidrBlock:        aws.String(*IPCIDR),
		VpcId:            aws.String(*VPCID),
		AvailabilityZone: availabilityZone,
	}

	createSubnetResult, createSubnetError := vpcInstance.CreateSubnet(createSubnetInputObject)
	if createSubnetError != nil {
		return createSubnetResult, errors.New("|" + "HayMaker->haymakerengines->vpc_engine->InitializeVPCComponents->CreateSubnet:" + createSubnetError.Error() + "|")
	}

	resources := []*string{createSubnetResult.Subnet.SubnetId}

	_, createTagsErr := createTagsStub(resources, tags)
	if createTagsErr != nil {
		return createSubnetResult, errors.New("|" + "HayMaker->haymakerengines->vpc_engine->InitializeVPCComponents->CreateTagsStub:" + createTagsErr.Error() + "|")
	}

	return createSubnetResult, createSubnetError
}

func describeSubnetsStub(Filters []*ec2.Filter) (*ec2.DescribeSubnetsOutput, error) {

	describeSubnetsInputObject := &ec2.DescribeSubnetsInput{
		Filters: Filters,
	}

	return vpcInstance.DescribeSubnets(describeSubnetsInputObject)

}

func deleteSubnetsStub(subnetID *string) (*ec2.DeleteSubnetOutput, error) {

	deleteSubnetInputObject := &ec2.DeleteSubnetInput{
		SubnetId: subnetID,
	}

	return vpcInstance.DeleteSubnet(deleteSubnetInputObject)

}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func describeSecurityGroupsStub(filters []*ec2.Filter) (*ec2.DescribeSecurityGroupsOutput, error) {

	describeSecurityGroupsObject := &ec2.DescribeSecurityGroupsInput{
		Filters: filters,
	}

	return vpcInstance.DescribeSecurityGroups(describeSecurityGroupsObject)

}

func deleteSecurityGroupStub(groupId *string) (*ec2.DeleteSecurityGroupOutput, error) {

	describeSecurityGroupsObject := &ec2.DeleteSecurityGroupInput{
		GroupId: groupId,
	}

	return vpcInstance.DeleteSecurityGroup(describeSecurityGroupsObject)

}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

/*
1. Delete Route Table
2. Delete subnet
3. Delete gateway
4. Delete VPC
*/
func DestroyNetworkResources() error {

	fmt.Println("Destroying AWS Network Resources")

	var haymakerTagFilters []*ec2.Filter

	for key, value := range haymakerVPCTags {

		filterKey := ("tag:" + key)
		filterValue := value.(string)
		filterValues := []*string{&filterValue}

		vpcFilter := &ec2.Filter{
			Name:   aws.String(filterKey),
			Values: filterValues,
		}

		haymakerTagFilters = append(haymakerTagFilters, vpcFilter)
	}

	//We put this here because we need it to delete the load balancers

	describeVpcsResult, describeVpcsErr := describeVPCsStub(haymakerTagFilters)
	if describeVpcsErr != nil {
		return errors.New("|" + "HayMaker->haymakerengines->vpc_engine->PerformNetworkConfigCleanup->describeVPCsStub:" + describeVpcsErr.Error() + "|")
	}

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

	describeSubnetsResult, describeSubnetsErr := describeSubnetsStub(haymakerTagFilters)
	if describeSubnetsErr != nil {
		return errors.New("|" + "HayMaker->haymakerengines->vpc_engine->PerformNetworkConfigCleanup->describeSubnetsStub:" + describeSubnetsErr.Error() + "|")
	}

	describeLoadBalancersResult, describeLoadBalancersErr := DescribeLoadBalancersStub()
	if describeLoadBalancersErr != nil {
		return errors.New("|" + "HayMaker->haymakerengines->vpc_engine->PerformNetworkConfigCleanup->DescribeLoadBalancersStub:" + describeLoadBalancersErr.Error() + "|")
	}

	var securityGroupFilters []*ec2.Filter = make([]*ec2.Filter, 0)

	for securityGroupTagKey, securityGroupTagValue := range securityGroupTags {

		filterKey := ("tag:" + securityGroupTagKey)
		filterValue := securityGroupTagValue.(string)
		filterValues := []*string{&filterValue}

		securityGroupFilter := &ec2.Filter{
			Name:   aws.String(filterKey),
			Values: filterValues,
		}

		securityGroupFilters = append(securityGroupFilters, securityGroupFilter)
	}

	describeSecurityGroupsResult, describeSecurityGroupsError := describeSecurityGroupsStub(securityGroupFilters)
	if describeSecurityGroupsError != nil {
		return errors.New("|" + "HayMaker->haymakerengines->vpc_engine->PerformNetworkConfigCleanup->describeSecurityGroupsStub:" + describeSecurityGroupsError.Error() + "|")
	}

	var interfaceFiltersUsingSubnetIds []*ec2.Filter = make([]*ec2.Filter, 0)
	var filterValuesForSubnetIds []*string = make([]*string, 0)

	for _, subnet := range describeSubnetsResult.Subnets {

		filterValue := *subnet.SubnetId
		filterValuesForSubnetIds = append(filterValuesForSubnetIds, &filterValue)

		for _, loadBalancer := range describeLoadBalancersResult.LoadBalancerDescriptions {
			for _, vpc := range describeVpcsResult.Vpcs {
				if *vpc.VpcId == *loadBalancer.VPCId {

					_, deleteLoadBalancerErr := DeleteLoadBalancerStub(loadBalancer.LoadBalancerName)
					if deleteLoadBalancerErr != nil {
						return errors.New("|" + "HayMaker->haymakerengines->vpc_engine->DestroyNetworkResources->describeNetworkInterfacesStub:" + deleteLoadBalancerErr.Error() + "|")
					}
				}
			}
		}
	}

	interfaceFilterUsingSubnetId := &ec2.Filter{
		Name:   aws.String("subnet-id"),
		Values: filterValuesForSubnetIds,
	}
	interfaceFiltersUsingSubnetIds = append(interfaceFiltersUsingSubnetIds, interfaceFilterUsingSubnetId)

	var describeNetworkInterfacesResult *ec2.DescribeNetworkInterfacesOutput
	var describeNetworkInterfacesError error
	//Otherwise, the filter will be empty and the SDK does not like null filters
	if len(describeSubnetsResult.Subnets) > 0 {
		describeNetworkInterfacesResult, describeNetworkInterfacesError = describeNetworkInterfacesStub(&interfaceFiltersUsingSubnetIds)
		if describeNetworkInterfacesError != nil {
			return errors.New("|" + "HayMaker->haymakerengines->vpc_engine->DestroyNetworkResources->describeNetworkInterfacesStub:" + describeNetworkInterfacesError.Error() + "|")
		}

		for _, netInterface := range describeNetworkInterfacesResult.NetworkInterfaces {
			//sometimes the interface is in use for a while....
			/*
				if netInterface.Attachment != nil {
					_, detachNetworkInterfacesError := detachNetworkInterfacesStub(netInterface.Attachment.AttachmentId)
					if detachNetworkInterfacesError != nil {
						return errors.New("|" + "HayMaker->haymakerengines->vpc_engine->DestroyNetworkResources->detachNetworkInterfacesStub:" + detachNetworkInterfacesError.Error() + "|")
					}
				}
			*/

			var interfaceFiltersUsingInterfaceIds []*ec2.Filter = make([]*ec2.Filter, 0)
			var filterValuesForInterfaceId []*string = []*string{netInterface.NetworkInterfaceId}

			interfaceFilterUsingInterfaceId := &ec2.Filter{
				Name:   aws.String("network-interface-id"),
				Values: filterValuesForInterfaceId,
			}
			interfaceFiltersUsingInterfaceIds = append(interfaceFiltersUsingInterfaceIds, interfaceFilterUsingInterfaceId)

			for true {
				describeNetworkInterfacesResult, describeNetworkInterfacesError = describeNetworkInterfacesStub(&interfaceFiltersUsingInterfaceIds)
				if describeNetworkInterfacesError == nil {
					if *describeNetworkInterfacesResult.NetworkInterfaces[0].Status == ec2.NetworkInterfaceStatusInUse ||
						*describeNetworkInterfacesResult.NetworkInterfaces[0].Status == ec2.NetworkInterfaceStatusAssociated ||
						*describeNetworkInterfacesResult.NetworkInterfaces[0].Status == ec2.NetworkInterfaceStatusDetaching ||
						*describeNetworkInterfacesResult.NetworkInterfaces[0].Status == ec2.NetworkInterfaceStatusAttaching {
						time.Sleep(sleepDelay)
						continue
					} else if *describeNetworkInterfacesResult.NetworkInterfaces[0].Status == ec2.NetworkInterfaceStatusAvailable {
						_, deleteNetworkInterfacesError := deleteNetworkInterfaceStub(netInterface.NetworkInterfaceId)
						if deleteNetworkInterfacesError != nil {
							return errors.New("|" + "HayMaker->haymakerengines->vpc_engine->DestroyNetworkResources->deleteNetworkInterfaceStub:" + deleteNetworkInterfacesError.Error() + "|")
						}
						break
					}
				} else {
					//I expect this to happen if the interface does not exist.
					fmt.Println(errors.New("|" + "HayMaker->haymakerengines->vpc_engine->DestroyNetworkResources->deleteNetworkInterfaceStub:" + describeNetworkInterfacesError.Error() + "|"))
					break
				}
			}

		}
	}

	for _, securityGroup := range describeSecurityGroupsResult.SecurityGroups {
		_, deleteSecurityGroupErr := deleteSecurityGroupStub(securityGroup.GroupId)
		if deleteSecurityGroupErr != nil {
			return errors.New("|" + "HayMaker->haymakerengines->vpc_engine->DestroyNetworkResources->deleteSecurityGroupStub:" + deleteSecurityGroupErr.Error() + "|")
		}
	}

	//there may be some delay once more so i retry..
	for _, subnet := range describeSubnetsResult.Subnets {
		_, describeSubnetsErr := deleteSubnetsStub(subnet.SubnetId)
		if describeSubnetsErr != nil {
			return errors.New("|" + "HayMaker->haymakerengines->vpc_engine->DestroyNetworkResources->deleteSubnetsStub:" + describeSubnetsErr.Error() + "|")
		}
	}

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

	describeRouteTableResult, describeRouteTableErr := describeRouteTableStub(haymakerTagFilters)
	if describeRouteTableErr != nil {
		return errors.New("|" + "HayMaker->haymakerengines->vpc_engine->DestroyNetworkResources->describeRouteTableStub:" + describeRouteTableErr.Error() + "|")
	}

	for _, routeTable := range describeRouteTableResult.RouteTables {
		_, deleteRouteTableErr := deleteRouteTableStub(routeTable.RouteTableId)
		if deleteRouteTableErr != nil {
			return errors.New("|" + "HayMaker->haymakerengines->vpc_engine->DestroyNetworkResources->deleteRouteTableStub:" + deleteRouteTableErr.Error() + "|")
		}

	}

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

	describeInternetGatewaysResult, describeInternetGatewaysErr := describeInternetGatewaysStub(haymakerTagFilters)
	if describeInternetGatewaysErr != nil {
		fmt.Println(errors.New("|" + "HayMaker->haymakerengines->vpc_engine->DestroyNetworkResources->describeInternetGatewaysStub:" + describeInternetGatewaysErr.Error() + "|"))
	}

	for _, vpc := range describeVpcsResult.Vpcs {
		for _, gateway := range describeInternetGatewaysResult.InternetGateways {
			_, detachInternetGatewayErr := detachInternetGatewayStub(vpc.VpcId, gateway.InternetGatewayId)
			if detachInternetGatewayErr != nil {
				return errors.New("|" + "HayMaker->haymakerengines->vpc_engine->DestroyNetworkResources->detachInternetGatewayStub:" + detachInternetGatewayErr.Error() + "|")
			}

		}
	}

	for _, gateway := range describeInternetGatewaysResult.InternetGateways {
		_, deleteInternetGatewayErr := deleteInternetGatewayStub(gateway.InternetGatewayId)
		if deleteInternetGatewayErr != nil {
			return errors.New("|" + "HayMaker->haymakerengines->vpc_engine->DestroyNetworkResources->deleteInternetGatewayStub:" + deleteInternetGatewayErr.Error() + "|")
		}
	}

	for _, vpc := range describeVpcsResult.Vpcs {
		_, deleteVPCErr := deleteVPCStub(vpc.VpcId)
		if deleteVPCErr != nil {
			return errors.New("|" + "HayMaker->haymakerengines->vpc_engine->DestroyNetworkResources->deleteVPCStub:" + deleteVPCErr.Error() + "|")
		}

	}

	return nil
}

func preparePrivateSubnets() error {

	fmt.Println("Creating Private Subnets")

	for _, privateNetworkConfigElement := range privateNetworkConfigStruct {
		cidr := privateNetworkConfigElement.(map[string]interface{})["cidr"].(string)
		region := privateNetworkConfigElement.(map[string]interface{})["region"].(string)
		subnetTagsDict := privateNetworkConfigElement.(map[string]interface{})["subnet_tags"].(map[string]interface{})

		subnetTags := []*ec2.Tag{}

		for key, value := range subnetTagsDict {
			subnetTag := &ec2.Tag{
				Key:   aws.String(key),
				Value: aws.String(value.(string)),
			}

			subnetTags = append(subnetTags, subnetTag)
		}

		for key, value := range haymakerVPCTags {
			subnetTag := &ec2.Tag{
				Key:   aws.String(key),
				Value: aws.String(value.(string)),
			}

			subnetTags = append(subnetTags, subnetTag)
		}

		createSubnetResult, createSubnetError := createSubnetStub(vpcid, &cidr, &region, &subnetTags)
		if createSubnetError != nil {
			return errors.New("|" + "HayMaker->haymakerengines->vpc_engine->preparePrivateSubnets->createSubnetStub:" + createSubnetError.Error() + "|")
		}
		subnetIds[cidr] = createSubnetResult.Subnet.SubnetId
		////////////

		routeTableTags := []*ec2.Tag{}

		for key, value := range haymakerVPCTags {
			routeTableTag := &ec2.Tag{
				Key:   aws.String(key),
				Value: aws.String(value.(string)),
			}

			routeTableTags = append(routeTableTags, routeTableTag)
		}

		createRouteTableResult, createRouteTableError := createRouteTableStub(vpcid, &routeTableTags)
		if createRouteTableError != nil {
			return errors.New("|" + "HayMaker->haymakerengines->vpc_engine->preparePrivateSubnets->createRouteTableStub:" + createRouteTableError.Error() + "|")
		}

		routeTableIDs[cidr] = createRouteTableResult.RouteTable.RouteTableId

		destinationCIDRForDefaultGatewayLocal := destinationCIDRForDefaultGateway
		_, createRouteError := createRouteStub(&destinationCIDRForDefaultGatewayLocal, internetGatewayID, routeTableIDs[cidr], &routeTableTags)
		if createRouteError != nil {
			return errors.New("|" + "HayMaker->haymakerengines->vpc_engine->preparePrivateSubnets->createRouteStub:" + createRouteError.Error() + "|")
		}

		_, associateRouteTableError := associateRouteTableStub(routeTableIDs[cidr], subnetIds[cidr])
		if associateRouteTableError != nil {
			return errors.New("|" + "HayMaker->haymakerengines->vpc_engine->preparePrivateSubnets->associateRouteTableStub:" + associateRouteTableError.Error() + "|")
		}
	}

	return nil

}

func preparePublicSubnets() error {

	fmt.Println("Creating Public Subnets")

	for _, publicNetworkConfigElement := range publicNetworkConfigStruct {
		cidr := publicNetworkConfigElement.(map[string]interface{})["cidr"].(string)
		region := publicNetworkConfigElement.(map[string]interface{})["region"].(string)
		subnetTagsDict := publicNetworkConfigElement.(map[string]interface{})["subnet_tags"].(map[string]interface{})

		subnetTags := []*ec2.Tag{}

		for key, value := range subnetTagsDict {
			subnetTag := &ec2.Tag{
				Key:   aws.String(key),
				Value: aws.String(value.(string)),
			}

			subnetTags = append(subnetTags, subnetTag)
		}

		for key, value := range haymakerVPCTags {
			subnetTag := &ec2.Tag{
				Key:   aws.String(key),
				Value: aws.String(value.(string)),
			}

			subnetTags = append(subnetTags, subnetTag)
		}

		createSubnetResult, createSubnetError := createSubnetStub(vpcid, &cidr, &region, &subnetTags)
		if createSubnetError != nil {
			return errors.New("|" + "HayMaker->haymakerengines->vpc_engine->preparePublicSubnets->createSubnetStub:" + createSubnetError.Error() + "|")
		}
		subnetIds[cidr] = createSubnetResult.Subnet.SubnetId

		routeTableTags := []*ec2.Tag{}

		for key, value := range haymakerVPCTags {
			routeTableTag := &ec2.Tag{
				Key:   aws.String(key),
				Value: aws.String(value.(string)),
			}

			routeTableTags = append(routeTableTags, routeTableTag)
		}

		createRouteTableResult, createRouteTableError := createRouteTableStub(vpcid, &routeTableTags)
		if createRouteTableError != nil {
			return errors.New("|" + "HayMaker->haymakerengines->vpc_engine->preparePublicSubnets->createRouteTableStub:" + createRouteTableError.Error() + "|")
		}

		routeTableIDs[cidr] = createRouteTableResult.RouteTable.RouteTableId

		destinationCIDRForDefaultGatewayLocal := destinationCIDRForDefaultGateway
		_, createRouteError := createRouteStub(&destinationCIDRForDefaultGatewayLocal, internetGatewayID, routeTableIDs[cidr], &routeTableTags)
		if createRouteError != nil {
			return errors.New("|" + "HayMaker->haymakerengines->vpc_engine->preparePublicSubnets->createRouteStub:" + createRouteError.Error() + "|")
		}

		_, associateRouteTableError := associateRouteTableStub(routeTableIDs[cidr], subnetIds[cidr])
		if associateRouteTableError != nil {
			return errors.New("|" + "HayMaker->haymakerengines->vpc_engine->preparePublicSubnets->associateRouteTableStub:" + associateRouteTableError.Error() + "|")
		}
	}

	return nil

}

func CreateNetworkResources(deleteExisting bool) error {

	fmt.Println("Creating AWS Network Resources")

	var vpcTags []*ec2.Tag

	for key, value := range haymakerVPCTags {

		vpcTag := &ec2.Tag{
			Key:   aws.String(key),
			Value: aws.String(value.(string)),
		}

		vpcTags = append(vpcTags, vpcTag)
	}

	fmt.Println("Creating VPCs")
	createVPCResult, createVPCError := createVPCtub(*vpcCIDR, &vpcTags)
	if createVPCError != nil {
		return errors.New("|" + "HayMaker->haymakerengines->vpc_engine->CreateNetworkResources->createVPCtub:" + createVPCError.Error() + "|")
	}

	vpcid = createVPCResult.Vpc.VpcId

	enableVPCDNSError := enableVPCDNS(vpcid)
	if enableVPCDNSError != nil {
		return errors.New("|" + "HayMaker->haymakerengines->vpc_engine->CreateNetworkResources->enableVPCDNS:" + enableVPCDNSError.Error() + "|")
	}

	internetGatewayTags := []*ec2.Tag{}

	for key, value := range haymakerVPCTags {
		internetGatewayTag := &ec2.Tag{
			Key:   aws.String(key),
			Value: aws.String(value.(string)),
		}

		internetGatewayTags = append(internetGatewayTags, internetGatewayTag)
	}

	fmt.Println("Creating Gateway")
	createInternetGatewayResult, createInternetGatewayError := createInternetGatewayStub(&internetGatewayTags)
	if createInternetGatewayError != nil {
		return errors.New("|" + "HayMaker->haymakerengines->vpc_engine->CreateNetworkResources->createInternetGatewayStub:" + createInternetGatewayError.Error() + "|")
	}

	internetGatewayID = createInternetGatewayResult.InternetGateway.InternetGatewayId

	fmt.Println("Attaching Internet Gateway")
	_, attachInternetGatewayError := attachInternetGatewayStub(internetGatewayID, vpcid)
	if attachInternetGatewayError != nil {
		return errors.New("|" + "HayMaker->haymakerengines->vpc_engine->CreateNetworkResources->attachInternetGatewayStub:" + attachInternetGatewayError.Error() + "|")
	}

	preparePrivateSubnetsError := preparePrivateSubnets()

	if preparePrivateSubnetsError != nil {
		return errors.New("|" + "HayMaker->haymakerengines->vpc_engine->CreateNetworkResources->preparePrivateSubnets:" + preparePrivateSubnetsError.Error() + "|")
	}

	preparePublicSubnetsError := preparePublicSubnets()

	if preparePublicSubnetsError != nil {
		return errors.New("|" + "HayMaker->haymakerengines->vpc_engine->CreateNetworkResources->preparePublicSubnets:" + preparePublicSubnetsError.Error() + "|")
	}

	return nil
}

func InitVPCEngine(vInstance *ec2.EC2, vpcConfig interface{}) {
	vpcInstance = vInstance

	VPCCIDRTemp := vpcConfig.(map[string]interface{})["vpc_cidr"].(string)
	vpcCIDR = &VPCCIDRTemp

	securityGroupTags = vpcConfig.(map[string]interface{})["security_groups_tags"].(map[string]interface{})

	haymakerVPCTags = vpcConfig.(map[string]interface{})["haymaker_vpc_tags"].(map[string]interface{})

	privateNetworkConfigStruct = vpcConfig.(map[string]interface{})["private_network"].([]interface{})

	publicNetworkConfigStruct = vpcConfig.(map[string]interface{})["public_network"].([]interface{})

}

func GetSubnetIdsForEKSWorkerNodes() []*string {

	subnetIDsOut := make([]*string, 0)

	for _, privateNetworkConfigElement := range privateNetworkConfigStruct {
		cidr := privateNetworkConfigElement.(map[string]interface{})["cidr"].(string)
		subnetID := subnetIds[cidr]
		subnetIDsOut = append(subnetIDsOut, subnetID)
	}

	return subnetIDsOut
	/*
		subnetIDsOut := make([]*string, 0)

		for _, subnetId := range subnetIds {
			subnetIDsOut = append(subnetIDsOut, subnetId)

		}
		return subnetIDsOut
	*/
}

/*
func GetSubnetCIDRsForPublicAccess() []*string {

	subnetCIDRsOut := make([]*string, 0)

	for _, privateNetworkConfigElement := range publicNetworkConfigStruct.(map[string]interface{}) {
		cidr := privateNetworkConfigElement.(map[string]interface{})["cidr"].(string)
		subnetID := subnetIds[cidr]
		subnetIDsOut = append(subnetIDsOut, subnetID)
	}

	//	for ipCIDR, _ := range publicIPCIDRToAZs {
	//		ipCIDRTemp := ipCIDR
	//		subnetCIDRsOut = append(subnetCIDRsOut, &ipCIDRTemp)

	return subnetCIDRsOut
}*/
