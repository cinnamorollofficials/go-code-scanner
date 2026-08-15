# Panduan Deployment Website

Dokumen ini adalah panduan internal berbahasa Indonesia untuk operator yang
menjalankan website dokumentasi Go Code Scanner. Dokumentasi publik di
`website/docs/` tetap menggunakan bahasa Inggris.

Jalur yang didukung untuk VPS adalah Docker Compose dengan website disajikan di
root domain, misalnya `https://docs.example.com/`. Jalankan `npm run docs:verify`
di lingkungan pengembangan atau CI sebelum membuat image deployment.

## Berkas yang digunakan

- [Dockerfile](./Dockerfile) membangun aset statis dengan Node.js 20, lalu
  menyalinnya ke image Nginx 1.27.
- [docker-compose.yml](./docker-compose.yml) membangun service `website`,
  mengaktifkan restart otomatis, dan memetakan port host `80` ke container.
- [nginx.conf](./nginx.conf) menangani clean URL, cache aset, kompresi, header
  keamanan, dan halaman 404.
- [.dockerignore](./.dockerignore) menjaga konteks build tetap kecil.

## Prasyarat

Host deployment memerlukan Git, Docker Engine, Docker Compose v2, dan `curl`
untuk pemeriksaan akhir.

```sh
git --version
docker --version
docker compose version
curl --version
```

## Jalur deployment yang didukung

Clone repository dan jalankan Compose dari folder `website/`:

```sh
git clone https://github.com/cinnamorollofficials/go-code-scanner.git
cd go-code-scanner/website
docker compose config
docker compose up -d --build
docker compose ps
curl --fail --silent --show-error http://127.0.0.1/
```

Compose menetapkan `VITEPRESS_BASE=/`, sesuai dengan Nginx yang menyajikan hasil
build di root. Setelah pemeriksaan `curl` berhasil, website tersedia melalui
alamat IP atau domain host pada port 80.

## Memperbarui deployment

Ambil perubahan secara fast-forward, bangun ulang image, lalu periksa endpoint:

```sh
cd go-code-scanner/website
git pull --ff-only origin main
docker compose up -d --build --remove-orphans
docker compose ps
curl --fail --silent --show-error http://127.0.0.1/
```

Jika container tidak sehat atau endpoint gagal, lihat status dan log sebelum
mengubah konfigurasi:

```sh
docker compose ps
docker compose logs --tail=200 website
```

## Base path dan subpath

Konfigurasi Compose yang disertakan mendukung root path `/`. Untuk memeriksa
hasil build statis yang menargetkan subpath, gunakan nilai dengan slash pembuka
dan penutup:

```sh
VITEPRESS_BASE=/docs/ npm run docs:build-site
```

Perintah itu menghasilkan URL aset dengan prefix `/docs/`, tetapi image Nginx
yang disertakan tidak langsung melayani subpath. Web server atau reverse proxy
harus memetakan `/docs/` ke isi `docs/.vitepress/dist/`. Jangan hanya mengganti
build argument pada image root-path karena halaman akan meminta aset dari lokasi
yang tidak disajikan oleh konfigurasi Nginx saat ini.

## Opsi lanjutan

### Docker CLI tanpa Compose

Gunakan jalur ini hanya bila lifecycle container dikelola tanpa Compose:

```sh
cd go-code-scanner/website
docker build -t go-code-scanner-website:latest .
docker run -d \
  --name go-code-scanner-website \
  --restart unless-stopped \
  -p 80:80 \
  go-code-scanner-website:latest
curl --fail --silent --show-error http://127.0.0.1/
```

Hapus container lama sebelum memakai kembali nama yang sama pada deployment
baru.

### Reverse proxy dan TLS pada VPS

Jika host Nginx atau proxy lain mengelola domain dan TLS, ubah port Compose
menjadi loopback-only agar container tidak terekspos langsung:

```yaml
ports:
  - "127.0.0.1:8080:80"
```

Contoh virtual host Nginx di host:

```nginx
server {
    listen 80;
    server_name docs.example.com;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

Validasi dan reload konfigurasi host sebelum meminta sertifikat:

```sh
sudo nginx -t
sudo systemctl reload nginx
sudo certbot --nginx -d docs.example.com
```

Instalasi Nginx host, firewall, DNS, dan Certbot bergantung pada distribusi dan
kebijakan infrastruktur. Langkah tersebut bukan bagian dari image website.

## Operasi rutin

```sh
# Status service
docker compose ps

# Log terbaru; tambahkan --follow bila perlu memantau
docker compose logs --tail=200 website

# Hentikan service tanpa menghapus image
docker compose down
```
