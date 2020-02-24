package haymakerengines

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/elb"
)

var elbInstance *elb.ELB

func DescribeLoadBalancersStub() (*elb.DescribeLoadBalancersOutput, error) {

	describeLoadBalancersInputObject := &elb.DescribeLoadBalancersInput{}
	return elbInstance.DescribeLoadBalancers(describeLoadBalancersInputObject)

}

func DeleteLoadBalancerStub(loadBalancerArn *string) (*elb.DeleteLoadBalancerOutput, error) {

	fmt.Println("Deleting Load Balancer")

	deleteLoadBalancersInputObject := &elb.DeleteLoadBalancerInput{
		LoadBalancerName: loadBalancerArn,
	}

	return elbInstance.DeleteLoadBalancer(deleteLoadBalancersInputObject)

}

func InitELBEngine(elbIns *elb.ELB) {
	elbInstance = elbIns

}
