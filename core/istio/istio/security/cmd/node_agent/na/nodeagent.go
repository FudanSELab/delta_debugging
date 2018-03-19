// Copyright 2017 Istio Authors
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

package na

import (
	"fmt"
	"time"

	"istio.io/istio/pkg/log"
	"istio.io/istio/security/pkg/caclient/grpc"
	pkiutil "istio.io/istio/security/pkg/pki/util"
	"istio.io/istio/security/pkg/platform"
	"istio.io/istio/security/pkg/util"
	"istio.io/istio/security/pkg/workload"
	pb "istio.io/istio/security/proto"
)

// The real node agent implementation. This implements the "Start" function
// in the NodeAgent interface.
type nodeAgentInternal struct {
	// Configuration specific to Node Agent
	config       *Config
	pc           platform.Client
	cAClient     grpc.CAGrpcClient
	identity     string
	secretServer workload.SecretServer
	certUtil     util.CertUtil
}

// Start starts the node Agent.
func (na *nodeAgentInternal) Start() error {
	if na.config == nil {
		return fmt.Errorf("node Agent configuration is nil")
	}

	if !na.pc.IsProperPlatform() {
		return fmt.Errorf("node Agent is not running on the right platform")
	}

	log.Infof("Node Agent starts successfully.")

	retries := 0
	retrialInterval := na.config.CSRInitialRetrialInterval
	identity, err := na.pc.GetServiceIdentity()
	if err != nil {
		return err
	}
	na.identity = identity
	var success bool
	for {
		privateKey, req, reqErr := na.createRequest()
		if reqErr != nil {
			return reqErr
		}

		log.Infof("Sending CSR (retrial #%d) ...", retries)

		resp, err := na.cAClient.SendCSR(req, na.pc, na.config.IstioCAAddress)
		if err == nil && resp != nil && resp.IsApproved {
			waitTime, ttlErr := na.certUtil.GetWaitTime(resp.SignedCert, time.Now())
			if ttlErr != nil {
				log.Errorf("Error getting TTL from approved cert: %v", ttlErr)
				success = false
			} else {
				if writeErr := na.secretServer.SetServiceIdentityCert(
					append(resp.SignedCert, resp.CertChain...)); writeErr != nil {
					return writeErr
				}
				if writeErr := na.secretServer.SetServiceIdentityPrivateKey(privateKey); writeErr != nil {
					return writeErr
				}
				log.Infof("CSR is approved successfully. Will renew cert in %s", waitTime.String())
				retries = 0
				retrialInterval = na.config.CSRInitialRetrialInterval
				timer := time.NewTimer(waitTime)
				<-timer.C
				success = true
			}
		} else {
			success = false
		}

		if !success {
			if retries >= na.config.CSRMaxRetries {
				return fmt.Errorf(
					"node agent can't get the CSR approved from Istio CA after max number of retries (%d)", na.config.CSRMaxRetries)
			}
			if err != nil {
				log.Errorf("CSR signing failed: %v. Will retry in %s", err, retrialInterval.String())
			} else if resp == nil {
				log.Errorf("CSR signing failed: response empty. Will retry in %s", retrialInterval.String())
			} else if !resp.IsApproved {
				log.Errorf("CSR signing failed: request not approved. Will retry in %s", retrialInterval.String())
			} else {
				log.Errorf("Certificate parsing error. Will retry in %s", retrialInterval.String())
			}
			retries++
			timer := time.NewTimer(retrialInterval)
			// Exponentially increase the backoff time.
			retrialInterval = retrialInterval * 2
			<-timer.C
		}
	}
}

func (na *nodeAgentInternal) createRequest() ([]byte, *pb.CsrRequest, error) {
	csr, privKey, err := pkiutil.GenCSR(pkiutil.CertOptions{
		Host:       na.identity,
		Org:        na.config.ServiceIdentityOrg,
		RSAKeySize: na.config.RSAKeySize,
	})
	if err != nil {
		return nil, nil, err
	}

	cred, err := na.pc.GetAgentCredential()
	if err != nil {
		return nil, nil, fmt.Errorf("request creation fails on getting agent credential (%v)", err)
	}

	return privKey, &pb.CsrRequest{
		CsrPem:              csr,
		NodeAgentCredential: cred,
		CredentialType:      na.pc.GetCredentialType(),
		RequestedTtlMinutes: int32(na.config.WorkloadCertTTL.Minutes()),
		ForCA:               false,
	}, nil
}
