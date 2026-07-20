# Cosign Signature Verification

All pgSCV docker images published to DockerHub are signed with [Cosign](https://docs.sigstore.dev/cosign/). Every release is signed with the same key introduced in version 0.15.3. This document provides comprehensive guidance on verifying images signatures.

## Overview

Every pgSCV docker images published to `registry-1.docker.io/cherts/pgscv` is signed with our private key and can be verified using the corresponding public key. This ensures:

- **Authenticity**: Confirms docker images are published by Mikhail Grigorev
- **Integrity**: Ensures docker images haven't been tampered with since signing
- **Supply Chain Security**: Provides end-to-end verification of images origins

## Public Key

All pgSCV docker images are signed with the following Cosign public key:

**Download:** [cosign.pub](https://raw.githubusercontent.com/cherts/pgscv/refs/heads/release/0.15/cosign.pub)

```
-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEOhKbSPiK5NZF40ZEeio+Vf4s7eQP
yjhhbVVDCvUcluVIPQZLFB4F4o1jxkpRwYQ0wj+JHai/b+efFC1XrJJwWQ==
-----END PUBLIC KEY-----
```

## Manual Verification

### Prerequisites

Install Cosign on your system:

```bash
# macOS (using Homebrew)
brew install cosign

# Linux (using curl)
curl -O -L "https://github.com/sigstore/cosign/releases/latest/download/cosign-linux-amd64"
sudo mv cosign-linux-amd64 /usr/local/bin/cosign
sudo chmod +x /usr/local/bin/cosign

# Windows (using winget)
winget install sigstore.cosign
```

### Step-by-Step Verification

1. **Download the public key:**

   ```bash
   # Option 1: Download directly from GitHub
   curl -o cosign.pub https://raw.githubusercontent.com/cherts/pgscv/refs/heads/release/0.15/cosign.pub

   # Option 2: Create manually
   cat > cosign.pub << 'EOF'
   -----BEGIN PUBLIC KEY-----
   MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEOhKbSPiK5NZF40ZEeio+Vf4s7eQP
   yjhhbVVDCvUcluVIPQZLFB4F4o1jxkpRwYQ0wj+JHai/b+efFC1XrJJwWQ==
   -----END PUBLIC KEY-----
   EOF
   ```

2. **Verify a specific chart:**

   ```bash
   # Replace <version> with actual values
   cosign verify --key cosign.pub registry-1.docker.io/cherts/pgscv:<version>

   # Examples:
   cosign verify --key cosign.pub registry-1.docker.io/cherts/pgscv:v0.15.3
   ```

3. **Successful verification output:**
   ```
   Verification for registry-1.docker.io/cherts/pgscv:v0.15.3 --
   The following checks were performed on each of these signatures:
     - The cosign claims were validated
     - Existence of the claims in the transparency log was verified offline
     - The signatures were verified against the specified public key
   ```

## Additional Resources

- [Sigstore Documentation](https://docs.sigstore.dev/)
- [Cosign Installation Guide](https://docs.sigstore.dev/cosign/installation)
- [Supply Chain Security Best Practices](https://slsa.dev/)