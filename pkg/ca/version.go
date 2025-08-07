// SPDX-License-Identifier: CC0-1.0

package ca

// Version represents the version of the CA package following Semantic Versioning.
//
// Version History:
//   - v2.0.0: BREAKING: New V2 API with simplified certificate requests, enhanced CN selection
//   - Added RequestCertificateV2() and IssueServiceCertificateV2() with automatic IP detection
//   - BREAKING: CN selection now uses first non-IP SAN or first IP (no .local suffix)
//   - BREAKING: CreateSecureDualProtocolServer() API changed to combine IP and SAN arrays
//   - Added CertRequestV2 struct with simplified ServiceName + SANs format
//   - Enhanced server to auto-detect V1 vs V2 request formats
//   - Added comprehensive validation for empty SANs with proper error handling
//   - v2.0.1: Improved error logging for certificate request validation with request payload details
//   - v2.0.2: CRITICAL FIX: Fixed infinite recursion in dual protocol RemoteAddr() causing stack overflow
//   - v2.0.3: CRITICAL FIX: Fixed TLS handshake in dual protocol server by preserving buffered data
//   - v2.0.4: ENHANCEMENT: Added certificate details printing to CreateSecureDualProtocolServer for debugging
//   - v2.0.5: RELEASE: Certificate details printing feature complete with comprehensive cert information
//   - v2.0.6: ENHANCEMENT: Added CreateHTTPClientWithSystemAndCustomCAs() for TLS certificate handling

// Version of the CA package
const Version = "v2.0.6"
