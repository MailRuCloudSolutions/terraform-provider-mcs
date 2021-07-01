package mcs

import (
	"fmt"
)

func baseURL(c ContainerClient, api string) string {
	return c.ServiceURL(api)
}

func getURL(c ContainerClient, api string, id string) string {
	return c.ServiceURL(api, id)
}

func deleteURL(c ContainerClient, api string, id string) string {
	return c.ServiceURL(api, id)
}

func kubeConfigURL(c ContainerClient, api string, id string) string {
	return c.ServiceURL(api, id, "kube_config")
}

func actionsURL(c ContainerClient, api string, id string) string {
	return c.ServiceURL(api, id, "actions")
}

func upgradeURL(c ContainerClient, api string, id string) string {
	return c.ServiceURL(api, id, "actions", "upgrade")
}

func scaleURL(c ContainerClient, api string, id string) string {
	return c.ServiceURL(api, id, "actions", "scale")
}

func queryListURL(c ContainerClient, api string, queryParams map[string]string) string {
	queryStr := ""
	for key, value := range queryParams {
		queryStr += fmt.Sprintf("%s=%s&", key, value)
	}
	queryStr = queryStr[:len(queryStr)-1]
	return c.ServiceURL(api + "/?" + queryStr)
}
