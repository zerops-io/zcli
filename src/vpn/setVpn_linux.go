//go:build linux
// +build linux

package vpn

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path"
	"strconv"

	vpnproxy "github.com/zerops-io/zcli/src/proto/vpnproxy"

	"github.com/google/uuid"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/utils/cmdRunner"
)

func getNewVpnInterfaceName() (string, error) {
	for i := 0; i < 99; i++ {
		interfaceName := fmt.Sprintf("wg%d", i)
		_, err := net.InterfaceByName(interfaceName)
		if err == nil {
			continue
		}
		if err.Error() == "route ip+net: no such network interface" {
			return interfaceName, nil
		}
	}
	return "", errors.New(i18n.VpnStartInterfaceAssignFailed)
}

func (h *Handler) setVpn(selectedVpnAddress, privateKey string, mtu uint32, response *vpnproxy.StartVpnResponse) error {
	var err error

	interfaceName, err := getNewVpnInterfaceName()
	if err != nil {
		return err
	}

	_, err = cmdRunner.Run(exec.Command("ip", "link", "add", interfaceName, "type", "wireguard"))
	if err != nil {
		if !errors.Is(err, cmdRunner.IpAlreadySetErr) {
			return err
		}
	}

	_, err = cmdRunner.Run(exec.Command("ip", "link", "set", "mtu", strconv.Itoa(int(mtu)), "up", "dev", interfaceName))
	if err != nil {
		return err
	}

	{
		privateKeyName := uuid.New().String()
		tempPrivateKeyFile := path.Join(os.TempDir(), privateKeyName)
		err = ioutil.WriteFile(tempPrivateKeyFile, []byte(privateKey), 0755)
		if err != nil {
			return err
		}
		_, err = cmdRunner.Run(exec.Command("wg", "set", interfaceName, "private-key", tempPrivateKeyFile))
		if err != nil {
			return err
		}
		err = os.Remove(tempPrivateKeyFile)
		if err != nil {
			return err
		}
	}

	_, err = cmdRunner.Run(exec.Command("ip", "link", "set", interfaceName, "up"))
	if err != nil {
		return err
	}

	_, err = cmdRunner.Run(exec.Command("wg", "set", interfaceName, "listen-port", wireguardPort))
	if err != nil {
		return err
	}

	clientIp := vpnproxy.FromProtoIP(response.GetVpn().GetAssignedClientIp())
	vpnRange := vpnproxy.FromProtoIPRange(response.GetVpn().GetVpnIpRange())

	args := []string{
		"set", interfaceName,
		"peer", response.GetVpn().GetServerPublicKey(),
		"allowed-ips", vpnRange.String(),
		"endpoint", selectedVpnAddress,
		"persistent-keepalive", "25",
	}
	_, err = cmdRunner.Run(exec.Command("wg", args...))
	if err != nil {
		if !errors.Is(err, cmdRunner.IpAlreadySetErr) {
			panic(err)
		}
	}

	_, err = cmdRunner.Run(exec.Command("ip", "-6", "address", "add", clientIp.String(), "dev", interfaceName))
	if err != nil {
		return err
	}

	_, err = cmdRunner.Run(exec.Command("ip", "route", "add", vpnRange.String(), "dev", interfaceName))
	if err != nil {
		return err
	}

	return nil
}
