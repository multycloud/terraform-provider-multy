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
  name       = "example_vn"
  cidr_block = "10.0.0.0/16"
  cloud      = each.key
  location   = var.location
}
resource "multy_subnet" "subnet" {
  for_each           = var.clouds
  name               = "subnet"
  cidr_block         = "10.0.2.0/24"
  virtual_network_id = multy_virtual_network.example_vn[each.key].id
  availability_zone  = 2
}
resource "multy_network_interface" "public-nic" {
  for_each  = var.clouds
  name      = "test-public-nic"
  subnet_id = multy_subnet.subnet[each.key].id
  cloud     = each.key
  location  = var.location
}
resource "multy_network_interface" "private-nic" {
  for_each  = var.clouds
  name      = "test-private-nic"
  subnet_id = multy_subnet.subnet[each.key].id
  cloud     = each.key
  location  = var.location
}
resource "multy_public_ip" "ip" {
  for_each             = var.clouds
  name                 = "test-ip"
  network_interface_id = multy_network_interface.public-nic[each.key].id
  cloud                = each.key
  location             = var.location
}