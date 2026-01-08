<h1 align="center" style="border-bottom: none">
<img alt="Sniffly" src="/docs/ico.gif"><br>Sniffly
</h1>
<p align="center">
<a href="https://github.com/nrf24l01/sniffly/actions/workflows/frontend.yml"><img src="https://github.com/nrf24l01/sniffly/actions/workflows/frontend.yml/badge.svg" alt="Frontend build"/></a>
<a href="https://github.com/nrf24l01/sniffly/actions/workflows/capturer.yml"><img src="https://github.com/nrf24l01/sniffly/actions/workflows/capturer.yml/badge.svg" alt="Capturer build"/></a>
<a href="https://github.com/nrf24l01/sniffly/actions/workflows/capture_receiver.yml"><img src="https://github.com/nrf24l01/sniffly/actions/workflows/capture_receiver.yml/badge.svg" alt="capture_receiver build"/></a>
<a href="https://github.com/nrf24l01/sniffly/actions/workflows/backend.yml"><img src="https://github.com/nrf24l01/sniffly/actions/workflows/backend.yml/badge.svg" alt="Backend build"/></a>
<a href="https://github.com/nrf24l01/sniffly/actions/workflows/analyzer.yml"><img src="https://github.com/nrf24l01/sniffly/actions/workflows/analyzer.yml/badge.svg" alt="Analyzer build"/></a>
</p>

## About Sniffly

Sniffly is a lightweight tool for collecting and analyzing what IoT devices do on your network. It captures device network activity and translates it into human‑readable insights—what servers and domains devices contact, which countries those servers are in, connection protocols, data volumes.

Key features
- Per‑device network dashboards (requests, bytes, protocols, geo)
- Server/domain and country mapping for outbound connections
- Passive capturer agents for wide platform support
- Central ingestion, backend API, and analysis pipeline
- Easy deployment via Docker Compose; token‑based device auth
- Useful for security auditing, privacy analysis, and device inventory

Use Sniffly to quickly understand and monitor the real-world behavior of IoT devices on your network.

## Architecture
![arch](/docs/arch.png)

## How to run
### Server-side
- Pull images
  ```bash
  docker compose pull
  ```
- Create user
  Run
  ```bash
  docker compose run backend ./main create_user <username>
  ```
  And enter password
- Run compose
  ```bash
  docker compose up -d
  ```
### Capturer-side
- Download capturer binary
  ```bash
  ARCH="$(uname -m)"
  wget "https://github.com/nrf24l01/sniffly/releases/download/latest/capturer-$(
    case "$ARCH" in
      x86_64) echo amd64 ;;
      i386|i686) echo x86 ;;
      aarch64) echo arm64 ;;
      armv7l) echo armv7 ;;
      armv6l) echo armv6 ;;
      *) echo "unsupported-arch-$ARCH" >&2; exit 1 ;;
    esac
  )"
  ```
- Create .env and fill it
  ```bash
  SERVER_ADDRESS=<ip>:<port>
  API_TOKEN=<your api token from web ui>
  INTERFACE=<interface>
  ```
- Give permissions
  ```bash
  ARCH="$(uname -m)"
  case "$ARCH" in
    x86_64) F=amd64 ;;
    i386|i686) F=x86 ;;
    aarch64) F=arm64 ;;
    armv7l) F=armv7 ;;
    armv6l) F=armv6 ;;
    *) echo "unsupported-arch-$ARCH" >&2; exit 1 ;;
  esac
  FILE="capturer-$F"
  sudo setcap cap_net_raw+ep "$FILE"
  chmod +x "$FILE"
  ```
- Run capturer
  ```bash
  ./"$FILE"
  ```

## Some things
- Presentation - [click](https://docs.google.com/presentation/d/1BIs7U2hdOIE7XOnk9SHtjRfNMy3rvBSwfH_0rmnYHYA/edit?usp=sharing)
