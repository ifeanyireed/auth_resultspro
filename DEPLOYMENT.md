# Deployment Guide: Central Auth Service (`auth_service`)

This document outlines the steps to deploy the Central Auth service to a production environment using MySQL.

---

## 🏗 Prerequisites

- **Server**: DigitalOcean Droplet (Ubuntu 24.04 recommended).
- **Database**: MySQL (Hostinger/Remote).
- **Domain**: `auth.resultspro.ng` pointed to the Droplet IP.

---

## 1. Database Setup (MySQL)

1.  **Schema Import**: Import `schema.sql` into your MySQL database `u560700323_auth_servicedb`.
2.  **Remote Access**: Ensure the Droplet IP is whitelisted in your Hostinger remote access settings.
3.  **Seeding**: From your local machine, run the seeder:
    ```bash
    go run seed/main.go
    ```

---

## 2. Local Build & Upload

1.  **Compile for Linux**:
    ```bash
    GOOS=linux GOARCH=amd64 go build -o auth_service main.go
    ```
2.  **Upload to Droplet**:
    ```bash
    scp auth_service .env root@167.99.15.196:/var/www/auth_resultspro/
    ```

---

## 3. Systemd Configuration

Create `/etc/systemd/system/auth.service`:

```ini
[Unit]
Description=Central Auth Service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/var/www/auth_resultspro
ExecStart=/var/www/auth_resultspro/auth_service
Restart=always
RestartSec=5
EnvironmentFile=/var/www/auth_resultspro/.env

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
systemctl daemon-reload
systemctl enable auth
systemctl start auth
```

---

## 4. Nginx Reverse Proxy & SSL

```nginx
server {
    listen 80;
    server_name auth.resultspro.ng;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

Secure with Certbot:
```bash
certbot --nginx -d auth.resultspro.ng
```
