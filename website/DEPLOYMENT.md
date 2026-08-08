# 🚀 Panduan Deployment Website (VitePress) ke VPS

Dokumen ini berisi panduan lengkap langkah-demi-langkah untuk membuat build dan mendeploy website dokumentasi **Go Code Scanner** ke VPS (Virtual Private Server) menggunakan Docker, Docker Compose, dan Nginx.

---

## 📁 Berkas Deployment yang Disediakan

Semua konfigurasi telah disiapkan di dalam direktori `website/`:

- [Dockerfile](file:///Users/hadiyahku/code/go-code-scanner-1/website/Dockerfile): Multi-stage build (Node.js 20 Alpine untuk build static asset + Nginx 1.27 Alpine untuk web server produksi minimalis).
- [nginx.conf](file:///Users/hadiyahku/code/go-code-scanner-1/website/nginx.conf): Konfigurasi web server Nginx yang dioptimalkan untuk VitePress (`cleanUrls`, routing SPA, kompresi Gzip, caching static assets, & security headers).
- [docker-compose.yml](file:///Users/hadiyahku/code/go-code-scanner-1/website/docker-compose.yml): Berkas Docker Compose untuk menjalankan service web secara praktis.
- [.dockerignore](file:///Users/hadiyahku/code/go-code-scanner-1/website/.dockerignore): Mengabaikan `node_modules` dan file temporary dari konteks build Docker.

---

## 📌 Prasyarat di VPS

Sebelum memulai, pastikan VPS Anda sudah terpasang:
- **Git**
- **Docker** (`docker --version`)
- **Docker Compose** (`docker compose version`)

---

## 🚀 Cara 1: Deploy Menggunakan Docker Compose (Sangat Direkomendasikan)

### Step 1: Clone Repository ke VPS
```bash
git clone https://github.com/cinnamorollofficials/go-code-scanner.git
cd go-code-scanner/website
```

### Step 2: Build & Jalankan Container di Background
```bash
docker compose up -d --build
```

Setelah perintah selesai, website akan langsung dapat diakses melalui browser di IP VPS Anda di port `80` (`http://IP-VPS-ANDA`).

---

## 🐳 Cara 2: Deploy Menggunakan Docker CLI Manual

### Step 1: Build Docker Image
Masuk ke folder `website` dan jalankan perintah build:
```bash
cd website
docker build -t go-code-scanner-website:latest .
```

*Catatan (Custom Path):*
Jika website akan di-deploy ke subpath khusus (misalnya `https://example.com/docs/`), Anda dapat menyertakan `--build-arg VITEPRESS_BASE=/docs/`:
```bash
docker build --build-arg VITEPRESS_BASE=/docs/ -t go-code-scanner-website:latest .
```

### Step 2: Jalankan Container
```bash
docker run -d \
  --name go-code-scanner-website \
  --restart unless-stopped \
  -p 80:80 \
  go-code-scanner-website:latest
```

---

## 🔒 Cara 3: Konfigurasi HTTPS / Domain dengan Nginx Reverse Proxy & SSL (Certbot)

Jika Anda menyambungkan Domain (misalnya `docs.domainanda.com`) dan menggunakan Nginx / Certbot di host VPS:

### 1. Ubah Port pada `docker-compose.yml` agar tidak bentrok dengan Port 80 VPS:
Ubah mapping port ke `127.0.0.1:8080:80`:
```yaml
ports:
  - "127.0.0.1:8080:80"
```

Jalankan ulang docker compose:
```bash
docker compose up -d
```

### 2. Buat Virtual Host Nginx di VPS (`/etc/nginx/sites-available/docs.domainanda.com`)
```nginx
server {
    server_name docs.domainanda.com;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

Aktifkan situs & reload Nginx:
```bash
sudo ln -s /etc/nginx/sites-available/docs.domainanda.com /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

### 3. Pasang Sertifikat SSL gratis dengan Certbot:
```bash
sudo certbot --nginx -d docs.domainanda.com
```

---

## 🔄 Cara Update Website Saat Ada Perubahan Kode

Setiap ada update atau pembaruan dokumentasi di repository:

```bash
cd go-code-scanner/website
git pull origin main
docker compose up -d --build
```

---

## 📊 Monitoring & Command Bantuan

- **Melihat log aplikasi:**
  ```bash
  docker compose logs -f website
  ```
- **Cek status container:**
  ```bash
  docker compose ps
  ```
- **Menghentikan container:**
  ```bash
  docker compose down
  ```
