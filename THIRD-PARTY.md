# Attribution

- **3x-ui v3.8.5 / Xray**: runs as a separate official executable downloaded from [MHSanaei/3x-ui](https://github.com/MHSanaei/3x-ui/tree/v3.8.5). 3x-ui is GPL-3.0. Its source, notices and license are available at that tag; the release's `bin/LICENSE` travels with the extracted engine. This project is not endorsed by MHSanaei, Xray or Velixir.
- **QR Code Generator for JavaScript**, Kazuhiko Arase: MIT. Source [qrcode-generator](https://github.com/kazuhikoarase/qrcode-generator). Vendored as `web/qr.js`; license in `web/QR-LICENSE.txt`. QR generation is entirely in the browser.
- **Vazirmatn**, Saber Rastikerdar: SIL Open Font License 1.1. Source [Vazirmatn](https://github.com/rastikerdar/vazirmatn). Local font `web/vazirmatn.woff2`; license in `web/FONT-LICENSE.txt`.

The original launcher's README is retained as `LEGACY-LAUNCHER-README.md` for provenance. The new panel's installation guide is `INSTALL-FA.md`.

- **Caddy 2.11.4**: optional VPS TLS/reverse proxy via official container; Apache-2.0. [Source/license](https://github.com/caddyserver/caddy/tree/v2.11.4).
- **TUIC sidecar**: included in the official 3x-ui bundle; native TUIC implementation is from [tuic-protocol/tuic](https://github.com/tuic-protocol/tuic), GPL-3.0. The panel does not redistribute a modified engine binary in this source ZIP.

## v3 artwork and reference

`web/fire-natural.png` is an original image generated for this project with OpenAI ImageGen; no third-party fire photo is embedded. Prompt direction: realistic irregular orange and gold flame, fine translucent wisps and sparks, isolated transparent background, no text or flat emblem.

The API shape of the developerAmira/telegram-bot Marzban classic connector was inspected and tested against this project. Its implementation is not bundled: https://github.com/developerAmira/telegram-bot . Telegram verification follows https://core.telegram.org/bots/webapps . Classic Marzban routes were checked against https://github.com/Gozargah/Marzban/blob/master/app/routers/user.py .

## v4 functional references
The sales journeys in [x4gpanell/Faoxima](https://github.com/x4gpanell/Faoxima) and [developerAmira/telegram-bot](https://github.com/developerAmira/telegram-bot) were read for functional comparison. The v4 Go gateway, flow, reminder and Telegram document implementations are independent; no PHP or JavaScript bot source from those repositories is vendored in this ZIP. The unmodified developerAmira Marzban connector is used only in the external QA harness, not shipped. Payment API shapes were checked against the providers' primary documentation and samples.
