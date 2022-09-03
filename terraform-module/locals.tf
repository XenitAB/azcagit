locals {
  eln = join("-", [var.environment, var.location_short, var.name])
}
