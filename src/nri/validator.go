package nri

import (
	"github.com/newrelic/infra-integrations-sdk/log"
	"regexp"
	"strings"
)

//Validator groups the rules to validate the service name
type Validator struct {
	allowList  string
	denyList   string
	allowRegex string
}

//ValidateServiceName validates the serviceName against allowList, denyList, allowRegex
func (v Validator) ValidateServiceName(serviceName string) bool {
	for _, as := range strings.Split(v.denyList, ",") {
		if as == serviceName {
			return false
		}
	}
	for _, as := range strings.Split(v.allowList, ",") {
		if as == serviceName {
			return true
		}
	}
	if v.allowRegex == "" {
		return false
	}

	valid, err := regexp.MatchString(v.allowRegex, serviceName)
	if err != nil {
		log.Warn(err.Error())
		return false
	}

	return valid
}

//NewValidator create a new Validator instance
func NewValidator(allowList string, denyList string, allowRegex string) Validator {
	return Validator{allowList: allowList, denyList: denyList, allowRegex: allowRegex}
}
