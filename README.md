# Sniffly
*Your IoT devices traffic watcher*

## Docs
![arch](/docs/arch.png)

## CI/CD
Automated builds, tests and analysis for every component — PR validation and main-branch releases.

<p align="center">
<a href="https://github.com/nrf24l01/sniffly/actions/workflows/frontend.yml"><img src="https://github.com/nrf24l01/sniffly/actions/workflows/frontend.yml/badge.svg" alt="Frontend build"/></a>
<a href="https://github.com/nrf24l01/sniffly/actions/workflows/capturer.yml"><img src="https://github.com/nrf24l01/sniffly/actions/workflows/capturer.yml/badge.svg" alt="Capturer build"/></a>
<a href="https://github.com/nrf24l01/sniffly/actions/workflows/capture_receiver.yml"><img src="https://github.com/nrf24l01/sniffly/actions/workflows/capture_receiver.yml/badge.svg" alt="capture_receiver build"/></a>
<a href="https://github.com/nrf24l01/sniffly/actions/workflows/backend.yml"><img src="https://github.com/nrf24l01/sniffly/actions/workflows/backend.yml/badge.svg" alt="Backend build"/></a>
<a href="https://github.com/nrf24l01/sniffly/actions/workflows/analyzer.yml"><img src="https://github.com/nrf24l01/sniffly/actions/workflows/analyzer.yml/badge.svg" alt="Analyzer build"/></a>
</p>

### Components
- Frontend — web UI (build & test)
- Capturer — packet capture agent (build & unit tests)
- capture_receiver — capture ingestion service (build & tests)
- Backend — API & data services (build, tests, migrations)
- Analyzer — traffic analysis pipeline (build & pipeline tests)

## Some things
- Presentation - [click](https://docs.google.com/presentation/d/1BIs7U2hdOIE7XOnk9SHtjRfNMy3rvBSwfH_0rmnYHYA/edit?usp=sharing)

## Typical Dashboard per device, *n - configurable time range*:
- Speeds up and down last *n time*
- Requests rate per ip last *n time*
- Requests rate per domain last *n time*
- Requests rate per proto last *n time*
- Requests rate per country *n time*
- Total send bytes per *n time*