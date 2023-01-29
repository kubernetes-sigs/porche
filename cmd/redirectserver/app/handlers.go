/*
Copyright 2022 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package app

import (
	"net/http"
	"strings"

	"k8s.io/klog/v2"

	"k8s.io/registry.k8s.io/pkg/clientip"
	"k8s.io/registry.k8s.io/pkg/net/cidrs"
	"k8s.io/registry.k8s.io/pkg/net/cidrs/aws"

	"sigs.k8s.io/porche/cmd/redirectserver/pkg/blobcache"
)

type MirrorConfig struct {
	// CanonicalFallback is the fallback URL direct to a public bucket or similar,
	// this is used when we're having problems finding a redirect bucket.
	CanonicalFallback string

	InfoURL    string
	PrivacyURL string
}

type Server struct {
	MirrorConfig

	// regionMapper maps from a client IP to a cloud & region
	regionMapper cidrs.IPMapper[string]

	// mirrorCache records whether the mirrors have the various files.
	// We use a lookup so we don't require 100% perfect replication to all mirrors.
	mirrorCache blobcache.BlobChecker
}

func NewServer(cfg MirrorConfig, mirrorCache blobcache.BlobChecker) *Server {
	s := &Server{
		MirrorConfig: cfg,
	}
	// initialize map of clientIP to AWS region
	s.regionMapper = aws.NewAWSRegionMapper()
	// cache of whether the mirrors have blobs
	s.mirrorCache = mirrorCache

	return s
}

// ServeHTTP is the main entry point
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// only allow GET, HEAD; we don't allow mutation!
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "Only GET and HEAD are allowed.", http.StatusMethodNotAllowed)
		return
	}
	path := r.URL.Path
	switch {
	case strings.HasPrefix(path, "/binaries/"):
		s.serveBinaries(w, r)
	case path == "/":
		http.Redirect(w, r, s.InfoURL, http.StatusTemporaryRedirect)
	case strings.HasPrefix(path, "/privacy"):
		http.Redirect(w, r, s.PrivacyURL, http.StatusTemporaryRedirect)
	default:
		klog.V(2).InfoS("unknown request", "path", path)
		http.NotFound(w, r)
	}
}

func (s *Server) serveBinaries(w http.ResponseWriter, r *http.Request) {
	rPath := r.URL.Path

	// for blob requests, check the client IP and determine the best backend
	clientIP, err := clientip.Get(r)
	if err != nil {
		// this should not happen
		klog.ErrorS(err, "failed to get client IP")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// check if client is known to be coming from an AWS region
	awsRegion, ipIsKnown := s.regionMapper.GetIP(clientIP)
	if !ipIsKnown {
		// no region match, redirect to fallback location
		klog.V(2).InfoS("region not known; redirecting request to fallback location", "path", rPath)
		s.redirectToCanonical(w, r)
		return
	}

	// check if blob is available in our S3 bucket for the region
	mirrorBase, found := awsRegionToS3URL(awsRegion)
	if !found {
		// fall back to redirect to upstream
		klog.InfoS("mirror not found for region; redirecting request to fallback location", "region", awsRegion)
		s.redirectToCanonical(w, r)
		return
	}

	mirrorURL := urlJoin(mirrorBase, rPath)
	if s.mirrorCache.BlobExists(mirrorURL, mirrorBase, rPath) {
		// blob known to be available in S3, redirect client there
		klog.V(2).InfoS("redirecting request to mirror", "path", rPath, "mirror", mirrorBase)
		http.Redirect(w, r, mirrorURL, http.StatusTemporaryRedirect)
		return
	}

	// fall back to redirect to upstream
	klog.V(2).InfoS("blob not found; redirecting request to fallback location", "path", rPath)
	s.redirectToCanonical(w, r)
}

func (s *Server) redirectToCanonical(w http.ResponseWriter, r *http.Request) {
	rPath := r.URL.Path

	redirectURL := urlJoin(s.CanonicalFallback, rPath)
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

// urlJoin performs simple url joining without canonicalization
func urlJoin(base string, path string) string {
	var s strings.Builder
	s.WriteString(base)
	if !strings.HasSuffix(base, "/") {
		s.WriteString("/")
	}
	s.WriteString(strings.TrimPrefix(path, "/"))
	return s.String()
}
