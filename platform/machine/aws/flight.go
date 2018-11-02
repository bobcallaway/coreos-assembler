// Copyright 2016 CoreOS, Inc.
// Copyright 2018 Red Hat
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package aws

import (
	ctplatform "github.com/coreos/container-linux-config-transpiler/config/platform"
	"github.com/coreos/pkg/capnslog"

	"github.com/coreos/mantle/platform"
	"github.com/coreos/mantle/platform/api/aws"
)

const (
	Platform platform.Name = "aws"
)

var (
	plog = capnslog.NewPackageLogger("github.com/coreos/mantle", "platform/machine/aws")
)

type flight struct {
	*platform.BaseFlight
	api *aws.API
}

// NewFlight creates an instance of a Flight suitable for spawning
// instances on Amazon Web Services' Elastic Compute platform.
//
// NewFlight will consume the environment variables $AWS_REGION,
// $AWS_ACCESS_KEY_ID, and $AWS_SECRET_ACCESS_KEY to determine the region to
// spawn instances in and the credentials to use to authenticate.
func NewFlight(opts *aws.Options) (platform.Flight, error) {
	api, err := aws.New(opts)
	if err != nil {
		return nil, err
	}

	bf, err := platform.NewBaseFlight(opts.Options, Platform, ctplatform.EC2)
	if err != nil {
		return nil, err
	}

	af := &flight{
		BaseFlight: bf,
		api:        api,
	}

	return af, nil
}

// NewCluster creates an instance of a Cluster suitable for spawning
// instances on Amazon Web Services' Elastic Compute platform.
func (af *flight) NewCluster(rconf *platform.RuntimeConfig) (platform.Cluster, error) {
	bc, err := platform.NewBaseCluster(af.BaseFlight, rconf)
	if err != nil {
		return nil, err
	}

	ac := &cluster{
		BaseCluster: bc,
		flight:      af,
	}

	if !rconf.NoSSHKeyInMetadata {
		keys, err := ac.Keys()
		if err != nil {
			return nil, err
		}

		if err := af.api.AddKey(bc.Name(), keys[0].String()); err != nil {
			return nil, err
		}
	}

	af.AddCluster(ac)

	return ac, nil
}
