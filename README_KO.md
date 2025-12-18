<div align="center">

# Nginx Proxy Guard

### Next-Generation Nginx Reverse Proxy Manager with Enterprise Security

[English](./README.md) | **한국어**

[![Nginx](https://img.shields.io/badge/Nginx-1.28.0-009639?style=for-the-badge&logo=nginx&logoColor=white)](https://nginx.org/)
[![ModSecurity](https://img.shields.io/badge/ModSecurity-v3.0.14-red?style=for-the-badge)](https://modsecurity.org/)
[![OWASP CRS](https://img.shields.io/badge/OWASP_CRS-v4.21.0-orange?style=for-the-badge)](https://coreruleset.org/)
[![HTTP/3](https://img.shields.io/badge/HTTP/3-QUIC-blue?style=for-the-badge)]()
[![License](https://img.shields.io/badge/License-MIT-yellow?style=for-the-badge)](LICENSE)

<p align="center">
  <strong>강력한 WAF, 봇 보호, GeoIP 차단, Rate Limiting을 갖춘<br/>현대적인 리버스 프록시 관리 시스템</strong>
</p>

[데모](#-스크린샷) • [기능](#-주요-기능) • [설치](#-빠른-시작) • [문서](#-api-문서) • [기여](#-기여)

---

</div>

## 📋 목차

- [소개](#-소개)
- [스크린샷](#-스크린샷)
- [주요 기능](#-주요-기능)
  - [보안 처리 순서](#️-보안-처리-순서)
- [빠른 시작](#-빠른-시작)
- [아키텍처](#-아키텍처)
- [기술 스택](#-기술-스택)
- [설정](#-설정)
- [API 문서](#-api-문서)
- [개발](#-개발)
- [보안 고려사항](#-보안-고려사항)
- [문제 해결](#-문제-해결)
- [로드맵](#-로드맵)
- [기여](#-기여)
- [라이선스](#-라이선스)

---

## 🎯 소개

**Nginx Proxy Guard**는 Nginx 기반의 차세대 리버스 프록시 관리 시스템입니다. 직관적인 웹 UI와 함께 엔터프라이즈급 보안 기능을 제공하며, ModSecurity WAF, 봇 보호, GeoIP 차단, Rate Limiting 등 포괄적인 보안 레이어를 통해 웹 애플리케이션을 보호합니다.

### 왜 Nginx Proxy Guard인가?

| 특징 | Nginx Proxy Guard | 기타 솔루션 |
|------|-----------|-------------|
| **WAF** | ModSecurity v3 + OWASP CRS v4.21 | 제한적/없음 |
| **봇 보호** | 200+ 봇 시그니처, AI 봇 탐지 | 기본적인 차단만 |
| **GeoIP** | MaxMind 통합, 국가별 차단/챌린지 | 제한적 |
| **HTTP/3** | QUIC 완벽 지원 | 대부분 미지원 |
| **챌린지 시스템** | reCAPTCHA, hCaptcha, Turnstile | 없음/기본적 |
| **로그 분석** | GeoIP, 시각화, 고급 필터링 | 기본 로그만 |
| **다국어** | 한국어/영어 완벽 지원 | 영어만 |

---

## 📸 스크린샷

<div align="center">

### 대시보드
실시간 트래픽 모니터링, 보안 이벤트, 시스템 상태를 한눈에 확인

### 프록시 호스트 관리
다중 도메인, SSL, WAF, 봇 필터 등 호스트별 상세 설정

### WAF 설정
OWASP CRS 기반 룰 관리, 테스트, 예외 처리

### 로그 뷰어
GeoIP 정보, 고급 필터링, 실시간 시각화 차트

### GeoIP 지도
국가별 트래픽 시각화, 인터랙티브 세계지도

</div>

---

## ✨ 주요 기능

### 🔐 보안 기능

<table>
<tr>
<td width="50%">

#### 웹 애플리케이션 방화벽 (WAF)
- ModSecurity v3.0.14 + OWASP CRS v4.21.0
- 탐지/차단 모드 전환
- Paranoia Level 1-4 설정
- 호스트별 룰 예외 처리
- WAF 테스트 인터페이스
- 자동 IP 차단 (임계값 기반)

</td>
<td width="50%">

#### 봇 보호
- 80+ 악성 봇 시그니처
- 50+ AI 봇 탐지 (GPTBot, ClaudeBot 등)
- 검색 엔진 허용 (Google, Bing 등)
- 의심스러운 클라이언트 차단
- 커스텀 User-Agent 규칙
- 챌린지 모드 지원

</td>
</tr>
<tr>
<td>

#### Rate Limiting
- 초당 요청 수 제한
- 버스트 크기 설정
- IP/URI/IP+URI 기반 제한
- 화이트리스트 IP 지원
- 커스텀 응답 코드

</td>
<td>

#### Geo 제한
- 국가별 화이트리스트/블랙리스트
- MaxMind GeoIP2 통합
- 우선 허용 IP 설정
- 챌린지 모드 (차단 대신 CAPTCHA)
- 대륙/지역 프리셋

</td>
</tr>
<tr>
<td>

#### 챌린지 시스템
- reCAPTCHA v2/v3
- hCaptcha
- Cloudflare Turnstile
- 커스텀 챌린지 페이지
- 토큰 유효기간 설정

</td>
<td>

#### IP 차단
- Fail2ban 스타일 자동 차단
- WAF 이벤트 기반 자동 차단
- 수동 IP 관리
- 임시/영구 차단
- 차단 이력 추적

</td>
</tr>
</table>

### 🛡️ 보안 처리 순서

요청이 백엔드 서버에 도달하기 전에 여러 보안 레이어를 통과합니다. 성능 최적화를 위해 GeoIP 차단은 WAF보다 먼저 처리됩니다.

#### Direct Block 모드 (GeoIP 챌린지 비활성화)

```
요청 → GeoIP 차단 → WAF → Access List → Exploit Block → Bot Filter → URI 차단 → 백엔드
         ↓
    [차단된 국가]
         ↓
     403 반환 (WAF 처리 생략으로 성능 향상)
```

| 순서 | 보안 레이어 | 설명 |
|------|------------|------|
| 1 | **GeoIP 차단** | 국가 기반 차단 (WAF 전에 처리) |
| 2 | **WAF (ModSecurity)** | OWASP CRS 기반 공격 탐지/차단 |
| 3 | **Access List** | IP 기반 허용/거부 목록 |
| 4 | **Exploit Block** | SQL Injection, XSS, RFI 등 차단 |
| 5 | **Bot Filter** | 악성 봇, AI 봇 차단 |
| 6 | **URI 차단** | 특정 경로 차단 (/wp-login.php 등) |

#### Challenge 모드 (GeoIP 챌린지 활성화)

```
요청 → WAF → GeoIP 설정 → Exploit Block → Bot Filter → URI 차단 → Challenge 검증 → 백엔드
                ↓                                            ↓
         [차단된 국가]                               [토큰 없음/무효]
                ↓                                            ↓
         $geo_blocked=1 설정                          CAPTCHA 페이지로 리다이렉트
```

| 순서 | 보안 레이어 | 설명 |
|------|------------|------|
| 1 | **WAF (ModSecurity)** | 공격 탐지/차단 (먼저 처리) |
| 2 | **GeoIP 변수 설정** | 차단 여부만 표시, 즉시 차단 안함 |
| 3 | **Exploit Block** | 익스플로잇 패턴 차단 |
| 4 | **Bot Filter** | 봇 필터링 |
| 5 | **URI 차단** | 경로 기반 차단 |
| 6 | **Challenge 검증** | 토큰 유효성 검증, 실패 시 CAPTCHA |

#### 예외 처리 (Bypass)

각 보안 레이어에서 다음 조건은 차단을 우회합니다:

| 예외 조건 | 적용 대상 | 설명 |
|----------|----------|------|
| **검색봇 허용** | GeoIP, Bot Filter | Googlebot, Bingbot 등 검색 엔진 |
| **Private IP 허용** | GeoIP | 10.x, 172.16-31.x, 192.168.x |
| **우선 허용 IP/CIDR** | GeoIP | 지정된 IP/대역 무조건 허용 |
| **URI 예외** | Exploit Block | /wp-json/, /api/ 등 지정 경로 |

> **성능 팁**: Direct Block 모드는 차단된 요청이 WAF를 거치지 않아 서버 부하가 감소합니다. Challenge 모드는 WAF를 먼저 통과시켜 공격 시도도 기록할 수 있습니다.

### 🌐 프록시 기능

<table>
<tr>
<td width="50%">

#### 프록시 호스트 관리
- 호스트당 다중 도메인
- HTTP/HTTPS 포워딩
- WebSocket 업그레이드
- 커스텀 location 블록
- 고급 nginx 설정
- 호스트 테스트/검증

</td>
<td width="50%">

#### SSL/TLS 인증서
- Let's Encrypt 자동 발급/갱신
- DNS-01 챌린지 (Cloudflare, Route53 등)
- 커스텀 인증서 업로드
- HTTP/2 & HTTP/3 (QUIC)
- 인증서 만료 모니터링

</td>
</tr>
<tr>
<td>

#### 업스트림/로드 밸런싱
- 다중 백엔드 서버
- Round Robin, Least Conn, IP Hash
- 헬스 체크
- 서버 가중치
- Keepalive 연결

</td>
<td>

#### 보안 헤더
- HSTS (preload 지원)
- X-Frame-Options
- X-Content-Type-Options
- Content-Security-Policy
- 커스텀 헤더 지원

</td>
</tr>
</table>

### 📊 모니터링 & 분석

<table>
<tr>
<td width="50%">

#### 대시보드
- 실시간 시스템 상태
- 24시간 요청 통계
- 대역폭 모니터링
- 보안 이벤트 요약
- 인터랙티브 세계지도
- Docker 컨테이너 상태

</td>
<td width="50%">

#### 로그 뷰어
- 접근/오류/WAF 로그 통합
- GeoIP 정보 표시
- 고급 필터링 & 검색
- 제외 필터 지원
- 시각화 차트
- 원시 로그 파일 접근

</td>
</tr>
<tr>
<td>

#### 감사 로깅
- 전체 관리자 활동 추적
- API 토큰 사용 로깅
- 설정 변경 기록
- IP/User-Agent 추적

</td>
<td>

#### 시스템 로그
- Docker 컨테이너 로그
- 소스/심각도별 필터
- 자동 정리 정책
- 패턴 기반 제외

</td>
</tr>
</table>

### 🛠️ 시스템 관리

<table>
<tr>
<td width="50%">

#### 인증 & 권한
- JWT 기반 인증
- 2단계 인증 (TOTP)
- API 토큰 (세분화된 권한)
- IP 제한, 만료일 설정
- 사용량 추적

</td>
<td width="50%">

#### 백업 & 복원
- 전체 시스템 백업
- 예약 백업
- 원클릭 복원
- 보존 정책
- 설정/인증서/데이터베이스 포함

</td>
</tr>
<tr>
<td>

#### 다국어 지원
- 한국어 완벽 지원
- 영어 완벽 지원
- 브라우저 자동 감지
- 사용자별 언어 설정

</td>
<td>

#### UI/UX
- 다크 모드
- 반응형 디자인
- 직관적인 인터페이스
- 실시간 업데이트

</td>
</tr>
</table>

---

## 🚀 빠른 시작

### 필요 조건

- Docker 24.0+ 및 Docker Compose v2
- (선택) GeoIP용 [MaxMind 라이선스 키](https://www.maxmind.com/en/geolite2/signup)

### 설치

```bash
# 1. 디렉토리 생성
mkdir -p ~/nginx-proxy-guard && cd ~/nginx-proxy-guard

# 2. 파일 다운로드
wget https://raw.githubusercontent.com/svrforum/nginxproxyguard/main/docker-compose.yml
wget -O .env https://raw.githubusercontent.com/svrforum/nginxproxyguard/main/.env.example

# 3. 보안 시크릿 자동 생성
sed -i "s/DB_PASSWORD=.*/DB_PASSWORD=$(openssl rand -base64 24)/" .env
sed -i "s/JWT_SECRET=.*/JWT_SECRET=$(openssl rand -hex 32)/" .env

# 4. 시간대 자동 감지
TZ=$(cat /etc/timezone 2>/dev/null || readlink /etc/localtime | sed 's|/usr/share/zoneinfo/||' 2>/dev/null || echo "UTC")
sed -i "s|TZ=.*|TZ=$TZ|" .env

# 5. 서비스 시작
docker compose up -d
```

### 접속

| 서비스 | URL |
|--------|-----|
| 관리 패널 | https://localhost:81 |
| HTTP 프록시 | http://localhost:80 |
| HTTPS 프록시 | https://localhost:443 |

**기본 로그인**: `admin` / `admin` (첫 로그인 후 반드시 변경!)

### 업데이트

```bash
docker compose pull
docker compose up -d
```

### 제거

```bash
# 컨테이너 중지
docker compose down

# 데이터 삭제 (선택)
docker volume rm npg_postgres_data npg_valkey_data npg_nginx_data npg_api_data npg_ui_data
```

---

## 🏗️ 아키텍처

```
┌─────────────────────────────────────────────────────────────────────┐
│                            인터넷                                    │
└─────────────────────────────┬───────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    Nginx Proxy (포트 80/443)                         │
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
│                        관리 레이어                                   │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌────────────┐ │
│  │  React UI   │◄─│   Go API    │◄─│ PostgreSQL  │  │  Valkey    │ │
│  │  (포트 81)  │  │ (포트 8080) │  │    17       │  │ (Redis)    │ │
│  │ TypeScript  │  │  Echo v4    │  │             │  │  Cache     │ │
│  │ Tailwind    │  │  JWT+TOTP   │  │             │  │  Ban List  │ │
│  └─────────────┘  └─────────────┘  └─────────────┘  └────────────┘ │
└─────────────────────────────────────────────────────────────────────┘
```

### Docker 서비스

| 서비스 | 이미지 | 설명 | 포트 |
|--------|--------|------|------|
| `nginx` | Custom Build | 리버스 프록시 + WAF + GeoIP | 80, 443 |
| `ui` | Custom Build | React 웹 인터페이스 | 81 |
| `api` | Custom Build | Go 백엔드 API | 8080 (내부) |
| `db` | postgres:17-alpine | 데이터베이스 | 5432 (내부) |
| `valkey` | valkey/valkey:8-alpine | 캐시 (Redis 호환) | 6379 (내부) |

---

## 🔧 기술 스택

### 백엔드

| 기술 | 버전 | 용도 |
|------|------|------|
| Go | 1.22+ | API 서버 |
| Echo | v4 | 웹 프레임워크 |
| PostgreSQL | 17 | 메인 데이터베이스 |
| Valkey | 8 | 캐시 & 세션 |
| JWT | - | 인증 토큰 |
| TOTP | - | 2단계 인증 |

### 프론트엔드

| 기술 | 버전 | 용도 |
|------|------|------|
| React | 18 | UI 프레임워크 |
| TypeScript | 5 | 타입 안전성 |
| Tailwind CSS | 3 | 스타일링 |
| TanStack Query | 5 | 서버 상태 관리 |
| Vite | 5 | 빌드 도구 |
| react-simple-maps | - | 지도 시각화 |

### 프록시

| 기술 | 버전 | 용도 |
|------|------|------|
| Nginx | 1.28.0 | 리버스 프록시 |
| ModSecurity | 3.0.14 | WAF 엔진 |
| OWASP CRS | 4.21.0 | WAF 룰셋 |
| MaxMind GeoIP2 | - | 지리 정보 |
| Brotli/Zstd | - | 압축 |

---

## ⚙️ 설정

### 환경 변수

| 변수 | 설명 | 기본값 | 필수 |
|------|------|--------|------|
| `DB_PASSWORD` | 데이터베이스 비밀번호 | - | ✅ |
| `JWT_SECRET` | JWT 서명 시크릿 | - | ✅ |
| `DB_USER` | DB 사용자 | `postgres` | |
| `DB_NAME` | DB 이름 | `nginx_proxy_guard` | |
| `MAXMIND_LICENSE_KEY` | GeoIP 라이선스 | - | |
| `MAXMIND_ACCOUNT_ID` | GeoIP 계정 ID | - | |
| `ACME_EMAIL` | Let's Encrypt 이메일 | - | |
| `ACME_STAGING` | 스테이징 모드 | `false` | |

### 데이터베이스 테이블

주요 테이블 구조:

| 테이블 | 설명 |
|--------|------|
| `users` | 사용자 인증 (TOTP 포함) |
| `proxy_hosts` | 프록시 호스트 설정 |
| `certificates` | SSL 인증서 |
| `logs` | 통합 로그 (GeoIP 포함) |
| `banned_ips` | 차단된 IP |
| `uri_blocks` | URI 차단 규칙 |
| `global_uri_blocks` | 전역 URI 차단 |
| `rate_limits` | Rate Limiting 설정 |
| `fail2ban_configs` | Fail2ban 설정 |
| `bot_filters` | 봇 필터 규칙 |
| `geo_restrictions` | 지역 제한 |
| `waf_rule_exclusions` | WAF 룰 예외 |
| `system_settings` | 시스템 설정 |
| `audit_logs` | 감사 로그 |

---

## 📖 API 문서

### 인증

모든 API는 인증이 필요합니다 (`/api/v1/auth/login` 제외):

```bash
# JWT 토큰
Authorization: Bearer <jwt_token>

# API 토큰
Authorization: Bearer ng_<api_token>
```

### 주요 엔드포인트

#### 인증

| 메서드 | 엔드포인트 | 설명 |
|--------|-----------|------|
| POST | `/api/v1/auth/login` | 로그인 |
| POST | `/api/v1/auth/logout` | 로그아웃 |
| GET | `/api/v1/auth/me` | 현재 사용자 |
| POST | `/api/v1/auth/2fa/setup` | 2FA 설정 |

#### 프록시 호스트

| 메서드 | 엔드포인트 | 설명 |
|--------|-----------|------|
| GET | `/api/v1/proxy-hosts` | 목록 조회 |
| POST | `/api/v1/proxy-hosts` | 생성 |
| GET | `/api/v1/proxy-hosts/:id` | 상세 조회 |
| PUT | `/api/v1/proxy-hosts/:id` | 수정 |
| DELETE | `/api/v1/proxy-hosts/:id` | 삭제 |
| POST | `/api/v1/proxy-hosts/:id/test` | 테스트 |

#### 보안 설정 (호스트별)

| 메서드 | 엔드포인트 | 설명 |
|--------|-----------|------|
| GET/PUT | `/api/v1/proxy-hosts/:id/rate-limit` | Rate Limiting |
| GET/PUT | `/api/v1/proxy-hosts/:id/fail2ban` | Fail2ban |
| GET/PUT | `/api/v1/proxy-hosts/:id/bot-filter` | 봇 필터 |
| GET/PUT | `/api/v1/proxy-hosts/:id/waf` | WAF 설정 |
| GET/PUT | `/api/v1/proxy-hosts/:id/geo` | 지역 제한 |
| GET/PUT | `/api/v1/proxy-hosts/:id/uri-block` | URI 차단 |

#### 전역 설정

| 메서드 | 엔드포인트 | 설명 |
|--------|-----------|------|
| GET/PUT | `/api/v1/settings` | 전역 Nginx 설정 |
| GET/PUT | `/api/v1/system-settings` | 시스템 설정 |
| GET/PUT | `/api/v1/security/global-banned-ips` | 전역 IP 차단 |
| GET/PUT | `/api/v1/security/global-uri-block` | 전역 URI 차단 |

#### 로그 & 모니터링

| 메서드 | 엔드포인트 | 설명 |
|--------|-----------|------|
| GET | `/api/v1/dashboard` | 대시보드 데이터 |
| GET | `/api/v1/dashboard/geoip` | GeoIP 통계 |
| GET | `/api/v1/logs` | 로그 조회 |
| GET | `/api/v1/logs/stats` | 로그 통계 |
| GET | `/api/v1/audit-logs` | 감사 로그 |

---

## 💻 개발

### 로컬 개발 환경

#### 백엔드

```bash
cd api
go mod download
go run ./cmd/server
```

#### 프론트엔드

```bash
cd ui
npm install
npm run dev
```

### Docker 개발 환경

```bash
# 개발 모드 실행 (핫 리로딩)
docker compose -f docker-compose.dev.yml up -d --build
```

### 프로젝트 구조

```
nginx-proxy-guard/
├── api/                    # Go 백엔드
│   ├── cmd/server/         # 메인 엔트리포인트
│   ├── internal/
│   │   ├── handler/        # HTTP 핸들러
│   │   ├── model/          # 데이터 모델
│   │   ├── repository/     # DB 접근 레이어
│   │   ├── service/        # 비즈니스 로직
│   │   ├── database/       # DB 마이그레이션
│   │   └── nginx/          # Nginx 설정 생성
│   └── pkg/                # 공유 패키지
├── ui/                     # React 프론트엔드
│   ├── src/
│   │   ├── components/     # React 컴포넌트
│   │   ├── api/            # API 클라이언트
│   │   ├── hooks/          # 커스텀 훅
│   │   ├── i18n/           # 다국어 리소스
│   │   └── types/          # TypeScript 타입
│   └── public/
├── nginx/                  # Nginx 설정
│   ├── nginx.conf          # 메인 설정
│   ├── modsec/             # ModSecurity 설정
│   ├── includes/           # 공통 설정
│   └── scripts/            # 헬퍼 스크립트
└── docker-compose.yml      # Docker 구성
```

---

## 🔒 보안 고려사항

### 필수 조치

1. **기본 비밀번호 변경**: 설치 후 즉시 `admin` 비밀번호 변경
2. **강력한 시크릿 사용**: `JWT_SECRET`, `DB_PASSWORD`는 32자 이상 랜덤 문자열
3. **2FA 활성화**: 모든 관리자 계정에 TOTP 설정
4. **API 토큰 관리**: IP 제한 및 만료일 설정

### 권장 조치

5. **네트워크 분리**: 관리 포트(81)는 내부 네트워크에서만 접근
6. **정기 백업**: 자동 백업 설정 및 복원 테스트
7. **로그 모니터링**: 감사 로그 정기 검토
8. **업데이트 유지**: 보안 패치 적용을 위한 정기 업데이트

### WAF 운영

- **탐지 모드 우선**: 새 호스트는 Detection 모드로 시작
- **점진적 강화**: 오탐 확인 후 Blocking 모드 전환
- **룰 예외 관리**: 필요한 경우만 룰 비활성화

---

## 🔧 문제 해결

### 일반적인 문제

<details>
<summary><b>인증서 발급 실패</b></summary>

1. DNS 설정 확인 (도메인이 서버 IP를 가리키는지)
2. DNS Provider API 토큰 권한 확인
3. Let's Encrypt Rate Limit 확인
4. 스테이징 환경에서 먼저 테스트

```bash
# 인증서 로그 확인
docker logs nginx-proxy-guard-api | grep -i cert
```
</details>

<details>
<summary><b>WAF 오탐 (False Positive)</b></summary>

1. WAF 테스트 기능으로 어떤 룰이 트리거되는지 확인
2. 해당 룰 ID를 호스트별 예외 목록에 추가
3. 탐지 모드로 변경하여 모니터링 후 차단 모드 적용
</details>

<details>
<summary><b>GeoIP 작동 안함</b></summary>

1. MaxMind 라이선스 키 확인
2. GeoIP 데이터베이스 존재 확인

```bash
docker exec nginx-proxy-guard-proxy ls -la /etc/nginx/geoip/
```

3. Nginx 재시작
```bash
docker compose restart nginx
```
</details>

<details>
<summary><b>로그 조회가 느림</b></summary>

1. 로그 보존 기간 단축 (설정 > 시스템 설정)
2. 필터 범위 축소
3. 파티션 상태 확인

```sql
-- 파티션 확인
SELECT * FROM pg_partitions WHERE tablename = 'logs_partitioned';
```
</details>

---

## 🗺️ 로드맵

### 진행 중

- [ ] 클러스터 모드 지원
- [ ] Prometheus 메트릭 내보내기
- [ ] 추가 언어 지원

### 계획됨

- [ ] Kubernetes Helm 차트
- [ ] API Rate Limiting (토큰별)
- [ ] 이메일 알림 시스템
- [ ] 사용자 역할 관리

---

## 🤝 기여

기여를 환영합니다! 다음 가이드라인을 참조해주세요:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### 개발 가이드라인

- Go 코드는 `gofmt` 적용
- React 컴포넌트는 TypeScript 사용
- 커밋 메시지는 [Conventional Commits](https://www.conventionalcommits.org/) 준수
- PR 전 테스트 통과 확인

---

## 📄 라이선스

이 프로젝트는 MIT 라이선스에 따라 배포됩니다. 자세한 내용은 [LICENSE](LICENSE) 파일을 참조하세요.

---

## 📞 지원

- **GitHub Issues**: 버그 리포트 및 기능 요청
- **Discussions**: 질문 및 토론

---

<div align="center">

**[⬆ 맨 위로](#nginx-proxy-guard)**

Made with ❤️ by the Nginx Proxy Guard Team

</div>
