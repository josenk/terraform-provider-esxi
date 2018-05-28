package main

import (
    "fmt"
    "log"
    "os"

    "github.com/hashicorp/terraform/helper/schema"
    "github.com/hashicorp/terraform/terraform"
)

func init() {
    // Terraform is already adding the timestamp for us
    log.SetFlags(log.Lshortfile)
    log.SetPrefix(fmt.Sprintf("pid-%d-", os.Getpid()))
}

func Provider() terraform.ResourceProvider {
    return &schema.Provider{
        ResourcesMap: map[string]*schema.Resource{
            "esxi_guest": resourceGUEST(),
        },
    }
}
