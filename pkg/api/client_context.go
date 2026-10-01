/*
Copyright 2026 Pextra Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package api

import "context"

// ClusterFederationIdHeader is the header that must be sent with any operation on a cluster, or any resource below the
// cluster level, when the target cluster is part of a cluster federation. Requests without this header are not routed
// to the correct federation member, and will result in errors.
const ClusterFederationIdHeader = "X-Pce-Cluster-Federation-Id"

// clusterFederationCtxKey is the context key used to store the cluster federation ID in a context.
type clusterFederationCtxKey struct{}

// WithClusterFederationId returns a new context with the given cluster federation ID. If the ID is empty, the original
// context is returned unchanged.
func WithClusterFederationId(ctx context.Context, id string) context.Context {
	if id == "" {
		return ctx
	}
	return context.WithValue(ctx, clusterFederationCtxKey{}, id)
}

// ClusterFederationIdFromContext retrieves the cluster federation ID from the given context. If no cluster federation
// ID is present, an empty string is returned.
func ClusterFederationIdFromContext(ctx context.Context) string {
	id, _ := ctx.Value(clusterFederationCtxKey{}).(string)
	return id
}
