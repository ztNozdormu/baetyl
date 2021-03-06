package initz

import (
	"fmt"
	"github.com/baetyl/baetyl/config"
	mc "github.com/baetyl/baetyl/mock"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var (
	collectorBadCases = []struct {
		name         string
		fingerprints []config.Fingerprint
		err          error
	}{
		{
			name: "0: BootID Node Error",
			fingerprints: []config.Fingerprint{
				{
					Proof: config.ProofBootID,
				},
			},
			err: ErrProofValueNotFound,
		},
		{
			name: "1: SystemUUID Node Error",
			fingerprints: []config.Fingerprint{
				{
					Proof: config.ProofSystemUUID,
				},
			},
			err: ErrProofValueNotFound,
		},
		{
			name: "2: MachineID Node Error",
			fingerprints: []config.Fingerprint{
				{
					Proof: config.ProofMachineID,
				},
			},
			err: ErrProofValueNotFound,
		},
		{
			name: "3: SN File Error",
			fingerprints: []config.Fingerprint{
				{
					Proof: config.ProofSN,
					Value: "fv.txt",
				},
			},
		},
		{
			name: "4: Default",
			fingerprints: []config.Fingerprint{
				{
					Proof: config.Proof("Error"),
				},
			},
			err: ErrProofTypeNotSupported,
		},
		{
			name: "5: HostName Node Error",
			fingerprints: []config.Fingerprint{
				{
					Proof: config.ProofHostName,
				},
			},
			err: ErrProofValueNotFound,
		},
	}
)

func TestInitialize_Activate_Err_Collector(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	c := &config.Config{}
	c.Engine.Kind = "kubernetes"
	c.Engine.Kubernetes.InCluster = true
	c.Init.Cloud.Active.Interval = 5 * time.Second

	init, err := NewInit(c)
	assert.Error(t, err)

	ami := mc.NewMockAMI(mockCtl)
	ami.EXPECT().CollectNodeInfo().Return(nil, nil).AnyTimes()
	init = genInitialize(t, c, ami)

	init.Start()
	init.Close()

	for _, tt := range collectorBadCases {
		t.Run(tt.name, func(t *testing.T) {
			c.Init.ActivateConfig.Fingerprints = tt.fingerprints
			_, err := init.collect()
			if tt.fingerprints[0].Proof == config.ProofSN {
				assert.NotNil(t, err)
			} else {
				assert.Equal(t, tt.err, errors.Cause(err))
			}
		})
	}
}

func TestInitialize_Activate_Err_Ami(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	ami := mc.NewMockAMI(mockCtl)
	ami.EXPECT().CollectNodeInfo().Return(nil, fmt.Errorf("ami error")).AnyTimes()

	c := &config.Config{}
	c.Init.Cloud.Active.Interval = 5 * time.Second
	c.Init.ActivateConfig.Fingerprints = collectorBadCases[0].fingerprints
	init := genInitialize(t, c, ami)
	_, err := init.collect()
	assert.NotNil(t, err)
}
