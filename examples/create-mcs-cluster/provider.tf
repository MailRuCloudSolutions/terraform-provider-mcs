provider "mcs" {
    username          = "user@mail.ru"
    password          = "password"
    project_id        = "project_id"
}

provider "openstack" {
    user_name        = "user@mail.ru"
    password         = "password"
    tenant_id        = "project_id"
    user_domain_name = "users"
    auth_url         = "https://infra.mail.ru/identity/v3/"
}
