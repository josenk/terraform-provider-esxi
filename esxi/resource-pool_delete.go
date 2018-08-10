package esxi

import (
	"fmt"
	"log"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceRESOURCEPOOLDelete(d *schema.ResourceData, m interface{}) error {
  c := m.(*Config)
	esxiSSHinfo := SshConnectionStruct{c.esxiHostName, c.esxiHostPort, c.esxiUserName, c.esxiPassword}
	log.Println("[resourceRESOURCEPOOLDelete]" )

  var remote_cmd, stdout string
	var err error

	pool_id := d.Id()

	remote_cmd  = fmt.Sprintf("vim-cmd hostsvc/rsrc/destroy %s", pool_id)
  stdout, err = runRemoteSshCommand(esxiSSHinfo, remote_cmd, "destroy resource pool")
  if err != nil {
		// todo more descriptive err message
    log.Printf("[resourcePoolDELETE] Failed destroy resource pool id: %s\n", stdout)
    return err
  }

  d.SetId("")
  return nil
}
