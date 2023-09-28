// Copyright (c) 2023 Michael D Henderson. All rights reserved.

package jot_test

import (
	"errors"
	"github.com/mdhender/fh/internal/jot"
	"testing"
	"time"
)

func TestNewFactory(t *testing.T) {
	f := jot.NewFactory("", "", 5*time.Minute)
	if f == nil {
		t.Fatalf("NewFactory: expected non-nil: got nil\n")
	}
}

func TestFactory_AddSigner(t *testing.T) {
	f := jot.NewFactory("", "", 5*time.Minute)

	// add a new signer
	aSigner, err := jot.NewHS256Signer("a", []byte("secret"), 5*time.Minute)
	if err != nil {
		t.Fatalf("NewSigner: err: expected nil: got %v\n", err)
	} else if aSigner == nil {
		t.Fatalf("NewSigner: expected non-nil: got nil\n")
	}
	err = f.AddSigner(aSigner)
	if err != nil {
		t.Fatalf("AddSigner: err: expected nil: got %v\n", err)
	} else if _, ok := f.Lookup(aSigner.Id(), aSigner.Algorithm()); !ok {
		t.Fatalf("AddSigner: Lookup: expected ok: got !ok\n")
	}

	// add a new signer
	bSigner, err := jot.NewHS256Signer("b", []byte("secret"), 5*time.Minute)
	if err != nil {
		t.Fatalf("NewSigner: err: expected nil: got %v\n", err)
	} else if bSigner == nil {
		t.Fatalf("NewSigner: expected non-nil: got nil\n")
	}
	err = f.AddSigner(bSigner)
	if err != nil {
		t.Fatalf("AddSigner: err: expected nil: got %v\n", err)
	} else if _, ok := f.Lookup(bSigner.Id(), bSigner.Algorithm()); !ok {
		t.Fatalf("AddSigner: Lookup: expected ok: got !ok\n")
	}

	// add an expired signer
	expSigner, err := jot.NewHS256Signer("exp", []byte("secret"), -5*time.Minute)
	if err != nil {
		t.Fatalf("NewSigner: err: expected nil: got %v\n", err)
	} else if expSigner == nil {
		t.Fatalf("NewSigner: expected non-nil: got nil\n")
	} else if !expSigner.Expired() {
		t.Fatalf("Expired: expected true: got false\n")
	}
	err = f.AddSigner(expSigner)
	if !errors.Is(err, jot.ErrSignerExpired) {
		t.Fatalf("AddSigner: err: expected ErrSignerExpired: got %v\n", err)
	}
}

func TestFactory_DeleteSigner(t *testing.T) {
	f := jot.NewFactory("", "", 5*time.Minute)

	// add a new signer
	aSigner, _ := jot.NewHS256Signer("a", []byte("secret"), 5*time.Minute)
	_ = f.AddSigner(aSigner)

	// verify signer is present
	if _, ok := f.Lookup(aSigner.Id(), aSigner.Algorithm()); !ok {
		t.Fatalf("LookupSigner: Lookup: expected ok: got !ok\n")
	}

	// delete the signer
	f.DeleteSigner("a")

	// verify signer is deleted
	if _, ok := f.Lookup(aSigner.Id(), aSigner.Algorithm()); ok {
		t.Fatalf("LookupSigner: Lookup: expected ok: got !ok\n")
	}

	// add the signer again and verify that it is present
	_ = f.AddSigner(aSigner)
	if _, ok := f.Lookup(aSigner.Id(), aSigner.Algorithm()); !ok {
		t.Fatalf("LookupSigner: Lookup: expected ok: got !ok\n")
	}
}

func TestFactory_ClaimsFromToken(t *testing.T) {
	// create a factory with a single signer
	f := jot.NewFactory("", "", 5*time.Minute)
	if f == nil {
		t.Fatalf("NewFactory: expected non-nil: got nil\n")
	} else if aSigner, err := jot.NewHS256Signer("a", []byte("secret"), 5*time.Minute); err != nil {
		t.Fatalf("NewSigner: err: expected nil: got %v\n", err)
	} else if err = f.AddSigner(aSigner); err != nil {
		t.Fatalf("AddSigner: err: expected nil: got %v\n", err)
	}
	claims := jot.Claims{
		Subject: "joe",
		Roles:   make(map[string]bool),
	}
	claims.Roles["guest"] = true
	// create a token
	token, err := f.ClaimsToToken(5*time.Minute, claims)
	if err != nil {
		t.Fatalf("ClaimsToTokens: err: expected nil: got %v\n", err)
	}
	tClaims, err := f.ClaimsFromToken(token)
	if err != nil {
		t.Fatalf("ClaimsFromTokens: err: expected nil: got %v\n", err)
	} else if tClaims == nil {
		t.Fatalf("ClaimsFromTokens: claims: expected non-nil: got nil\n")
	}
	if claims.Subject != tClaims.Subject {
		t.Errorf("ClaimsFromTokens: subject: expected %q: got %q\n", claims.Subject, tClaims.Subject)
	}
	for role := range claims.Roles {
		if !tClaims.Roles[role] {
			t.Errorf("ClaimsFromTokens: claims.roles: %q not in tClaims.roles\n", role)
		}
	}
	for role := range tClaims.Roles {
		if !claims.Roles[role] {
			t.Errorf("ClaimsFromTokens: tClaims.roles: %q not in claims.roles\n", role)
		}
	}
	// create an expired token
	expToken, err := f.ClaimsToToken(-5*time.Minute, claims)
	if err != nil {
		t.Fatalf("ClaimsToTokens: err: expected nil: got %v\n", err)
	}
	tClaims, err = f.ClaimsFromToken(expToken)
	if !errors.Is(err, jot.ErrClaimsExpired) {
		t.Fatalf("ClaimsFromTokens: err: expected ErrClaimsExpired: got %v\n", err)
	}
}

func TestFactory_PurgeExpiredSigners(t *testing.T) {
	f := jot.NewFactory("", "", 5*time.Minute)
	if f == nil {
		t.Fatalf("NewFactory: expected non-nil: got nil\n")
	}
	// add a new signer
	aSigner, err := jot.NewHS256Signer("a", []byte("secret"), 5*time.Minute)
	if err != nil {
		t.Fatalf("NewSigner: err: expected nil: got %v\n", err)
	} else if aSigner == nil {
		t.Fatalf("NewSigner: expected non-nil: got nil\n")
	}
	err = f.AddSigner(aSigner)
	if err != nil {
		t.Fatalf("AddSigner: err: expected nil: got %v\n", err)
	} else if _, ok := f.Lookup(aSigner.Id(), aSigner.Algorithm()); !ok {
		t.Fatalf("AddSigner: Lookup: expected ok: got !ok\n")
	}
	bSigner, err := jot.NewHS256Signer("b", []byte("secret"), 5*time.Minute)
	if err != nil {
		t.Fatalf("NewSigner: err: expected nil: got %v\n", err)
	} else if bSigner == nil {
		t.Fatalf("NewSigner: expected non-nil: got nil\n")
	}
	err = f.AddSigner(bSigner)
	if err != nil {
		t.Fatalf("AddSigner: err: expected nil: got %v\n", err)
	} else if _, ok := f.Lookup(bSigner.Id(), bSigner.Algorithm()); !ok {
		t.Fatalf("AddSigner: Lookup: expected ok: got !ok\n")
	}
	// add an expired signer
	expSigner, err := jot.NewHS256Signer("exp", []byte("secret"), -5*time.Minute)
	if err != nil {
		t.Fatalf("NewSigner: err: expected nil: got %v\n", err)
	} else if expSigner == nil {
		t.Fatalf("NewSigner: expected non-nil: got nil\n")
	} else if !expSigner.Expired() {
		t.Fatalf("Expired: expected true: got false\n")
	}
	err = f.AddSigner(expSigner)
	if !errors.Is(err, jot.ErrSignerExpired) {
		t.Fatalf("AddSigner: err: expected ErrSignerExpired: got %v\n", err)
	}
	// purge and verify
	f.PurgeExpiredSigners()
	if _, ok := f.Lookup(aSigner.Id(), aSigner.Algorithm()); !ok {
		t.Errorf("PurgeSigner: Lookup: a: expected ok: got !ok\n")
	}
	if _, ok := f.Lookup(bSigner.Id(), aSigner.Algorithm()); !ok {
		t.Errorf("PurgeSigner: Lookup: b: expected ok: got !ok\n")
	}
	// force expire the aSigner
	aSigner.Expire()
	if !aSigner.Expired() {
		t.Fatalf("Expire: expected true: got false\n")
	}
	// purge and verify
	f.PurgeExpiredSigners()
	if _, ok := f.Lookup(aSigner.Id(), aSigner.Algorithm()); ok {
		t.Errorf("PurgeSigner: Lookup: a: expected !ok: got ok\n")
	}
	if _, ok := f.Lookup(bSigner.Id(), aSigner.Algorithm()); !ok {
		t.Errorf("PurgeSigner: Lookup: b: expected ok: got !ok\n")
	}
}
