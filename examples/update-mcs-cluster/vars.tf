variable "k8s-template-test" {
    type = string
    default = "9e7a9856-4e63-4d38-b105-65850df5eb4e"
}

variable "k8s-template" {
    type = string
    default = "95663bae-6763-4a53-9424-831975285cc1"
}

variable "public-key-file" {
    type = string
    default = "~/.ssh/id_rsa.pub"
}

variable "k8s-flavor-test" {
    type = string
    default = "Basic-1-1-10"
}

variable "k8s-flavor" {
    type = string
    default = "Basic-1-2-20"
}

variable "k8s-network-id" {
    type = string
    default = "95663bae-6763-4a53-9424-831975285cc1"
}

variable "k8s-router-id" {
    type = string
    default = "95663bae-6763-4a53-9424-831975285cc1"
}

variable "k8s-subnet-id" {
    type = string
    default = "95663bae-6763-4a53-9424-831975285cc1"
}

variable "new-master-flavor" {
    type = string
    default = "d659fa16-c7fb-42cf-8a5e-9bcbe80a7538"
}