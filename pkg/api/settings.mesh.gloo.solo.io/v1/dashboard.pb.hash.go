// Code generated by protoc-gen-ext. DO NOT EDIT.
// source: github.com/solo-io/gloo-mesh/api/settings/v1/dashboard.proto

package v1

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash"
	"hash/fnv"

	"github.com/mitchellh/hashstructure"
	safe_hasher "github.com/solo-io/protoc-gen-ext/pkg/hasher"
)

// ensure the imports are used
var (
	_ = errors.New("")
	_ = fmt.Print
	_ = binary.LittleEndian
	_ = new(hash.Hash64)
	_ = fnv.New64
	_ = hashstructure.Hash
	_ = new(safe_hasher.SafeHasher)
)

// Hash function
func (m *DashboardSpec) Hash(hasher hash.Hash64) (uint64, error) {
	if m == nil {
		return 0, nil
	}
	if hasher == nil {
		hasher = fnv.New64()
	}
	var err error
	if _, err = hasher.Write([]byte("settings.mesh.gloo.solo.io.github.com/solo-io/gloo-mesh/pkg/api/settings.mesh.gloo.solo.io/v1.DashboardSpec")); err != nil {
		return 0, err
	}

	if h, ok := interface{}(m.GetAuth()).(safe_hasher.SafeHasher); ok {
		if _, err = hasher.Write([]byte("Auth")); err != nil {
			return 0, err
		}
		if _, err = h.Hash(hasher); err != nil {
			return 0, err
		}
	} else {
		if fieldValue, err := hashstructure.Hash(m.GetAuth(), nil); err != nil {
			return 0, err
		} else {
			if _, err = hasher.Write([]byte("Auth")); err != nil {
				return 0, err
			}
			if err := binary.Write(hasher, binary.LittleEndian, fieldValue); err != nil {
				return 0, err
			}
		}
	}

	return hasher.Sum64(), nil
}

// Hash function
func (m *SessionConfig) Hash(hasher hash.Hash64) (uint64, error) {
	if m == nil {
		return 0, nil
	}
	if hasher == nil {
		hasher = fnv.New64()
	}
	var err error
	if _, err = hasher.Write([]byte("settings.mesh.gloo.solo.io.github.com/solo-io/gloo-mesh/pkg/api/settings.mesh.gloo.solo.io/v1.SessionConfig")); err != nil {
		return 0, err
	}

	if h, ok := interface{}(m.GetCookieOptions()).(safe_hasher.SafeHasher); ok {
		if _, err = hasher.Write([]byte("CookieOptions")); err != nil {
			return 0, err
		}
		if _, err = h.Hash(hasher); err != nil {
			return 0, err
		}
	} else {
		if fieldValue, err := hashstructure.Hash(m.GetCookieOptions(), nil); err != nil {
			return 0, err
		} else {
			if _, err = hasher.Write([]byte("CookieOptions")); err != nil {
				return 0, err
			}
			if err := binary.Write(hasher, binary.LittleEndian, fieldValue); err != nil {
				return 0, err
			}
		}
	}

	switch m.Backend.(type) {

	case *SessionConfig_Cookie:

		if h, ok := interface{}(m.GetCookie()).(safe_hasher.SafeHasher); ok {
			if _, err = hasher.Write([]byte("Cookie")); err != nil {
				return 0, err
			}
			if _, err = h.Hash(hasher); err != nil {
				return 0, err
			}
		} else {
			if fieldValue, err := hashstructure.Hash(m.GetCookie(), nil); err != nil {
				return 0, err
			} else {
				if _, err = hasher.Write([]byte("Cookie")); err != nil {
					return 0, err
				}
				if err := binary.Write(hasher, binary.LittleEndian, fieldValue); err != nil {
					return 0, err
				}
			}
		}

	case *SessionConfig_Redis:

		if h, ok := interface{}(m.GetRedis()).(safe_hasher.SafeHasher); ok {
			if _, err = hasher.Write([]byte("Redis")); err != nil {
				return 0, err
			}
			if _, err = h.Hash(hasher); err != nil {
				return 0, err
			}
		} else {
			if fieldValue, err := hashstructure.Hash(m.GetRedis(), nil); err != nil {
				return 0, err
			} else {
				if _, err = hasher.Write([]byte("Redis")); err != nil {
					return 0, err
				}
				if err := binary.Write(hasher, binary.LittleEndian, fieldValue); err != nil {
					return 0, err
				}
			}
		}

	}

	return hasher.Sum64(), nil
}

// Hash function
func (m *OidcConfig) Hash(hasher hash.Hash64) (uint64, error) {
	if m == nil {
		return 0, nil
	}
	if hasher == nil {
		hasher = fnv.New64()
	}
	var err error
	if _, err = hasher.Write([]byte("settings.mesh.gloo.solo.io.github.com/solo-io/gloo-mesh/pkg/api/settings.mesh.gloo.solo.io/v1.OidcConfig")); err != nil {
		return 0, err
	}

	if _, err = hasher.Write([]byte(m.GetClientId())); err != nil {
		return 0, err
	}

	if h, ok := interface{}(m.GetClientSecret()).(safe_hasher.SafeHasher); ok {
		if _, err = hasher.Write([]byte("ClientSecret")); err != nil {
			return 0, err
		}
		if _, err = h.Hash(hasher); err != nil {
			return 0, err
		}
	} else {
		if fieldValue, err := hashstructure.Hash(m.GetClientSecret(), nil); err != nil {
			return 0, err
		} else {
			if _, err = hasher.Write([]byte("ClientSecret")); err != nil {
				return 0, err
			}
			if err := binary.Write(hasher, binary.LittleEndian, fieldValue); err != nil {
				return 0, err
			}
		}
	}

	if _, err = hasher.Write([]byte(m.GetIssuerUrl())); err != nil {
		return 0, err
	}

	{
		var result uint64
		innerHash := fnv.New64()
		for k, v := range m.GetAuthEndpointQueryParams() {
			innerHash.Reset()

			if _, err = innerHash.Write([]byte(v)); err != nil {
				return 0, err
			}

			if _, err = innerHash.Write([]byte(k)); err != nil {
				return 0, err
			}

			result = result ^ innerHash.Sum64()
		}
		err = binary.Write(hasher, binary.LittleEndian, result)
		if err != nil {
			return 0, err
		}

	}

	{
		var result uint64
		innerHash := fnv.New64()
		for k, v := range m.GetTokenEndpointQueryParams() {
			innerHash.Reset()

			if _, err = innerHash.Write([]byte(v)); err != nil {
				return 0, err
			}

			if _, err = innerHash.Write([]byte(k)); err != nil {
				return 0, err
			}

			result = result ^ innerHash.Sum64()
		}
		err = binary.Write(hasher, binary.LittleEndian, result)
		if err != nil {
			return 0, err
		}

	}

	if _, err = hasher.Write([]byte(m.GetAppUrl())); err != nil {
		return 0, err
	}

	if _, err = hasher.Write([]byte(m.GetCallbackPath())); err != nil {
		return 0, err
	}

	if _, err = hasher.Write([]byte(m.GetLogoutPath())); err != nil {
		return 0, err
	}

	for _, v := range m.GetScopes() {

		if _, err = hasher.Write([]byte(v)); err != nil {
			return 0, err
		}

	}

	if h, ok := interface{}(m.GetSession()).(safe_hasher.SafeHasher); ok {
		if _, err = hasher.Write([]byte("Session")); err != nil {
			return 0, err
		}
		if _, err = h.Hash(hasher); err != nil {
			return 0, err
		}
	} else {
		if fieldValue, err := hashstructure.Hash(m.GetSession(), nil); err != nil {
			return 0, err
		} else {
			if _, err = hasher.Write([]byte("Session")); err != nil {
				return 0, err
			}
			if err := binary.Write(hasher, binary.LittleEndian, fieldValue); err != nil {
				return 0, err
			}
		}
	}

	if h, ok := interface{}(m.GetHeader()).(safe_hasher.SafeHasher); ok {
		if _, err = hasher.Write([]byte("Header")); err != nil {
			return 0, err
		}
		if _, err = h.Hash(hasher); err != nil {
			return 0, err
		}
	} else {
		if fieldValue, err := hashstructure.Hash(m.GetHeader(), nil); err != nil {
			return 0, err
		} else {
			if _, err = hasher.Write([]byte("Header")); err != nil {
				return 0, err
			}
			if err := binary.Write(hasher, binary.LittleEndian, fieldValue); err != nil {
				return 0, err
			}
		}
	}

	if h, ok := interface{}(m.GetDiscoveryOverride()).(safe_hasher.SafeHasher); ok {
		if _, err = hasher.Write([]byte("DiscoveryOverride")); err != nil {
			return 0, err
		}
		if _, err = h.Hash(hasher); err != nil {
			return 0, err
		}
	} else {
		if fieldValue, err := hashstructure.Hash(m.GetDiscoveryOverride(), nil); err != nil {
			return 0, err
		} else {
			if _, err = hasher.Write([]byte("DiscoveryOverride")); err != nil {
				return 0, err
			}
			if err := binary.Write(hasher, binary.LittleEndian, fieldValue); err != nil {
				return 0, err
			}
		}
	}

	if h, ok := interface{}(m.GetDiscoveryPollInterval()).(safe_hasher.SafeHasher); ok {
		if _, err = hasher.Write([]byte("DiscoveryPollInterval")); err != nil {
			return 0, err
		}
		if _, err = h.Hash(hasher); err != nil {
			return 0, err
		}
	} else {
		if fieldValue, err := hashstructure.Hash(m.GetDiscoveryPollInterval(), nil); err != nil {
			return 0, err
		} else {
			if _, err = hasher.Write([]byte("DiscoveryPollInterval")); err != nil {
				return 0, err
			}
			if err := binary.Write(hasher, binary.LittleEndian, fieldValue); err != nil {
				return 0, err
			}
		}
	}

	if h, ok := interface{}(m.GetJwksCacheRefreshPolicy()).(safe_hasher.SafeHasher); ok {
		if _, err = hasher.Write([]byte("JwksCacheRefreshPolicy")); err != nil {
			return 0, err
		}
		if _, err = h.Hash(hasher); err != nil {
			return 0, err
		}
	} else {
		if fieldValue, err := hashstructure.Hash(m.GetJwksCacheRefreshPolicy(), nil); err != nil {
			return 0, err
		} else {
			if _, err = hasher.Write([]byte("JwksCacheRefreshPolicy")); err != nil {
				return 0, err
			}
			if err := binary.Write(hasher, binary.LittleEndian, fieldValue); err != nil {
				return 0, err
			}
		}
	}

	return hasher.Sum64(), nil
}

// Hash function
func (m *JwksOnDemandCacheRefreshPolicy) Hash(hasher hash.Hash64) (uint64, error) {
	if m == nil {
		return 0, nil
	}
	if hasher == nil {
		hasher = fnv.New64()
	}
	var err error
	if _, err = hasher.Write([]byte("settings.mesh.gloo.solo.io.github.com/solo-io/gloo-mesh/pkg/api/settings.mesh.gloo.solo.io/v1.JwksOnDemandCacheRefreshPolicy")); err != nil {
		return 0, err
	}

	switch m.Policy.(type) {

	case *JwksOnDemandCacheRefreshPolicy_Never:

		if h, ok := interface{}(m.GetNever()).(safe_hasher.SafeHasher); ok {
			if _, err = hasher.Write([]byte("Never")); err != nil {
				return 0, err
			}
			if _, err = h.Hash(hasher); err != nil {
				return 0, err
			}
		} else {
			if fieldValue, err := hashstructure.Hash(m.GetNever(), nil); err != nil {
				return 0, err
			} else {
				if _, err = hasher.Write([]byte("Never")); err != nil {
					return 0, err
				}
				if err := binary.Write(hasher, binary.LittleEndian, fieldValue); err != nil {
					return 0, err
				}
			}
		}

	case *JwksOnDemandCacheRefreshPolicy_Always:

		if h, ok := interface{}(m.GetAlways()).(safe_hasher.SafeHasher); ok {
			if _, err = hasher.Write([]byte("Always")); err != nil {
				return 0, err
			}
			if _, err = h.Hash(hasher); err != nil {
				return 0, err
			}
		} else {
			if fieldValue, err := hashstructure.Hash(m.GetAlways(), nil); err != nil {
				return 0, err
			} else {
				if _, err = hasher.Write([]byte("Always")); err != nil {
					return 0, err
				}
				if err := binary.Write(hasher, binary.LittleEndian, fieldValue); err != nil {
					return 0, err
				}
			}
		}

	case *JwksOnDemandCacheRefreshPolicy_MaxIdpReqPerPollingInterval:

		err = binary.Write(hasher, binary.LittleEndian, m.GetMaxIdpReqPerPollingInterval())
		if err != nil {
			return 0, err
		}

	}

	return hasher.Sum64(), nil
}

// Hash function
func (m *DashboardStatus) Hash(hasher hash.Hash64) (uint64, error) {
	if m == nil {
		return 0, nil
	}
	if hasher == nil {
		hasher = fnv.New64()
	}
	var err error
	if _, err = hasher.Write([]byte("settings.mesh.gloo.solo.io.github.com/solo-io/gloo-mesh/pkg/api/settings.mesh.gloo.solo.io/v1.DashboardStatus")); err != nil {
		return 0, err
	}

	err = binary.Write(hasher, binary.LittleEndian, m.GetObservedGeneration())
	if err != nil {
		return 0, err
	}

	err = binary.Write(hasher, binary.LittleEndian, m.GetState())
	if err != nil {
		return 0, err
	}

	for _, v := range m.GetErrors() {

		if _, err = hasher.Write([]byte(v)); err != nil {
			return 0, err
		}

	}

	return hasher.Sum64(), nil
}

// Hash function
func (m *DashboardSpec_AuthConfig) Hash(hasher hash.Hash64) (uint64, error) {
	if m == nil {
		return 0, nil
	}
	if hasher == nil {
		hasher = fnv.New64()
	}
	var err error
	if _, err = hasher.Write([]byte("settings.mesh.gloo.solo.io.github.com/solo-io/gloo-mesh/pkg/api/settings.mesh.gloo.solo.io/v1.DashboardSpec_AuthConfig")); err != nil {
		return 0, err
	}

	switch m.Backend.(type) {

	case *DashboardSpec_AuthConfig_Oidc:

		if h, ok := interface{}(m.GetOidc()).(safe_hasher.SafeHasher); ok {
			if _, err = hasher.Write([]byte("Oidc")); err != nil {
				return 0, err
			}
			if _, err = h.Hash(hasher); err != nil {
				return 0, err
			}
		} else {
			if fieldValue, err := hashstructure.Hash(m.GetOidc(), nil); err != nil {
				return 0, err
			} else {
				if _, err = hasher.Write([]byte("Oidc")); err != nil {
					return 0, err
				}
				if err := binary.Write(hasher, binary.LittleEndian, fieldValue); err != nil {
					return 0, err
				}
			}
		}

	}

	return hasher.Sum64(), nil
}

// Hash function
func (m *SessionConfig_CookieSession) Hash(hasher hash.Hash64) (uint64, error) {
	if m == nil {
		return 0, nil
	}
	if hasher == nil {
		hasher = fnv.New64()
	}
	var err error
	if _, err = hasher.Write([]byte("settings.mesh.gloo.solo.io.github.com/solo-io/gloo-mesh/pkg/api/settings.mesh.gloo.solo.io/v1.SessionConfig_CookieSession")); err != nil {
		return 0, err
	}

	return hasher.Sum64(), nil
}

// Hash function
func (m *SessionConfig_RedisSession) Hash(hasher hash.Hash64) (uint64, error) {
	if m == nil {
		return 0, nil
	}
	if hasher == nil {
		hasher = fnv.New64()
	}
	var err error
	if _, err = hasher.Write([]byte("settings.mesh.gloo.solo.io.github.com/solo-io/gloo-mesh/pkg/api/settings.mesh.gloo.solo.io/v1.SessionConfig_RedisSession")); err != nil {
		return 0, err
	}

	if _, err = hasher.Write([]byte(m.GetHost())); err != nil {
		return 0, err
	}

	err = binary.Write(hasher, binary.LittleEndian, m.GetDb())
	if err != nil {
		return 0, err
	}

	err = binary.Write(hasher, binary.LittleEndian, m.GetPoolSize())
	if err != nil {
		return 0, err
	}

	if _, err = hasher.Write([]byte(m.GetKeyPrefix())); err != nil {
		return 0, err
	}

	if _, err = hasher.Write([]byte(m.GetCookieName())); err != nil {
		return 0, err
	}

	if h, ok := interface{}(m.GetAllowRefreshing()).(safe_hasher.SafeHasher); ok {
		if _, err = hasher.Write([]byte("AllowRefreshing")); err != nil {
			return 0, err
		}
		if _, err = h.Hash(hasher); err != nil {
			return 0, err
		}
	} else {
		if fieldValue, err := hashstructure.Hash(m.GetAllowRefreshing(), nil); err != nil {
			return 0, err
		} else {
			if _, err = hasher.Write([]byte("AllowRefreshing")); err != nil {
				return 0, err
			}
			if err := binary.Write(hasher, binary.LittleEndian, fieldValue); err != nil {
				return 0, err
			}
		}
	}

	return hasher.Sum64(), nil
}

// Hash function
func (m *SessionConfig_CookieOptions) Hash(hasher hash.Hash64) (uint64, error) {
	if m == nil {
		return 0, nil
	}
	if hasher == nil {
		hasher = fnv.New64()
	}
	var err error
	if _, err = hasher.Write([]byte("settings.mesh.gloo.solo.io.github.com/solo-io/gloo-mesh/pkg/api/settings.mesh.gloo.solo.io/v1.SessionConfig_CookieOptions")); err != nil {
		return 0, err
	}

	if h, ok := interface{}(m.GetMaxAge()).(safe_hasher.SafeHasher); ok {
		if _, err = hasher.Write([]byte("MaxAge")); err != nil {
			return 0, err
		}
		if _, err = h.Hash(hasher); err != nil {
			return 0, err
		}
	} else {
		if fieldValue, err := hashstructure.Hash(m.GetMaxAge(), nil); err != nil {
			return 0, err
		} else {
			if _, err = hasher.Write([]byte("MaxAge")); err != nil {
				return 0, err
			}
			if err := binary.Write(hasher, binary.LittleEndian, fieldValue); err != nil {
				return 0, err
			}
		}
	}

	err = binary.Write(hasher, binary.LittleEndian, m.GetNotSecure())
	if err != nil {
		return 0, err
	}

	if h, ok := interface{}(m.GetPath()).(safe_hasher.SafeHasher); ok {
		if _, err = hasher.Write([]byte("Path")); err != nil {
			return 0, err
		}
		if _, err = h.Hash(hasher); err != nil {
			return 0, err
		}
	} else {
		if fieldValue, err := hashstructure.Hash(m.GetPath(), nil); err != nil {
			return 0, err
		} else {
			if _, err = hasher.Write([]byte("Path")); err != nil {
				return 0, err
			}
			if err := binary.Write(hasher, binary.LittleEndian, fieldValue); err != nil {
				return 0, err
			}
		}
	}

	if _, err = hasher.Write([]byte(m.GetDomain())); err != nil {
		return 0, err
	}

	return hasher.Sum64(), nil
}

// Hash function
func (m *OidcConfig_HeaderConfig) Hash(hasher hash.Hash64) (uint64, error) {
	if m == nil {
		return 0, nil
	}
	if hasher == nil {
		hasher = fnv.New64()
	}
	var err error
	if _, err = hasher.Write([]byte("settings.mesh.gloo.solo.io.github.com/solo-io/gloo-mesh/pkg/api/settings.mesh.gloo.solo.io/v1.OidcConfig_HeaderConfig")); err != nil {
		return 0, err
	}

	if _, err = hasher.Write([]byte(m.GetIdTokenHeader())); err != nil {
		return 0, err
	}

	if _, err = hasher.Write([]byte(m.GetAccessTokenHeader())); err != nil {
		return 0, err
	}

	return hasher.Sum64(), nil
}

// Hash function
func (m *OidcConfig_DiscoveryOverride) Hash(hasher hash.Hash64) (uint64, error) {
	if m == nil {
		return 0, nil
	}
	if hasher == nil {
		hasher = fnv.New64()
	}
	var err error
	if _, err = hasher.Write([]byte("settings.mesh.gloo.solo.io.github.com/solo-io/gloo-mesh/pkg/api/settings.mesh.gloo.solo.io/v1.OidcConfig_DiscoveryOverride")); err != nil {
		return 0, err
	}

	if _, err = hasher.Write([]byte(m.GetAuthEndpoint())); err != nil {
		return 0, err
	}

	if _, err = hasher.Write([]byte(m.GetTokenEndpoint())); err != nil {
		return 0, err
	}

	if _, err = hasher.Write([]byte(m.GetJwksUri())); err != nil {
		return 0, err
	}

	for _, v := range m.GetScopes() {

		if _, err = hasher.Write([]byte(v)); err != nil {
			return 0, err
		}

	}

	for _, v := range m.GetResponseTypes() {

		if _, err = hasher.Write([]byte(v)); err != nil {
			return 0, err
		}

	}

	for _, v := range m.GetSubjects() {

		if _, err = hasher.Write([]byte(v)); err != nil {
			return 0, err
		}

	}

	for _, v := range m.GetIdTokenAlgs() {

		if _, err = hasher.Write([]byte(v)); err != nil {
			return 0, err
		}

	}

	for _, v := range m.GetAuthMethods() {

		if _, err = hasher.Write([]byte(v)); err != nil {
			return 0, err
		}

	}

	for _, v := range m.GetClaims() {

		if _, err = hasher.Write([]byte(v)); err != nil {
			return 0, err
		}

	}

	return hasher.Sum64(), nil
}
