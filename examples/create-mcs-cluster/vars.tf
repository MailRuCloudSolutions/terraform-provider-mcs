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