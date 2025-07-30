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
//   - v1.8.0: Added dual protocol transport server for handling HTTP and HTTPS on same port
//   - v1.7.0: Added Google Cloud emulator environment variable detection
//   - v1.6.0: Added UpdateTransportMust function for panic-based transport updates
//   - v1.5.0: HTTPS-only server enforcement, API returns SecureHTTPSServer
const Version = "2.0.0"
