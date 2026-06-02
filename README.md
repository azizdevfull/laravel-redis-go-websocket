# Redis Broadcasting + Go WebSocket Server

Laravel eventi brauzerga real-time yetib borishi uchun to'liq zanjir:
**Laravel → Redis → Go → Brauzer**

---

## Arxitektura

```
┌─────────────┐        ┌───────────┐        ┌────────────────┐        ┌─────────────┐
│   Laravel   │─publish▶│   Redis   │◀─sub───│   Go Server    │─WebSocket▶│   Brauzer   │
│  (PHP app)  │        │  pub/sub  │        │  (ws-server/)  │        │             │
└─────────────┘        └───────────┘        └────────────────┘        └─────────────┘
       │                                            ▲
       │ GET /ws-token                              │ token tekshiradi
       ◀────────────────────────────────────────────┘
```

### Nima uchun Go server kerak?

Brauzer Redis bilan to'g'ridan-to'g'ri gaplasha olmaydi — Redis o'z protokolida ishlaydi, brauzer faqat WebSocket/HTTP tushunadi. Go server ikki tomon orasida ko'prik vazifasini bajaradi:
- Redis kanallariga subscribe bo'lib turadi
- Ulanib turgan brauzerlarga WebSocket orqali xabar yuboradi

---

## Xavfsizlik: HMAC Token

Har qanday odam WebSocket serveriga ulana olmasligi uchun token tizimi ishlatiladi.

### Token formati

```
userId:timestamp:channel:HMAC_SHA256(userId:timestamp:channel, WS_SECRET)

Misol:
  userId    = 42
  timestamp = 1780378499
  channel   = chat
  secret    = "my-strong-secret"
  signature = HMAC_SHA256("42:1780378499:chat", "my-strong-secret")

  token = "42:1780378499:chat:b308f6a7..."
```

Token **channelga bog'langan** — `chat` uchun olingan token `private-chat` ga ishlamaydi.

### Token qanday tekshiriladi (Go)

1. Tokenni `:` bo'yicha 4 qismga ajratadi
2. `timestamp` ni tekshiradi — 60 soniyadan eski bo'lsa → **401**
3. Tokendagi `channel` so'ralgan channel bilan solishtiriladi — mos kelmasa → **401**
4. HMAC ni qayta hisoblaydi — mos kelmasa → **401**
5. Hammasi to'g'ri → WebSocket ulanishiga ruxsat

### Nima uchun oddiy parol emas?

Oddiy parol hech qachon o'zgarmaydi — birov URL ni ko'rsa, abadiy kirish imkoni bo'ladi.

HMAC token **60 soniyada eskiradi** va **faqat bitta channelga** ishlaydi.

---

## O'rnatish

### Talablar

- PHP 8.4+
- Composer
- Node.js & npm
- Go 1.21+
- Redis
- `php8.4-redis` extension

### 1. Reponi klonlash

```bash
git clone <repo-url>
cd redis-broadcasting
```

### 2. PHP dependency'larni o'rnatish

```bash
composer install
```

### 3. Node dependency'larini o'rnatish

```bash
npm install
```

### 4. `.env` sozlash

```bash
cp .env.example .env
php artisan key:generate
```

`.env` ichida quyidagilarni to'g'irlang:

```env
BROADCAST_CONNECTION=redis

REDIS_CLIENT=phpredis
REDIS_HOST=127.0.0.1
REDIS_PASSWORD=null
REDIS_PORT=6379
REDIS_PREFIX=broadcast_redis_

# Go server bilan umumiy sir — ikki tomonda bir xil bo'lishi shart
WS_SECRET=your-strong-random-secret
```

### 5. `phpredis` extension o'rnatish

```bash
sudo apt-get install php8.4-redis
php -m | grep redis   # "redis" chiqishi kerak
```

### 6. Ma'lumotlar bazasini tayyorlash

```bash
php artisan migrate
```

### 7. Go server dependency'larini o'rnatish

```bash
cd ws-server
cp .env.example .env
# .env ichida WS_SECRET ni Laravel bilan bir xil qiling
go mod download
```

---

## Ishga tushirish

### 1. Redis

```bash
redis-cli ping   # PONG chiqishi kerak
```

### 2. Laravel

```bash
php artisan serve
# yoki to'liq dev muhit:
composer run dev
```

### 3. Go WebSocket server

```bash
cd ws-server
WS_SECRET=your-strong-random-secret \
REDIS_PREFIX=broadcast_redis_ \
go run .
```

Muvaffaqiyatli ishga tushsa:
```
redis connected
WebSocket server listening on :8080
redis subscriber ready  pattern=broadcast_redis_*
```

---

## Token endpointlari

| Endpoint | Auth | Kimga | Channel cheklov |
|---|---|---|---|
| `GET /api/ws-token?channel=chat` | Sanctum token | Login qilgan user | Istalgan channel |
| `GET /api/ws-token/guest?channel=chat` | Yo'q (60 req/min) | Ochiq sahifalar | `private-*` dan tashqari |

Guest `private-*` channel so'rasa **403** qaytadi.

---

## To'liq oqim (qadamma-qadam)

### 1. Brauzer token oladi

**Login qilgan user:**
```
GET /api/ws-token?channel=chat
Authorization: Bearer <sanctum-token>
```

**Guest (ochiq sahifa):**
```
GET /api/ws-token/guest?channel=public-chat
```

Laravel javobi (ikkalasida bir xil format):
```json
{
  "token": "42:1780378499:chat:b308f6a7...",
  "channel": "chat"
}
```

### 2. Brauzer WebSocket ulanishini ochadi

```javascript
const { token, channel } = await fetch('/api/ws-token?channel=chat', {
    headers: { 'Authorization': 'Bearer ' + sanctumToken }
}).then(r => r.json());

const ws = new WebSocket(`ws://localhost:8080/ws?token=${token}&channel=${channel}`);

ws.onmessage = (e) => {
    const data = JSON.parse(e.data);
    console.log(data.event);  // "App\Events\MessageSent"
    console.log(data.data);   // { message: "Salom!" }
};
```

### 3. Laravel event yuboradi

```php
broadcast(new MessageSent('Salom!'));
// yoki
MessageSent::dispatch('Salom!');
```

### 4. Brauzer xabarni oladi

```json
{
  "event": "App\\Events\\MessageSent",
  "data": {
    "message": "Salom!",
    "socket": null
  }
}
```

---

## Redis kanal nomi

Laravel `REDIS_PREFIX` + channel nomini birlashtiradi:

```
REDIS_PREFIX=broadcast_redis_
channel=chat

→ Redis kanal: "broadcast_redis_chat"
```

Go server ham xuddi shu prefixni ishlatadi — `broadcast_redis_*` patterniga subscribe bo'ladi.

---

## Loyiha tuzilmasi

```
├── app/Events/
│   └── MessageSent.php        — ShouldBroadcastNow event
├── config/
│   └── broadcasting.php       — redis driver sozlamasi
├── routes/
│   ├── web.php                — / (demo) va /ws-token endpoint
│   └── channels.php           — broadcast channel auth
├── ws-server/
│   ├── main.go                — HTTP/WebSocket server, Redis subscriber
│   ├── hub.go                 — ulanishlarni channel bo'yicha boshqaradi
│   ├── auth.go                — HMAC token validatsiyasi
│   ├── go.mod
│   └── .env.example
└── resources/js/
    └── echo.js                — Laravel Echo sozlamasi
```

---

## v2 — Production yaxshilanishlari

`v2` branch da Go server to'liq production tayyor qilindi:

| | v1 | v2 |
|---|---|---|
| Write | Bloklaydi | Har clientga `chan []byte` buffer (256) |
| Sekin client | Hamma kutadi | Avtomatik drop qilinadi |
| Ping/Pong | Yo'q | 30s ping, 60s pong timeout |
| Connection limit | Yo'q | `MAX_CONNECTIONS` env (default 10000) |
| Graceful shutdown | Yo'q | SIGTERM → 10s ichida yopadi |
| Metrics | Yo'q | `GET /metrics` — aktiv ulanishlar soni |

### Metrics

```
GET http://localhost:8080/metrics
```

```
ws_connections_active 42
ws_connections_by_channel 42
```

### `MAX_CONNECTIONS` sozlash

```env
MAX_CONNECTIONS=50000
```

---

## Production uchun eslatmalar

- `WS_SECRET` ni kamida 32 belgili tasodifiy qator qiling
- `ALLOWED_ORIGINS` ni aniq domenga sozlang (hozir `http://localhost:8000`)
- Go serverni `systemd` yoki `supervisor` orqali ishga tushiring
- WebSocket uchun `wss://` (TLS) ishlatish tavsiya etiladi
- Horizontal scale: bir nechta Go instance, hammasi bir Redis ga — ishlaydi

### Nginx + subdomain sozlash

`ws.loyiha.uz` kabi subdomain orqali portni yashirish va SSL qo'shish:

```nginx
upstream ws_backend {
    server localhost:8080;
    server localhost:8081;
    server localhost:8082;
}

server {
    listen 80;
    server_name ws.loyiha.uz;

    location / {
        proxy_pass http://ws_backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

SSL sertifikat (Let's Encrypt):

```bash
certbot --nginx -d ws.loyiha.uz
```

Brauzerda port ko'rsatish shart emas:

```javascript
const ws = new WebSocket('wss://ws.loyiha.uz/ws?token=...&channel=...');
```

> **Muhim:** `proxy_http_version 1.1` va `Upgrade` headerlari bo'lmasa WebSocket ishlamaydi — oddiy HTTP bo'lib qoladi.

### Horizontal scaling

Har Go instance bir Redis ga subscribe bo'ladi — message hammaga keladi:

```
                ┌─ Go instance 1 (:8080) ─┐
                │                          │
Redis pub/sub ──┼─ Go instance 2 (:8081) ─┼──── Brauzerlar
                │                          │
                └─ Go instance 3 (:8082) ─┘
```

```bash
PORT=8080 go run . &
PORT=8081 go run . &
PORT=8082 go run . &
```

Har instance ~10k ulanish → 3 instance = ~30k. Kerak bo'lsa yana qo'shiladi.

---

## Litsenziya

MIT
