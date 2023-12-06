package formatter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

//nolint:stylecheck
const expected = `[31;1mRESULTS[0;22m:
[31;1mDeprecated APIs[0;22m:
[31;1mSomeKind[0;22m found in [90;1msomegroup[0;22m/[90;1mv3[0;22m
		-> [34;1mObject[0;22m: myobj [95;1mlocation:[0;22m /some/location


[31;1mDeleted APIs[0;22m:
	 [37;41;1mAPIs REMOVED FROM THE CURRENT VERSION AND SHOULD BE MIGRATED IMMEDIATELY!![0;0;22m
[31;1mSomeKind1[0;22m found in [90;1msomegroup2[0;22m/[90;1mv4[0;22m
		-> [34;1mObject[0;22m: myobj2 [95;1mlocation:[0;22m /some/location3



Kubepug validates the APIs using Kubernetes markers. To know what are the deprecated and deleted APIS it checks, please go to https://kubepug.xyz/status/
`

func TestStoutOutput(t *testing.T) {
	f := &stdout{}

	out, err := f.Output(mockResult)
	require.NoError(t, err)
	require.Equal(t, expected, string(out))
}
