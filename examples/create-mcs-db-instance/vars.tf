variable "public-key-file" {
  type    = string
  default = "~/.ssh/id_rsa.pub"
}

variable "db-instance-flavor" {
  type    = string
  default = "Basic-1-2-20"
}

variable "external_network_id" {
  description = "ID of already existing external network (ext-net)"
  type        = string
}
