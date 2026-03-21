package es

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/elastic/go-elasticsearch/v9"
)

func NewESClient() (*elasticsearch.TypedClient, error) {
	cfg := GetConfig()

	certPEM, err := os.ReadFile("configs/elastic.pem")
	if err != nil {
		log.Fatalf("Error reading CA certificate: %s", err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(certPEM) {
		log.Fatalf("Failed to append CA certificate")
	}

	esClient, err := elasticsearch.NewTypedClient(elasticsearch.Config{
		Addresses: cfg.Addresses,
		APIKey:    cfg.APIKey,
		CACert:    certPEM, // указываем сертификат
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: certPool,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	// Проверка соединения
	res, err := esClient.Cluster.Health().Do(context.Background())
	if err != nil {
		return nil, fmt.Errorf("cluster health error: %w", err)
	}

	if res.Status.Name == "red" {
		return nil, fmt.Errorf("cluster status is RED")
	}

	return esClient, nil
}
