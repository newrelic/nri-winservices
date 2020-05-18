package nri

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestValidateServiceNameAllowList(t *testing.T) {
	valid := NewValidator(",,,casa,test", "", "").ValidateServiceName("test")
	require.True(t, valid)
}

func TestValidateServiceNameNoRules(t *testing.T) {
	valid := NewValidator("", "", "").ValidateServiceName("test")
	require.False(t, valid)
}

func TestValidateServiceRegex(t *testing.T) {
	valid := NewValidator(",,,casa,test", ",,,deny,", "^win").ValidateServiceName("win")
	require.True(t, valid)
	valid = NewValidator(",,,casa,test", ",,,deny,", ".*").ValidateServiceName("win")
	require.True(t, valid)
	valid = NewValidator(",,,casa,test", ",,,deny,", "[a-z]").ValidateServiceName("win")
	require.True(t, valid)
	valid = NewValidator(",,,casa,test", ",,,deny,", "^difwin").ValidateServiceName("win")
	require.False(t, valid)
}

func TestValidateServiceNamePrecedenceDenyList(t *testing.T) {
	valid := NewValidator(",,,casa,test", ",,,test,", ".*").ValidateServiceName("test")
	require.False(t, valid)
}
