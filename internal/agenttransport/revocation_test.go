package agenttransport

import (
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/imfeelingtheagi/probectl/internal/crypto"
)

// The agent-transport server wires the registry deny-list into its mTLS
// verification (U-038): once an agent's cert is revoked via the registry, the
// handshake refuses it. Exercises the server's real TLS config end to end
// (without a live gRPC dial — the verify hook is the unit under test).
func TestServerMTLSRefusesRevokedAgentCert(t *testing.T) {
	dir := t.TempDir()
	ca, err := crypto.GenerateCA("transport-ca", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	srvCert, srvKey, err := ca.IssueServerCert("control", []string{"localhost"}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	certFile := writeFile(t, dir, "srv.crt", srvCert)
	keyFile := writeFile(t, dir, "srv.key", srvKey)
	caFile := writeFile(t, dir, "ca.crt", ca.CertPEM())

	rl := crypto.NewRevocationList()
	cfg, err := crypto.ServerMTLSConfigRevocable(certFile, keyFile, caFile, rl)
	if err != nil {
		t.Fatalf("server mtls: %v", err)
	}

	id := crypto.AgentSPIFFEID("t1", "agent-1")
	agentPEM, _, err := ca.IssueClientCert("agent-1", id, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	leaf := leafFromPEM(t, agentPEM)

	// A live (non-revoked) agent cert is accepted.
	if err := cfg.VerifyPeerCertificate([][]byte{leaf.Raw}, nil); err != nil {
		t.Fatalf("live agent cert refused: %v", err)
	}

	// The registry revokes the agent (Replace = the refresh path); the SAME
	// server config now refuses the cert at the handshake.
	rl.Replace([]string{leaf.SerialNumber.Text(16)}, nil)
	if err := cfg.VerifyPeerCertificate([][]byte{leaf.Raw}, nil); err == nil ||
		!strings.Contains(err.Error(), "REVOKED") {
		t.Fatalf("revoked agent cert not refused: %v", err)
	}
}

func writeFile(t *testing.T, dir, name string, content []byte) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, content, 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func leafFromPEM(t *testing.T, certPEM []byte) *x509.Certificate {
	t.Helper()
	block, _ := pem.Decode(certPEM)
	if block == nil {
		t.Fatal("no PEM block")
	}
	c, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("parse leaf: %v", err)
	}
	return c
}
