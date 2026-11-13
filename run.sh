#!/usr/bin/env bash

docker compose up -d --build
echo "✅ Приложение запущено!"
echo "Локальный доступ: http://localhost:8888"
echo "Если настроен Cloudflare Tunnel — используйте ваш домен."