<?php

use Illuminate\Http\Request;
use Illuminate\Support\Facades\Route;

Route::get('/user', function (Request $request) {
    return $request->user();
})->middleware('auth:sanctum');

Route::get('/ws-token', function (Request $request) {
    $channel = $request->query('channel', 'chat');
    $userId = (string) $request->user()->id;
    $timestamp = (string) time();
    $signature = hash_hmac('sha256', $userId.':'.$timestamp.':'.$channel, env('WS_SECRET'));

    return response()->json([
        'token' => $userId.':'.$timestamp.':'.$channel.':'.$signature,
        'channel' => $channel,
    ]);
})->middleware('auth:sanctum');

Route::get('/ws-token/guest', function (Request $request) {
    $channel = $request->query('channel', 'chat');

    if (str_starts_with($channel, 'private-')) {
        return response()->json(['error' => 'guests cannot access private channels'], 403);
    }

    $userId = 'guest';
    $timestamp = (string) time();
    $signature = hash_hmac('sha256', $userId.':'.$timestamp.':'.$channel, env('WS_SECRET'));

    return response()->json([
        'token' => $userId.':'.$timestamp.':'.$channel.':'.$signature,
        'channel' => $channel,
    ]);
})->middleware('throttle:60,1');
