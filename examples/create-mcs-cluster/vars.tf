variable "public-key-file" {
    type = string
    default = "~/.ssh/id_rsa.pub"
}

variable "k8s-flavor" {
    type = string
    default = "b7d20f15-82f1-4ed4-a12e-e60277fe955f" # Standard 2-4-50
}
