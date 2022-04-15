variable "location" {
  type    = string
  default = "ireland"
}

variable "clouds" {
  type    = set(string)
  default = ["aws", "azure"]
}

resource "multy_virtual_network" "example_vn" {
  for_each   = var.clouds
  name       = "nic_test"
  cidr_block = "10.0.0.0/16"
  cloud      = each.key
  location   = var.location
}
resource "multy_subnet" "subnet" {
  for_each        = var.clouds
  name            = "nic_test"
  cidr_block      = "10.0.2.0/24"
  virtual_network = multy_virtual_network.example_vn.id
}
resource "multy_network_interface" "private-nic" {
  for_each  = var.clouds
  name      = "test-private-nic"
  subnet_id = multy_subnet.subnet[each.key].id
  cloud     = each.key
  location  = var.location
}