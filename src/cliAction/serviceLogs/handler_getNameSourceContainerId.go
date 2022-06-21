package serviceLogs

import (
	"fmt"
	"github.com/zerops-io/zcli/src/i18n"
	"strconv"
	"strings"
)

func (h *Handler) getNameSourceContainerId(config RunConfig) (serviceName, source string, containerId int, err error) {
	sn := config.ServiceName
	source = RUNTIME

	if !strings.Contains(sn, AT) {
		return sn, source, 0, nil
	}
	split := strings.Split(sn, AT)
	sn = split[0]
	suffix := split[1]
	if strings.Contains(suffix, AT) {
		return "", "", 0, fmt.Errorf("%s", i18n.LogServiceNameInvalid)
	}

	if suffix == "" {
		return sn, source, 0, nil
	}

	containerIndex, err := strconv.Atoi(suffix)
	if err == nil {
		return sn, source, containerIndex, nil
	}

	if suffix != BUILD {
		return "", "", 0, fmt.Errorf("%s", i18n.LogSuffixInvalid)
	}
	source = BUILD
	return sn, source, 0, nil
}
