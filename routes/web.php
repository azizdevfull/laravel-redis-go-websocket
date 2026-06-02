<?php

use App\Events\MessageSent;
use Illuminate\Support\Facades\Route;

Route::get('/', function () {
    MessageSent::dispatch('Hello, world!');
    return view('welcome');
});
