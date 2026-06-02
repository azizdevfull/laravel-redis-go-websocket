<?php

use App\Events\MessageSent;
use Illuminate\Support\Facades\Route;

Route::get('/', function () {
    MessageSent::dispatch('Hello, world!');

    return view('welcome');
});

// Token endpoint — add auth middleware in production: ->middleware('auth')
Route::get('/ws-token', function () {
    $userId = (string) (auth()->id() ?? 'guest');
    $timestamp = (string) time();
    $signature = hash_hmac('sha256', $userId.':'.$timestamp, env('WS_SECRET'));

    return response()->json([
        'token' => $userId.':'.$timestamp.':'.$signature,
        'channel' => 'chat',
    ]);
});
