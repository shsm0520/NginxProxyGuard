<div align="center">

# Nginx Proxy Guard

### Next-Generation Nginx Reverse Proxy Manager with Enterprise Security

**English** | [한국어](./README_KO.md)

[![Nginx](https://img.shields.io/badge/Nginx-1.28.0-009639?style=for-the-badge&logo=nginx&logoColor=white)](https://nginx.org/)
[![ModSecurity](https://img.shields.io/badge/ModSecurity-v3.0.14-red?style=for-the-badge)](https://modsecurity.org/)
[![OWASP CRS](https://img.shields.io/badge/OWASP_CRS-v4.21.0-orange?style=for-the-badge)](https://coreruleset.org/)
[![HTTP/3](https://img.shields.io/badge/HTTP/3-QUIC-blue?style=for-the-badge)]()
[![License](https://img.shields.io/badge/License-MIT-yellow?style=for-the-badge)](LICENSE)

<p align="center">
  <strong>Modern reverse proxy management system with<br/>powerful WAF, Bot Protection, GeoIP Blocking, and Rate Limiting</strong>
</p>

[Demo](#-screenshots) • [Features](#-key-features) • [Install](#-quick-start) • [Docs](#-api-documentation) • [Contribute](#-contributing)

---

</div>

## 📋 Table of Contents

- [Introduction](#-introduction)
- [Screenshots](#-screenshots)
- [Key Features](#-key-features)
  - [Security Processing Order](#️-security-processing-order)
- [Quick Start](#-quick-start)
- [Architecture](#-architecture)
- [Tech Stack](#-tech-stack)
- [Configuration](#-configuration)
- [API Documentation](#-api-documentation)
- [Development](#-development)
- [Security Considerations](#-security-considerations)
- [Troubleshooting](#-troubleshooting)
- [Roadmap](#-roadmap)
- [Contributing](#-contributing)
- [License](#-license)

---

## 🎯 Introduction

**Nginx Proxy Guard** is a next-generation reverse proxy management system based on Nginx. It provides enterprise-grade security features with an intuitive web UI, protecting web applications through comprehensive security layers including ModSecurity WAF, bot protection, GeoIP blocking, and rate limiting.

### Why Nginx Proxy Guard?

| Feature | Nginx Proxy Guard | Other Solutions |
|---------|-----------|-----------------|
| **WAF** | ModSecurity v3 + OWASP CRS v4.21 | Limited/None |
| **Bot Protection** | 200+ bot signatures, AI bot detection | Basic blocking only |
| **GeoIP** | MaxMind integration, country blocking/challenge | Limited |
| **HTTP/3** | Full QUIC support | Mostly unsupported |
| **Challenge System** | reCAPTCHA, hCaptcha, Turnstile | None/Basic |
| **Log Analysis** | GeoIP, visualization, advanced filtering | Basic logs only |
| **Multilingual** | Full Korean/English support | English only |

---

## 📸 Screenshots

<div align="center">

### Dashboard
Real-time traffic monitoring, security events, and system status at a glance

### Proxy Host Management
Multi-domain, SSL, WAF, bot filter, and detailed per-host settings

### WAF Settings
OWASP CRS-based rule management, testing, and exception handling

### Log Viewer
GeoIP information, advanced filtering, real-time visualization charts

### GeoIP Map
Country-wise traffic visualization, interactive world map

</div>

---

## ✨ Key Features

### 🔐 Security Features

<table>
<tr>
<td width="50%">

#### Web Application Firewall (WAF)
- ModSecurity v3.0.14 + OWASP CRS v4.21.0
- Detection/Blocking mode toggle
- Paranoia Level 1-4 settings
- Per-host rule exceptions
- WAF testing interface
- Auto IP blocking (threshold-based)

</td>
<td width="50%">

#### Bot Protection
- 80+ malicious bot signatures
- 50+ AI bot detection (GPTBot, ClaudeBot, etc.)
- Search engine allowlist (Google, Bing, etc.)
- Suspicious client blocking
- Custom User-Agent rules
- Challenge mode support

</td>
</tr>
<tr>
<td>

#### Rate Limiting
- Requests per second limits
- Burst size configuration
- IP/URI/IP+URI based limiting
- Whitelist IP support
- Custom response codes

</td>
<td>

#### Geo Restriction
- Country whitelist/blacklist
- MaxMind GeoIP2 integration
- Priority allow IP settings
- Challenge mode (CAPTCHA instead of block)
- Continent/region presets

</td>
</tr>
<tr>
<td>

#### Challenge System
- reCAPTCHA v2/v3
- hCaptcha
- Cloudflare Turnstile
- Custom challenge pages
- Token validity settings

</td>
<td>

#### IP Blocking
- Fail2ban-style auto blocking
- WAF event-based auto blocking
- Manual IP management
- Temporary/permanent bans
- Block history tracking

</td>
</tr>
</table>

### 🛡️ Security Processing Order

Requests pass through multiple security layers before reaching the backend server. For performance optimization, GeoIP blocking is processed before WAF.

#### Direct Block Mode (GeoIP Challenge Disabled)

```
Request → GeoIP Block → WAF → Access List → Exploit Block → Bot Filter → URI Block → Backend
            ↓
      [Blocked Country]
            ↓
      Return 403 (Performance boost by skipping WAF)
```

| Order | Security Layer | Description |
|-------|---------------|-------------|
| 1 | **GeoIP Blocking** | Country-based blocking (before WAF) |
| 2 | **WAF (ModSecurity)** | OWASP CRS attack detection/blocking |
| 3 | **Access List** | IP-based allow/deny lists |
| 4 | **Exploit Block** | SQL Injection, XSS, RFI blocking |
| 5 | **Bot Filter** | Malicious bots, AI bots blocking |
| 6 | **URI Block** | Path-based blocking (/wp-login.php, etc.) |

#### Challenge Mode (GeoIP Challenge Enabled)

```
Request → WAF → GeoIP Setup → Exploit Block → Bot Filter → URI Block → Challenge Verify → Backend
                   ↓                                            ↓
            [Blocked Country]                           [No/Invalid Token]
                   ↓                                            ↓
            Set $geo_blocked=1                        Redirect to CAPTCHA page
```

| Order | Security Layer | Description |
|-------|---------------|-------------|
| 1 | **WAF (ModSecurity)** | Attack detection/blocking (first) |
| 2 | **GeoIP Variable Set** | Only marks block status, no immediate block |
| 3 | **Exploit Block** | Exploit pattern blocking |
| 4 | **Bot Filter** | Bot filtering |
| 5 | **URI Block** | Path-based blocking |
| 6 | **Challenge Verify** | Token validation, CAPTCHA on failure |

#### Exception Handling (Bypass)

The following conditions bypass blocking in each security layer:

| Exception | Applies To | Description |
|-----------|-----------|-------------|
| **Allow Search Bots** | GeoIP, Bot Filter | Googlebot, Bingbot, etc. |
| **Allow Private IPs** | GeoIP | 10.x, 172.16-31.x, 192.168.x |
| **Priority Allow IP/CIDR** | GeoIP | Specified IPs/ranges always allowed |
| **URI Exceptions** | Exploit Block | /wp-json/, /api/, etc. |

> **Performance Tip**: Direct Block mode reduces server load as blocked requests skip WAF processing. Challenge mode processes WAF first to log attack attempts.

### 🌐 Proxy Features

<table>
<tr>
<td width="50%">

#### Proxy Host Management
- Multiple domains per host
- HTTP/HTTPS forwarding
- WebSocket upgrade
- Custom location blocks
- Advanced nginx config
- Host testing/validation

</td>
<td width="50%">

#### SSL/TLS Certificates
- Let's Encrypt auto issue/renewal
- DNS-01 challenge (Cloudflare, Route53, etc.)
- Custom certificate upload
- HTTP/2 & HTTP/3 (QUIC)
- Certificate expiry monitoring

</td>
</tr>
<tr>
<td>

#### Upstream/Load Balancing
- Multiple backend servers
- Round Robin, Least Conn, IP Hash
- Health checks
- Server weights
- Keepalive connections

</td>
<td>

#### Security Headers
- HSTS (preload support)
- X-Frame-Options
- X-Content-Type-Options
- Content-Security-Policy
- Custom header support

</td>
</tr>
</table>

### 📊 Monitoring & Analytics

<table>
<tr>
<td width="50%">

#### Dashboard
- Real-time system status
- 24-hour request statistics
- Bandwidth monitoring
- Security event summary
- Interactive world map
- Docker container status

</td>
<td width="50%">

#### Log Viewer
- Access/Error/WAF log integration
- GeoIP information display
- Advanced filtering & search
- Exclusion filter support
- Visualization charts
- Raw log file access

</td>
</tr>
<tr>
<td>

#### Audit Logging
- Full admin activity tracking
- API token usage logging
- Configuration change history
- IP/User-Agent tracking

</td>
<td>

#### System Logs
- Docker container logs
- Source/severity filters
- Auto cleanup policies
- Pattern-based exclusion

</td>
</tr>
</table>

### 🛠️ System Management

<table>
<tr>
<td width="50%">

#### Authentication & Authorization
- JWT-based authentication
- Two-factor authentication (TOTP)
- API tokens (granular permissions)
- IP restriction, expiration dates
- Usage tracking

</td>
<td width="50%">

#### Backup & Restore
- Full system backup
- Scheduled backups
- One-click restore
- Retention policies
- Config/certificates/database included

</td>
</tr>
<tr>
<td>

#### Multilingual Support
- Full Korean support
- Full English support
- Browser auto-detection
- Per-user language settings

</td>
<td>

#### UI/UX
- Dark mode
- Responsive design
- Intuitive interface
- Real-time updates

</td>
</tr>
</table>

---

## 🚀 Quick Start

### Prerequisites

- Docker 24.0+ and Docker Compose v2
- (Optional) [MaxMind License Key](https://www.maxmind.com/en/geolite2/signup) for GeoIP

### Installation

```bash
# 1. Create directory
mkdir -p ~/nginx-proxy-guard && cd ~/nginx-proxy-guard

# 2. Download files
wget https://raw.githubusercontent.com/svrforum/nginxproxyguard/main/docker-compose.yml
wget -O .env https://raw.githubusercontent.com/svrforum/nginxproxyguard/main/.env.example

# 3. Auto-generate secure secrets
sed -i "s/DB_PASSWORD=.*/DB_PASSWORD=$(openssl rand -base64 24)/" .env
sed -i "s/JWT_SECRET=.*/JWT_SECRET=$(openssl rand -hex 32)/" .env

# 4. Auto-detect timezone
TZ=$(cat /etc/timezone 2>/dev/null || readlink /etc/localtime | sed 's|/usr/share/zoneinfo/||' 2>/dev/null || echo "UTC")
sed -i "s|TZ=.*|TZ=$TZ|" .env

# 5. Start services
docker compose up -d
```

### Access

| Service | URL |
|---------|-----|
| Admin Panel | https://localhost:81 |
| HTTP Proxy | http://localhost:80 |
| HTTPS Proxy | https://localhost:443 |

**Default Login**: `admin` / `admin` (Change immediately after first login!)

### Update

```bash
docker compose pull
docker compose up -d
```

### Uninstall

```bash
# Stop containers
docker compose down

# Remove data (optional)
docker volume rm npg_postgres_data npg_valkey_data npg_nginx_data npg_api_data npg_ui_data
```

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                             Internet                                  │
└─────────────────────────────┬───────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    Nginx Proxy (Port 80/443)                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌────────────┐ │
│  │ Rate Limit  │  │ModSecurity  │  │  Bot Filter │  │ Geo Block  │ │
│  │             │  │ WAF+CRS4    │  │  200+ Sigs  │  │  GeoIP2    │ │
│  └─────────────┘  └─────────────┘  └─────────────┘  └────────────┘ │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌────────────┐ │
│  │  SSL/TLS    │  │  Fail2ban   │  │  Challenge  │  │ URI Block  │ │
│  │HTTP/2 HTTP/3│  │ Auto-Ban    │  │  CAPTCHA    │  │            │ │
│  └─────────────┘  └─────────────┘  └─────────────┘  └────────────┘ │
└─────────────────────────────┬───────────────────────────────────────┘
                              │
          ┌───────────────────┼───────────────────┐
          ▼                   ▼                   ▼
    ┌───────────┐       ┌───────────┐       ┌───────────┐
    │ Backend 1 │       │ Backend 2 │       │ Backend N │
    │           │       │           │       │           │
    └───────────┘       └───────────┘       └───────────┘

┌─────────────────────────────────────────────────────────────────────┐
│                        Management Layer                               │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌────────────┐ │
│  │  React UI   │◄─│   Go API    │◄─│ PostgreSQL  │  │  Valkey    │ │
│  │  (Port 81)  │  │ (Port 8080) │  │    17       │  │ (Redis)    │ │
│  │ TypeScript  │  │  Echo v4    │  │             │  │  Cache     │ │
│  │ Tailwind    │  │  JWT+TOTP   │  │             │  │  Ban List  │ │
│  └─────────────┘  └─────────────┘  └─────────────┘  └────────────┘ │
└─────────────────────────────────────────────────────────────────────┘
```

### Docker Services

| Service | Image | Description | Ports |
|---------|-------|-------------|-------|
| `nginx` | Custom Build | Reverse Proxy + WAF + GeoIP | 80, 443 |
| `ui` | Custom Build | React Web Interface | 81 |
| `api` | Custom Build | Go Backend API | 8080 (internal) |
| `db` | postgres:17-alpine | Database | 5432 (internal) |
| `valkey` | valkey/valkey:8-alpine | Cache (Redis compatible) | 6379 (internal) |

---

## 🔧 Tech Stack

### Backend

| Technology | Version | Purpose |
|------------|---------|---------|
| Go | 1.22+ | API Server |
| Echo | v4 | Web Framework |
| PostgreSQL | 17 | Main Database |
| Valkey | 8 | Cache & Session |
| JWT | - | Auth Tokens |
| TOTP | - | 2FA |

### Frontend

| Technology | Version | Purpose |
|------------|---------|---------|
| React | 18 | UI Framework |
| TypeScript | 5 | Type Safety |
| Tailwind CSS | 3 | Styling |
| TanStack Query | 5 | Server State Management |
| Vite | 5 | Build Tool |
| react-simple-maps | - | Map Visualization |

### Proxy

| Technology | Version | Purpose |
|------------|---------|---------|
| Nginx | 1.28.0 | Reverse Proxy |
| ModSecurity | 3.0.14 | WAF Engine |
| OWASP CRS | 4.21.0 | WAF Ruleset |
| MaxMind GeoIP2 | - | Geo Information |
| Brotli/Zstd | - | Compression |

---

## ⚙️ Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `DB_PASSWORD` | Database password | - | ✅ |
| `JWT_SECRET` | JWT signing secret | - | ✅ |
| `DB_USER` | DB user | `postgres` | |
| `DB_NAME` | DB name | `nginx_proxy_guard` | |
| `MAXMIND_LICENSE_KEY` | GeoIP license | - | |
| `MAXMIND_ACCOUNT_ID` | GeoIP account ID | - | |
| `ACME_EMAIL` | Let's Encrypt email | - | |
| `ACME_STAGING` | Staging mode | `false` | |

### Database Tables

Main table structure:

| Table | Description |
|-------|-------------|
| `users` | User authentication (with TOTP) |
| `proxy_hosts` | Proxy host settings |
| `certificates` | SSL certificates |
| `logs` | Unified logs (with GeoIP) |
| `banned_ips` | Blocked IPs |
| `uri_blocks` | URI block rules |
| `global_uri_blocks` | Global URI blocks |
| `rate_limits` | Rate limiting settings |
| `fail2ban_configs` | Fail2ban settings |
| `bot_filters` | Bot filter rules |
| `geo_restrictions` | Geographic restrictions |
| `waf_rule_exclusions` | WAF rule exceptions |
| `system_settings` | System settings |
| `audit_logs` | Audit logs |

---

## 📖 API Documentation

### Authentication

All APIs require authentication (except `/api/v1/auth/login`):

```bash
# JWT Token
Authorization: Bearer <jwt_token>

# API Token
Authorization: Bearer ng_<api_token>
```

### Main Endpoints

#### Authentication

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/login` | Login |
| POST | `/api/v1/auth/logout` | Logout |
| GET | `/api/v1/auth/me` | Current user |
| POST | `/api/v1/auth/2fa/setup` | 2FA setup |

#### Proxy Hosts

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/proxy-hosts` | List |
| POST | `/api/v1/proxy-hosts` | Create |
| GET | `/api/v1/proxy-hosts/:id` | Get details |
| PUT | `/api/v1/proxy-hosts/:id` | Update |
| DELETE | `/api/v1/proxy-hosts/:id` | Delete |
| POST | `/api/v1/proxy-hosts/:id/test` | Test |

#### Security Settings (per host)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET/PUT | `/api/v1/proxy-hosts/:id/rate-limit` | Rate Limiting |
| GET/PUT | `/api/v1/proxy-hosts/:id/fail2ban` | Fail2ban |
| GET/PUT | `/api/v1/proxy-hosts/:id/bot-filter` | Bot Filter |
| GET/PUT | `/api/v1/proxy-hosts/:id/waf` | WAF Settings |
| GET/PUT | `/api/v1/proxy-hosts/:id/geo` | Geo Restriction |
| GET/PUT | `/api/v1/proxy-hosts/:id/uri-block` | URI Block |

#### Global Settings

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET/PUT | `/api/v1/settings` | Global Nginx Settings |
| GET/PUT | `/api/v1/system-settings` | System Settings |
| GET/PUT | `/api/v1/security/global-banned-ips` | Global IP Block |
| GET/PUT | `/api/v1/security/global-uri-block` | Global URI Block |

#### Logs & Monitoring

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/dashboard` | Dashboard Data |
| GET | `/api/v1/dashboard/geoip` | GeoIP Stats |
| GET | `/api/v1/logs` | Query Logs |
| GET | `/api/v1/logs/stats` | Log Statistics |
| GET | `/api/v1/audit-logs` | Audit Logs |

---

## 💻 Development

### Local Development Environment

#### Backend

```bash
cd api
go mod download
go run ./cmd/server
```

#### Frontend

```bash
cd ui
npm install
npm run dev
```

### Docker Development Environment

```bash
# Run in development mode (hot reloading)
docker compose -f docker-compose.dev.yml up -d --build
```

### Project Structure

```
nginx-proxy-guard/
├── api/                    # Go Backend
│   ├── cmd/server/         # Main entrypoint
│   ├── internal/
│   │   ├── handler/        # HTTP handlers
│   │   ├── model/          # Data models
│   │   ├── repository/     # DB access layer
│   │   ├── service/        # Business logic
│   │   ├── database/       # DB migrations
│   │   └── nginx/          # Nginx config generation
│   └── pkg/                # Shared packages
├── ui/                     # React Frontend
│   ├── src/
│   │   ├── components/     # React components
│   │   ├── api/            # API client
│   │   ├── hooks/          # Custom hooks
│   │   ├── i18n/           # Multilingual resources
│   │   └── types/          # TypeScript types
│   └── public/
├── nginx/                  # Nginx Configuration
│   ├── nginx.conf          # Main config
│   ├── modsec/             # ModSecurity config
│   ├── includes/           # Common configs
│   └── scripts/            # Helper scripts
└── docker-compose.yml      # Docker composition
```

---

## 🔒 Security Considerations

### Essential Actions

1. **Change Default Password**: Change `admin` password immediately after installation
2. **Use Strong Secrets**: `JWT_SECRET`, `DB_PASSWORD` should be random strings of 32+ characters
3. **Enable 2FA**: Set up TOTP for all admin accounts
4. **API Token Management**: Set IP restrictions and expiration dates

### Recommended Actions

5. **Network Isolation**: Management port (81) accessible only from internal network
6. **Regular Backups**: Configure automated backups and test restores
7. **Log Monitoring**: Regular review of audit logs
8. **Keep Updated**: Regular updates for security patches

### WAF Operation

- **Detection Mode First**: Start new hosts in Detection mode
- **Gradual Hardening**: Switch to Blocking mode after confirming false positives
- **Rule Exception Management**: Only disable rules when necessary

---

## 🔧 Troubleshooting

### Common Issues

<details>
<summary><b>Certificate Issuance Failure</b></summary>

1. Verify DNS settings (domain pointing to server IP)
2. Check DNS Provider API token permissions
3. Check Let's Encrypt Rate Limits
4. Test in staging environment first

```bash
# Check certificate logs
docker logs nginx-proxy-guard-api | grep -i cert
```
</details>

<details>
<summary><b>WAF False Positives</b></summary>

1. Use WAF test feature to identify triggering rules
2. Add rule ID to per-host exception list
3. Switch to detection mode for monitoring before applying blocking mode
</details>

<details>
<summary><b>GeoIP Not Working</b></summary>

1. Verify MaxMind license key
2. Check GeoIP database exists

```bash
docker exec nginx-proxy-guard-proxy ls -la /etc/nginx/geoip/
```

3. Restart Nginx
```bash
docker compose restart nginx
```
</details>

<details>
<summary><b>Slow Log Queries</b></summary>

1. Reduce log retention period (Settings > System Settings)
2. Narrow filter range
3. Check partition status

```sql
-- Check partitions
SELECT * FROM pg_partitions WHERE tablename = 'logs_partitioned';
```
</details>

---

## 🗺️ Roadmap

### In Progress

- [ ] Cluster mode support
- [ ] Prometheus metrics export
- [ ] Additional language support

### Planned

- [ ] Kubernetes Helm chart
- [ ] API rate limiting (per token)
- [ ] Email notification system
- [ ] User role management

---

## 🤝 Contributing

Contributions are welcome! Please refer to the following guidelines:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Apply `gofmt` to Go code
- Use TypeScript for React components
- Follow [Conventional Commits](https://www.conventionalcommits.org/)
- Ensure tests pass before PR

---

## 📄 License

This project is distributed under the MIT License. See [LICENSE](LICENSE) file for details.

---

## 📞 Support

- **GitHub Issues**: Bug reports and feature requests
- **Discussions**: Questions and discussions

---

<div align="center">

**[⬆ Back to Top](#nginx-proxy-guard)**

Made with ❤️ by the Nginx Proxy Guard Team

</div>
