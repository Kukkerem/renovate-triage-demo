// Command provider is a miniature stand-in for a Crossplane provider's
// manager binary — just enough plumbing to give every internal/ package a
// real call graph, so tools like govulncheck can tell which imported
// vulnerabilities are actually reachable versus merely present in go.sum.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/kukkerem/renovate-triage-demo/internal/auth"
	"github.com/kukkerem/renovate-triage-demo/internal/cfclient"
	"github.com/kukkerem/renovate-triage-demo/internal/config"
	"github.com/kukkerem/renovate-triage-demo/internal/format"
	"github.com/kukkerem/renovate-triage-demo/internal/htmlscan"
)

// marketplaceCatalogStub stands in for a fetched BTP service marketplace
// catalog page, since main() should not require network access to run.
const marketplaceCatalogStub = `<html><head><title>BTP Service Marketplace</title></head><body></body></html>`

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg := &config.ProviderConfig{SubaccountID: "demo-subaccount"}
	if err := config.Touch(cfg); err != nil {
		return err
	}
	fmt.Println(config.Describe(cfg))

	fmt.Println("service plan:", format.PlanTitle("standard"))

	secret := []byte("demo-signing-key-do-not-use-in-production")
	token, err := auth.IssueToken(secret, "provider@demo", []string{"btp.read"}, time.Hour)
	if err != nil {
		return fmt.Errorf("issue service token: %w", err)
	}
	claims, err := auth.VerifyToken(secret, token)
	if err != nil {
		return fmt.Errorf("verify service token: %w", err)
	}
	fmt.Printf("issued token for %s with scopes %v\n", claims.Subject, claims.Scope)

	title, err := htmlscan.ExtractTitle(strings.NewReader(marketplaceCatalogStub))
	if err != nil {
		return fmt.Errorf("extract marketplace catalog title: %w", err)
	}
	fmt.Println("marketplace catalog title:", title)

	// Only exercised when CF credentials are configured, so a plain `go run`
	// never dials out.
	if apiRoot := os.Getenv("CF_API_ROOT"); apiRoot != "" {
		lister, err := cfclient.NewSpaceLister(apiRoot, os.Getenv("CF_CLIENT_ID"), os.Getenv("CF_CLIENT_SECRET"))
		if err != nil {
			return fmt.Errorf("build cf space lister: %w", err)
		}
		spaces, err := lister(context.Background())
		if err != nil {
			return fmt.Errorf("list cf spaces: %w", err)
		}
		fmt.Println("cf spaces:", spaces)
	}
	return nil
}
