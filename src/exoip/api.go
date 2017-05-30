package exoip

import (
	"errors"
	"fmt"
	"os"

	"github.com/pyr/egoscale/src/egoscale"
)

func FetchMyNic(ego *egoscale.Client, mserver string) (string, error) {

	instanceID, err := FetchMetadata(mserver, "/latest/instance-id")
	if err != nil {
		return "", err
	}
	vmInfo, err := ego.GetVirtualMachine(instanceID)
	if err != nil {
		return "", err
	}
	if len(vmInfo.Nic) < 1 {
		return "", errors.New("cannot find virtual machine Nic ID")
	}
	return vmInfo.Nic[0].Id, nil
}

func (engine *Engine) FetchNicAndVm() {

	mserver, err := FindMetadataServer()
	AssertSuccess(err)
	instanceID, err := FetchMetadata(mserver, "/latest/instance-id")
	AssertSuccess(err)

	vmInfo, err := engine.Exo.GetVirtualMachine(instanceID)
	AssertSuccess(err)

	if len(vmInfo.Nic) < 1 {
		Logger.Crit("cannot find virtual machine Nic ID")
		fmt.Fprintln(os.Stderr, "cannot find virtual machine Nic ID")
		os.Exit(1)
	}
	engine.ExoVM = vmInfo.Id
	engine.NicId = vmInfo.Nic[0].Id
}

func (engine *Engine) ObtainNic(nic_id string) error {

	_, err := engine.Exo.AddIpToNic(nic_id, engine.ExoIP.String())
	if err != nil {
		Logger.Crit(fmt.Sprintf("could not add ip %s to nic %s: %s",
			engine.ExoIP.String(),
			nic_id,
			err))
		return err
	}
	Logger.Info(fmt.Sprintf("claimed ip %s on nic %s", engine.ExoIP.String(), nic_id))
	return nil
}

func (engine *Engine) ReleaseMyNic() error {
	vm, err := engine.Exo.GetVirtualMachine(engine.ExoVM)
	if err != nil {
		Logger.Crit(fmt.Sprintf("could not remove ip from nic: could get virtualmachine: %s",
			err))
		return err
	}
	nicAddressID := ""
	for _, sec_ip := range vm.Nic[0].Secondaryip {
		if sec_ip.IpAddress == engine.ExoIP.String() {
			nicAddressID = sec_ip.Id
			break
		}
	}
	if len(nicAddressID) == 0 {
		Logger.Warning("could not remove ip from nic: unknown association")
		return fmt.Errorf("could not remove ip from nic: unknown association")
	}

	_, err = engine.Exo.RemoveIpFromNic(nicAddressID)
	if err != nil {
		Logger.Crit(fmt.Sprintf("could not dissociate ip %s: %s",
			engine.ExoIP.String(), err))
		return err
	}
	Logger.Info(fmt.Sprintf("released ip %s", engine.ExoIP.String()))
	return nil
}

func (engine *Engine) ReleaseNic(nic_id string) {

	vms, err := engine.Exo.ListVirtualMachines()
	if err != nil {
		Logger.Crit(fmt.Sprintf("could not remove ip from nic: could not list virtualmachines: %s",
			err))
		return
	}

	nicAddressID := ""
	for _, vm := range vms {
		if vm.Nic[0].Id == nic_id {
			for _, sec_ip := range vm.Nic[0].Secondaryip {
				if sec_ip.IpAddress == engine.ExoIP.String() {
					nicAddressID = sec_ip.Id
					break
				}
			}
		}
	}

	if len(nicAddressID) == 0 {
		Logger.Warning("could not remove ip from nic: unknown association")
		return
	}

	_, err = engine.Exo.RemoveIpFromNic(nicAddressID)
	if err != nil {
		Logger.Crit(fmt.Sprintf("could not remove ip from nic %s (%s): %s",
			nic_id, nicAddressID, err))
	}
	Logger.Info(fmt.Sprintf("released ip %s from nic %s", engine.ExoIP.String(), nic_id))
}

func VMHasSecurityGroup(vm *egoscale.VirtualMachine, sgname string) bool {

	for _, sg := range vm.SecurityGroups {
		if sg.Name == sgname {
			return true
		}
	}
	return false
}

func GetSecurityGroupPeers(ego *egoscale.Client, sgname string) ([]string, error) {

	peers := make([]string, 0)
	vms, err := ego.ListVirtualMachines()
	if err != nil {
		return nil, err
	}

	for _, vm := range vms {
		if VMHasSecurityGroup(vm, sgname) {
			primary_ip := vm.Nic[0].Ipaddress
			peers = append(peers, fmt.Sprintf("%s:%d", primary_ip, DefaultPort))
		}
	}

	return peers, nil
}

func findPeerNic(ego *egoscale.Client, ip string) (string, error) {

	vms, err := ego.ListVirtualMachines()
	if err != nil {
		return "", err
	}

	for _, vm := range vms {

		if vm.Nic[0].Ipaddress == ip {
			return vm.Nic[0].Id, nil
		}
	}

	return "", fmt.Errorf("cannot find nic")
}
