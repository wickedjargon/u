# u

A personal, lightweight file upload service written in Go.

## Usage

Upload a file:
```bash
curl -X POST -H "Authorization: Bearer $TOKEN" -F "file=@photo.jpg" https://u.fftp.io/upload
```

Preserve original filename:
```bash
curl -X POST -H "Authorization: Bearer $TOKEN" -F "file=@photo.jpg" -F "preserve_filename=true" https://u.fftp.io/upload
```

Returns the URL to the uploaded file.

## Setup

### Prerequisites
- Go 1.21+
- [Nginx](https://nginx.org/) + [Certbot](https://certbot.eff.org/) (for reverse proxy + TLS)

### Deploy
```bash
# Generate an API key
openssl rand -hex 32

# Create env file on your server
echo 'UPLOAD_API_KEY=your-generated-key' | sudo tee /etc/u.env
sudo chmod 600 /etc/u.env

# Deploy
./deploy.sh root@your-server
```

### Run locally
```bash
UPLOAD_API_KEY=testkey UPLOAD_DIR=/tmp/uploads BASE_URL=http://localhost:8080 go run .
```

## Configuration

All config is via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |
| `UPLOAD_DIR` | `/var/www/uploads` | Where files are stored |
| `BASE_URL` | `https://u.fftp.io` | URL prefix for returned links |
| `UPLOAD_API_KEY` | *(required)* | Bearer token for auth |

## License

MIT