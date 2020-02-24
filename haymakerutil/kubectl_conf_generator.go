package haymakerutil

import (
	"fmt"
	"strings"
)

func GenerateKubectlConfFileFromStruct(configParameters map[string]interface{}) *string {

	confString := `
apiVersion: v1
clusters:
- cluster:
    server: %s
    certificate-authority-data: %s
  name: kubernetes
contexts:
- context:
    cluster: kubernetes
    user: aws
  name: aws
current-context: aws
kind: Config
preferences: {}
users:
- name: aws
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1alpha1
      command: aws
      args:
        - "eks"
        - "get-token"
        - "--cluster-name"
        - %s      
`
	server := *(configParameters["server"].(*string))
	cert := *(configParameters["cert"].(*string))
	name := *(configParameters["name"].(*string))

	formattedString := strings.TrimLeft(fmt.Sprintf(confString, server, cert, name), "\n")
	return &formattedString

}
