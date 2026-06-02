<?php

use App\Events\MessageSent;
use App\Models\User;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Auth;
use Illuminate\Support\Facades\Route;
use League\CommonMark\CommonMarkConverter;

Route::get('/', function () {
    MessageSent::dispatch('Hello, world!');

    $markdown = file_get_contents(base_path('README.md'));
    $converter = new CommonMarkConverter(['html_input' => 'strip']);
    $content = $converter->convert($markdown)->getContent();

    return view('welcome', compact('content'));
});
Route::get('/auth', function () {
    Auth::login(User::first()); // Log in the first user for testing purposes

    return auth()->user(); // Return the authenticated user
});

// Token endpoint — add auth middleware in production: ->middleware('auth')
Route::get('/ws-token', function (Request $request) {
    $channel = $request->query('channel', 'chat');
    $userId = (string) $request->user()->id;
    $timestamp = (string) time();
    $signature = hash_hmac('sha256', $userId . ':' . $timestamp . ':' . $channel, env('WS_SECRET'));

    return response()->json([
        'token' => $userId . ':' . $timestamp . ':' . $channel . ':' . $signature,
        'channel' => $channel,
    ]);
})->middleware('auth');