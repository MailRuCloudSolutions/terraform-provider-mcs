package mcs

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
