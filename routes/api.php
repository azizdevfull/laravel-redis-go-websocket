<?php

use Illuminate\Http\Request;
use Illuminate\Support\Facades\Route;

Route::get('/user', function (Request $request) {
    return $request->user();
})->middleware('auth:sanctum');

Route::get('/ws-token', function (Request $request) {
    $userId = (string) $request->user()->id;
    $timestamp = (string) time();
    $signature = hash_hmac('sha256', $userId.':'.$timestamp, env('WS_SECRET'));

    return response()->json([
        'token' => $userId.':'.$timestamp.':'.$signature,
        'channel' => 'chat',
    ]);
})->middleware('auth:sanctum');

Route::get('/ws-token/guest', function (Request $request) {
    $userId = 'guest';
    $timestamp = (string) time();
    $signature = hash_hmac('sha256', $userId.':'.$timestamp, env('WS_SECRET'));

    return response()->json([
        'token' => $userId.':'.$timestamp.':'.$signature,
        'channel' => 'chat',
    ]);
})->middleware('throttle:60,1');
