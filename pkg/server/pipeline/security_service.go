package pipeline

import (
	"fmt"
	"github.com/itskovichanton/core/pkg/core"
	"github.com/itskovichanton/goava/pkg/goava/errs"
	"github.com/itskovichanton/goava/pkg/goava/utils"
	"github.com/spf13/cast"
	"sort"
)

type ISecurityService interface {
	Check(a *core.CallParams, actionName string) error
}

type SecurityServiceImpl struct {
	ISecurityService

	Config                        *core.Config
	ServerSettingsProviderService IServerSettingsProviderService
}

const (
	ReasonIncorrectSignature = "INCORRECT_SIGNATURE"
	ReasonProfileDenied      = "PROFILE_IS_DENIED"
	ReasonProfileDeniedByIP  = "PROFILE_IS_DENIED_BY_IP"
)

func (c *SecurityServiceImpl) Check(a *core.CallParams, actionName string) error {

	if c.Config.GetBoolWithDefaultValue(false, "security", "validateSignature") {
		err := c.validateParamsSgn(a)
		if err != nil {
			return err
		}
	}
	if len(a.Caller.Type) == 0 {
		return nil
	}
	securitySettings := c.ServerSettingsProviderService.GetSecurity()
	if securitySettings == nil {
		return nil
	}
	profiles, actionExists := securitySettings.Actions[actionName]
	if actionExists {
		for _, p := range profiles {
			if a.Caller.Type == p {
				profile, profileExists := securitySettings.Profiles[p]
				if profileExists {
					if profile.Denied {
						return errs.NewBaseErrorWithReason(fmt.Sprintf("Profile %v is denied", profile), ReasonProfileDenied)
					}
				} else {
					if len(profile.Ips) == 0 {
						return nil
					}
					for _, ip := range profile.Ips {
						if a.Caller.IP == ip {
							return nil
						}
					}
					return errs.NewBaseErrorWithReason(fmt.Sprintf("IP %v is denied for profile %v", a.Caller.IP, profile), ReasonProfileDeniedByIP)

				}
			}
		}
	}

	return nil
}

func (c *SecurityServiceImpl) validateParamsSgn(a *core.CallParams) error {

	// валидирую подпись
	keys := make([]string, 0, len(a.Parameters))
	for k := range a.Parameters {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	arg := c.Config.App.Name + "_"
	for _, k := range keys {
		arg = arg + cast.ToString(a.Parameters[k][0])
	}

	if a.Context().Request().Header.Get("sgn") != c.calcSgn(arg) {
		se := errs.NewBaseErrorWithReason("access denied", ReasonIncorrectSignature)
		return se
	}

	return nil
}

func (c *SecurityServiceImpl) calcSgn(arg string) string {
	return utils.CalcSha256(c.Config.App.Name + utils.MD5(utils.CalcSha256(arg)))
}
