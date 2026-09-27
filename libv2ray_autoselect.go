package libv2ray

import (
	"github.com/xtls/xray-core/proxy/autoselect"
)

// AutoSelectStatus returns every running autoselect outbound as JSON: the member new connections
// go to, why the selection last changed, and what is known about each member (state, delay,
// failures, open connections). "[]" when none is running.
func AutoSelectStatus() (status string) {
	defer recoverCall("AutoSelectStatus", nil, func() { status = "[]" })
	return autoselect.StatusJSON()
}

// AutoSelectNetworkChanged makes every group re-measure all of its members now. Call it when the
// device moves to another network: every number measured on the previous one is stale.
func AutoSelectNetworkChanged() {
	defer recoverCall("AutoSelectNetworkChanged", nil, nil)
	autoselect.KickAll()
}

// AutoSelectPin forces the new connections of group onto member; an empty member returns the group
// to automatic selection. It reports whether the group and the member exist.
func AutoSelectPin(group string, member string) (ok bool) {
	defer recoverCall("AutoSelectPin", nil, func() { ok = false })
	return autoselect.PinMember(group, member)
}
