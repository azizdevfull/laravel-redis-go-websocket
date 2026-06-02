# Redis Broadcasting with Laravel

Laravel orqali Redis pub/sub yordamida real-time broadcasting. Go WebSocket server ko'prik vazifasini bajaradi — Redis va brauzer orasida.

## Arxitektura

```
Laravel Event
    │
    ▼ (broadcast)
Redis pub/sub  ←─── BROADCAST_CONNECTION=redis
    │
    ▼ (subscribe)
Go WebSocket Server  ←─── siz yozasiz
    │
    ▼ (WebSocket)
Brauzer (Laravel Echo)
```

### Nima uchun Go server kerak?

Brauzer Redis protokolini tushunmaydi — faqat WebSocket/HTTP. Go server ikki protokol orasida ko'prik bo'lib ishlaydi: Redis kanallariga subscribe bo'ladi va brauzerga WebSocket orqali yuboradi.

---

## Talablar

- PHP 8.4+
- Composer
- Node.js & npm
- Redis (lokal yoki remote)
- `php8.4-redis` extension

## O'rnatish

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

`.env` fayliga quyidagilarni to'g'irlang:

```env
BROADCAST_CONNECTION=redis

REDIS_CLIENT=phpredis
REDIS_HOST=127.0.0.1
REDIS_PASSWORD=null
REDIS_PORT=6379
```

### 5. `phpredis` extension o'rnatish

```bash
sudo apt-get install php8.4-redis
```

Tekshirish:

```bash
php -m | grep redis
```

### 6. Ma'lumotlar bazasini tayyorlash

```bash
php artisan migrate
```

### 7. Frontend build

```bash
npm run build
```

---

## Ishga tushirish

### Redis ishlaётganini tekshirish

```bash
redis-cli ping
# PONG
```

### Laravel serverni ishga tushirish

```bash
php artisan serve
```

yoki to'liq dev muhit (server + queue + logs + vite):

```bash
composer run dev
```

---

## Qanday ishlaydi

### Event

`app/Events/MessageSent.php` — `ShouldBroadcastNow` implement qiladi, ya'ni queue'siz darhol Redis ga publish bo'ladi.

```php
class MessageSent implements ShouldBroadcastNow
{
    public function __construct(public string $message) {}

    public function broadcastOn(): array
    {
        return [new Channel('chat')];
    }
}
```

### Redis kanal nomi

Laravel Redis prefix qo'shadi: `{APP_NAME}-database-{channel}`

```
laravel-database-chat
```

### Event dispatch qilish

```php
// Yerda dispatch
broadcast(new MessageSent('Salom!'));

// yoki
MessageSent::dispatch('Salom!');
```

Demo uchun `GET /` route event dispatch qiladi.

### Redis da tekshirish

```bash
# Terminal 1 — subscribe
redis-cli subscribe laravel-database-chat

# Terminal 2 — event trigger
curl http://localhost:8000
```

Terminal 1 da shu ko'rinishdagi natija keladi:

```json
{
  "event": "App\\Events\\MessageSent",
  "data": {
    "message": "Hello, world!",
    "socket": null
  },
  "socket": null
}
```

---

## Go WebSocket Server

Go server quyidagi vazifalarni bajaradi:

1. `laravel-database-{channel}` ga Redis subscribe
2. WebSocket orqali brauzer ulanishlarini qabul qilish
3. Redis dan kelgan xabarni barcha ulangan brauzerlarga yuborish

### Kerakli paketlar

```
go-redis/redis/v9      — Redis client
gorilla/websocket      — WebSocket server
```

---

## Frontend (Laravel Echo)

`resources/js/echo.js` — Laravel Echo sozlamasi. Hozirda `reverb` broadcaster ishlatilgan. Go server tayyor bo'lgach `broadcaster: 'reverb'` o'rniga `broadcaster: 'socket.io'` yoki oddiy WebSocket ga o'zgartirish kerak.

```js
window.Echo.channel('chat').listen('MessageSent', (e) => {
    console.log(e.message);
});
```

---

## Loyiha tuzilmasi

```
app/Events/
    MessageSent.php       — Broadcast event
config/
    broadcasting.php      — Redis connection sozlamasi
    reverb.php            — Reverb config (zaxira)
routes/
    channels.php          — Broadcast channel auth
    web.php               — Demo route
resources/js/
    echo.js               — Laravel Echo setup
    app.js                — JS entry point
```

---

## Litsenziya

MIT
