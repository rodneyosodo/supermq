// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package certs

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"os"

	"github.com/mainflux/mainflux/pkg/errors"
)

// ConfigsPage contains page related metadata as well as list
type Page struct {
	Total  uint64
	Offset uint64
	Limit  uint64
	Certs  []Cert
}

// Repository specifies a Config persistence API.
type Repository interface {
	// Save  saves cert for thing into database
	Save(ctx context.Context, cert Cert) (string, error)

	// RetrieveAll retrieve issued certificates for given owner ID
	RetrieveAll(ctx context.Context, ownerID string, offset, limit uint64) (Page, error)

	// Remove removes certificate from DB for a given thing ID
	Remove(ctx context.Context, ownerID, thingID string) error

	// RetrieveByThing retrieves issued certificates for a given thing ID
	RetrieveByThing(ctx context.Context, ownerID, thingID string, offset, limit uint64) (Page, error)

	// RetrieveBySerial retrieves a certificate for a given serial ID
	RetrieveBySerial(ctx context.Context, ownerID, serialID string) (Cert, error)
}

func LoadCertificates(caPath, caKeyPath string) (tls.Certificate, *x509.Certificate, error) {
	var tlsCert tls.Certificate
	var caCert *x509.Certificate

	if caPath == "" || caKeyPath == "" {
		return tlsCert, caCert, nil
	}

	if _, err := os.Stat(caPath); os.IsNotExist(err) {
		return tlsCert, caCert, err
	}

	if _, err := os.Stat(caKeyPath); os.IsNotExist(err) {
		return tlsCert, caCert, err
	}

	tlsCert, err := tls.LoadX509KeyPair(caPath, caKeyPath)
	if err != nil {
		return tlsCert, caCert, errors.Wrap(err, err)
	}

	b, err := ioutil.ReadFile(caPath)
	if err != nil {
		return tlsCert, caCert, err
	}

	caCert, err = ReadCert(b)
	if err != nil {
		return tlsCert, caCert, errors.Wrap(err, err)
	}

	return tlsCert, caCert, nil
}

func ReadCert(b []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(b)
	if block == nil {
		return nil, errors.New("failed to decode PEM data")
	}

	return x509.ParseCertificate(block.Bytes)
}
