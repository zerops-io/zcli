package startStopDelete

import (
	"fmt"
	"strings"

	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/i18n"
)

func askForConfirmation(parent constants.ParentCmd) bool {
	if parent == constants.Project {
		fmt.Print(i18n.ProjectDeleteConfirm)
	} else {
		fmt.Print(i18n.ServiceDeleteConfirm)
	}

	var response string

	_, err := fmt.Scan(&response)
	if err != nil {
		fmt.Println(err)
		return false
	}

	resp := strings.ToLower(response)
	if resp == "y" || resp == "yes" {
		return true
	} else if resp == "n" || resp == "no" {
		return false
	} else {
		return askForConfirmation(parent)
	}
}

func (c RunConfig) getConfirm() string {
	if !c.Confirm {
		// run confirm dialogue
		shouldDelete := askForConfirmation(c.ParentCmd)
		if !shouldDelete {
			return i18n.DeleteCanceledByUser
		}
	}
	return ""
}
