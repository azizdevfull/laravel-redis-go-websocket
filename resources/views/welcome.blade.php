<!DOCTYPE html>
<html lang="uz">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Redis Broadcasting</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500&display=swap" rel="stylesheet">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.10.0/styles/github-dark.min.css">
    <script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.10.0/highlight.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.10.0/languages/go.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.10.0/languages/php.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.10.0/languages/bash.min.js"></script>
    <style>
        *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

        body {
            font-family: 'Inter', -apple-system, sans-serif;
            font-size: 15px;
            line-height: 1.75;
            color: #1c1e21;
            background: #f6f8fa;
        }

        .page {
            max-width: 900px;
            margin: 0 auto;
            padding: 3rem 1.5rem 5rem;
        }

        h1 {
            font-size: 2rem;
            font-weight: 700;
            color: #0f172a;
            margin-bottom: .5rem;
            padding-bottom: .75rem;
            border-bottom: 2px solid #e2e8f0;
        }

        h2 {
            font-size: 1.35rem;
            font-weight: 600;
            color: #0f172a;
            margin-top: 2.5rem;
            margin-bottom: .75rem;
            padding-bottom: .4rem;
            border-bottom: 1px solid #e2e8f0;
        }

        h3 {
            font-size: 1.05rem;
            font-weight: 600;
            color: #334155;
            margin-top: 1.75rem;
            margin-bottom: .5rem;
        }

        p { margin-bottom: .85rem; color: #374151; }

        ul, ol {
            margin: .5rem 0 .85rem;
            padding-left: 1.5rem;
            color: #374151;
        }

        li { margin-bottom: .25rem; }

        code {
            font-family: 'JetBrains Mono', 'Fira Code', monospace;
            font-size: .82em;
            background: #e8edf3;
            color: #c7254e;
            padding: .15em .4em;
            border-radius: 4px;
        }

        pre {
            margin: 1rem 0;
            border-radius: 8px;
            overflow: hidden;
            box-shadow: 0 2px 8px rgba(0,0,0,.12);
        }

        pre code {
            font-family: 'JetBrains Mono', 'Fira Code', monospace;
            font-size: .83rem;
            line-height: 1.65;
            background: none;
            color: inherit;
            padding: 1.1rem 1.3rem;
            display: block;
            overflow-x: auto;
        }

        table {
            width: 100%;
            border-collapse: collapse;
            margin: 1rem 0;
            font-size: .9rem;
        }

        th {
            background: #f1f5f9;
            font-weight: 600;
            color: #334155;
            padding: .6rem .9rem;
            text-align: left;
            border: 1px solid #e2e8f0;
        }

        td {
            padding: .55rem .9rem;
            border: 1px solid #e2e8f0;
            color: #374151;
        }

        tr:nth-child(even) td { background: #f8fafc; }

        hr {
            border: none;
            border-top: 1px solid #e2e8f0;
            margin: 2.25rem 0;
        }

        strong { color: #0f172a; font-weight: 600; }

        a { color: #2563eb; text-decoration: none; }
        a:hover { text-decoration: underline; }

        blockquote {
            border-left: 3px solid #6366f1;
            padding: .5rem 1rem;
            background: #f8f7ff;
            border-radius: 0 6px 6px 0;
            color: #475569;
            margin: 1rem 0;
        }
    </style>
</head>
<body>
    <div class="page">
        {!! $content !!}
    </div>
    <script>
        document.addEventListener('DOMContentLoaded', () => {
            document.querySelectorAll('pre code').forEach(block => {
                hljs.highlightElement(block);
            });
        });
    </script>
</body>
</html>
