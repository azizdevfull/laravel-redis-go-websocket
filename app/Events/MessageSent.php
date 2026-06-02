<?php

namespace App\Events;

use Illuminate\Broadcasting\Channel;
use Illuminate\Broadcasting\InteractsWithSockets;
use Illuminate\Broadcasting\PrivateChannel;
use Illuminate\Contracts\Broadcasting\ShouldBroadcastNow;
use Illuminate\Foundation\Events\Dispatchable;
use Illuminate\Queue\SerializesModels;

class MessageSent implements ShouldBroadcastNow
{
    use Dispatchable, InteractsWithSockets, SerializesModels;

    public function __construct(public string $message)
    {
    }

    public function broadcastOn(): array
    {
        return [
            new PrivateChannel('chat'), // Use PrivateChannel for authenticated users
            // new Channel('chat'), // Use Channel for public access (no authentication)
        ];
    }
}
