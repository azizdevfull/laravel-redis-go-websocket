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

### Token qanday yaratiladi (Laravel)

```
token = userId + ":" + timestamp + ":" + HMAC_SHA256(userId:timestamp, WS_SECRET)

Misol:
  userId    = 42
  timestamp = 1780378499
  secret    = "my-strong-secret"
  signature = HMAC_SHA256("42:1780378499", "my-strong-secret")

  token = "42:1780378499:dcd1ae409bed76c7..."
```

### Token qanday tekshiriladi (Go)

1. Tokenni `:` bo'yicha 3 qismga ajratadi
2. `timestamp` ni tekshiradi — agar 60 soniyadan eski bo'lsa → **401 Unauthorized**
3. HMAC ni qayta hisoblaydi — agar mos kelmasa → **401 Unauthorized**
4. Hammasi to'g'ri → WebSocket ulanishiga ruxsat

### Nima uchun oddiy parol emas?

Oddiy parol (`?password=secret`) hech qachon o'zgarmaydi — birov URL ni ko'rsa yoki log fayldan o'qisa, abadiy kirish imkoni bo'ladi.

HMAC token esa **60 soniyada eskiradi** — birov tokenni ushlab qolsa ham foydasiz.

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
REDIS_PREFIX=broadcast_redis

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
REDIS_PREFIX=broadcast_redis \
go run .
```

Muvaffaqiyatli ishga tushsa:
```
redis connected
WebSocket server listening on :8080
redis subscriber ready  pattern=broadcast_redis*
```

---

## To'liq oqim (qadamma-qadam)

### 1. Brauzer token oladi

```
GET http://localhost:8000/ws-token
```

Laravel javobi:
```json
{
  "token": "guest:1780378499:dcd1ae409bed76c7...",
  "channel": "chat"
}
```

### 2. Brauzer WebSocket ulanishini ochadi

```javascript
const { token, channel } = await fetch('/ws-token').then(r => r.json());

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
REDIS_PREFIX=broadcast_redis
channel=chat

→ Redis kanal: "broadcast_redischat"
```

Go server ham xuddi shu prefixni ishlatadi — `broadcast_redis*` patterniga subscribe bo'ladi.

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

## Production uchun eslatmalar

- `/ws-token` routega `->middleware('auth')` qo'shing — faqat tizimga kirgan foydalanuvchilar token olsin
- `WS_SECRET` ni kamida 32 belgili tasodifiy qator qiling
- `ALLOWED_ORIGINS` ni aniq domenga sozlang (hozir `http://localhost:8000`)
- Go serverni `systemd` yoki `supervisor` orqali ishga tushiring
- WebSocket uchun `wss://` (TLS) ishlatish tavsiya etiladi

---

## Litsenziya

MIT
