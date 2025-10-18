# apoteker-bot

A whatsapp chatbot that provide doctor auto send message to pharmacy and auto send message about queue to customer

## Features
- Auto send message from doctor to pharmacy
- Auto send message to customer about the queue
- Getting Sheets note link for doctor

## Tech stack
This project using Go language and using some package as a core of the apps like:
- fiber V2
- GOWA: https://github.com/aldinokemal/go-whatsapp-web-multidevice
- sheets

## Installation
1. Clone the repo:
    git clone https://github.com/haikalnuril/apoteker-bot
3. Create Credential files:
    - clone from .env.example
    - create bot-credentials.json (you can get this file from google cloud console)
2. Install dependencies:
    go mod tidy

## Configuration
Create a `.env` file in the project root:
```
WHATSAPP_WEBHOOK_URL=
GOWA_USERNAME=
GOWA_PASSWORD=
GOWA_PORT=

APP_PORT=
WHATSAPP_API_URL=
ALLOWED_NUMBER=
EXCEL_OUTPUT_PATH=
SHEET_LINK=
PHARMACY_NUMBER=
SHEET_ID=
```
Get your credentials sheet from google cloud console

## Run
- Development (with auto-reload if you use nodemon):
  go run cmd/app/main.go
  ngrok http 8080
  docker-compose up -d --build

## Troubleshooting
- If bot doesn't start, check GOWA configuration its connected to your whatsapp or not.
- Inspect logs for stack traces and missing environment variables.

## Contributing
- Fork, create a feature branch, add tests, and open a pull request.
- Follow repository linting and commit message guidelines.

## License
Add your preferred license (e.g., MIT) in LICENSE file.

If you want, provide the project's actual package.json, main source file, or list of installed libraries and I will update this README to match exactly.