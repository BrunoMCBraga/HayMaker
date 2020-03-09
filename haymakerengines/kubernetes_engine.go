package haymakerengines

import (
	"errors"
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var clientConfig *kubernetes.Clientset

var kubeconfig string
var deploymentName string
var replicasLabels map[string]interface{}
var podsLabels map[string]interface{}
var containerName string
var imageName string
var containerPortName string
var containerPort int32

var serviceName string
var serviceLabels map[string]interface{}
var servicePort int32

func LoadKubeConfig() {

	config, buildConfigFromFlagsError := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if buildConfigFromFlagsError != nil {
		panic(buildConfigFromFlagsError)
	}

	clientset, newForConfigError := kubernetes.NewForConfig(config)
	if newForConfigError != nil {
		panic(newForConfigError)
	}

	clientConfig = clientset

	/*
		listOptions := v1.ListOptions{}
		nodesList, nodesListError := clientset.CoreV1().Nodes().List(context.Background(), listOptions)
		if newForConfigError != nil {
			panic(nodesListError)
		}

		for _, node := range nodesList.Items {
			fmt.Println(node.GetName())
		}
	*/

}

func DescribeService() error {

	fmt.Println("Getting Information for Service")

	serviceListResult, serviceListError := clientConfig.CoreV1().Services(apiv1.NamespaceDefault).List(metav1.ListOptions{})

	if serviceListError != nil {
		return errors.New("|" + "Striker->strikerengines->kubernetes_engine->DescribeService->Services.List:" + serviceListError.Error() + "|")
	}

	for _, service := range serviceListResult.Items {
		fmt.Println("Service: " + service.Name)
		for _, loadBalancerIngress := range service.Status.LoadBalancer.Ingress {
			fmt.Println("Hostname:" + loadBalancerIngress.Hostname)
			fmt.Println("IP:" + loadBalancerIngress.IP)
		}

	}

	return nil

}

func SpinupContainers() error {

	fmt.Println("Spining up Containers")

	replicasLabelsTemp := make(map[string]string, 0)

	for k, v := range replicasLabels {
		replicasLabelsTemp[k] = v.(string)
	}

	podsLabelsTemp := make(map[string]string, 0)

	for k, v := range podsLabels {
		podsLabelsTemp[k] = v.(string)
	}

	deploymentsClient := clientConfig.AppsV1().Deployments(apiv1.NamespaceDefault)
	var replicas int32 = 2

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: deploymentName,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: podsLabelsTemp,
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: podsLabelsTemp,
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  containerName,
							Image: imageName,
							Ports: []apiv1.ContainerPort{
								{
									Name:          containerPortName,
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: containerPort,
								},
							},
						},
					},
				},
			},
		},
	}

	// Create Deployment
	fmt.Println("Creating deployment...")
	decploymentsCreateResult, decploymentsCreateError := deploymentsClient.Create(deployment)
	if decploymentsCreateError != nil {
		return errors.New("|" + "Striker->strikerengines->kubernetes_engine->SpinupContainers->Deployments.Create:" + decploymentsCreateError.Error() + "|")
	}
	fmt.Printf("Created deployment %q.\n", decploymentsCreateResult.GetObjectMeta().GetName())

	return nil
}

func DeleteContainers() error {

	fmt.Println("Deleting Containers")

	deploymentsClient := clientConfig.AppsV1().Deployments(apiv1.NamespaceDefault)

	// Create Deployment
	fmt.Println("Deleting deployment...")
	deploymentDeleteError := deploymentsClient.Delete(deploymentName, &metav1.DeleteOptions{})

	if deploymentDeleteError != nil {
		if !strings.Contains(deploymentDeleteError.Error(), "not found") {
			return errors.New("|" + "Striker->strikerengines->kubernetes_engine->DeleteContainers->Deployments.Delete:" + deploymentDeleteError.Error() + "|")
		}
	}

	return nil
}

func CreateService() error {

	fmt.Println("Creating Service")

	podsLabelsTemp := make(map[string]string, 0)

	for k, v := range podsLabels {
		podsLabelsTemp[k] = v.(string)
	}

	createResult, createError := clientConfig.CoreV1().Services(apiv1.NamespaceDefault).Create(&apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: apiv1.NamespaceDefault,
		},
		Spec: apiv1.ServiceSpec{
			Ports: []v1.ServicePort{{
				Protocol: "TCP",
				Port:     containerPort,
			}},
			Selector: podsLabelsTemp,
			Type:     apiv1.ServiceTypeLoadBalancer,
		},
	})

	if createError != nil {
		return errors.New("|" + "Striker->strikerengines->kubernetes_engine->ExposeService->Services.Create:" + createError.Error() + "|")
	}

	fmt.Println("Service: " + createResult.Name)
	for _, loadBalancerIngress := range createResult.Status.LoadBalancer.Ingress {
		fmt.Println("Hostname:" + loadBalancerIngress.Hostname)
		fmt.Println("IP:" + loadBalancerIngress.IP)
	}

	return nil

}

func DeleteService() error {

	fmt.Println("Deleting Service")

	deleteError := clientConfig.CoreV1().Services(apiv1.NamespaceDefault).Delete(serviceName, &metav1.DeleteOptions{})

	if deleteError != nil {
		if !strings.Contains(deleteError.Error(), "not found") {
			return errors.New("|" + "Striker->strikerengines->kubernetes_engine->DeleteService->Services.Create:" + deleteError.Error() + "|")
		}
	}

	return nil

}

func InitKubernetesEngine(kubernetesConfig interface{}) {

	kubeconfig = kubernetesConfig.(map[string]interface{})["kubeconfig"].(string)

	//deployment config
	deployment := kubernetesConfig.(map[string]interface{})["deployment"]

	deploymentName = deployment.(map[string]interface{})["deployment_name"].(string)

	replicasLabels = deployment.(map[string]interface{})["replicas_labels"].(map[string]interface{})

	podsLabels = deployment.(map[string]interface{})["pods_labels"].(map[string]interface{})

	containerName = deployment.(map[string]interface{})["container_name"].(string)

	imageName = deployment.(map[string]interface{})["image_name"].(string)

	containerPortName = deployment.(map[string]interface{})["container_port_name"].(string)

	containerPort = int32(deployment.(map[string]interface{})["container_port"].(float64))

	//////Service configuration
	service := kubernetesConfig.(map[string]interface{})["service"]
	serviceName = service.(map[string]interface{})["service_name"].(string)

	serviceLabelsTemp := service.(map[string]interface{})["service_labels"].(map[string]interface{})
	serviceLabels = serviceLabelsTemp

	servicePort = int32(service.(map[string]interface{})["service_port"].(float64))

}
