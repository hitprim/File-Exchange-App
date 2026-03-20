# File Exchange App

Приложение для мгновенной передачи буфера обмена и файлов между ПК и смартфоном по локальной сети — без кабелей, облаков и ручного ввода IP-адресов.

## Как это работает

```
ПК (Wails-приложение)  ──mDNS──►  автообнаружение в сети
        │
        │  HTTP по локальной сети
        ▼
Смартфон (Flutter-приложение)
```

1. Запускаешь приложение на ПК — оно регистрирует себя в локальной сети через **mDNS**
2. Открываешь мобильное приложение — оно автоматически находит ПК, никаких ручных настроек
3. Копируешь текст или файл на ПК → нажимаешь кнопку → получаешь на телефоне

## Стек

| Часть | Технологии |
|-------|-----------|
| Десктоп (этот репо) | Go · Wails · JavaScript |
| Мобильный клиент | Flutter · Dart |
| Обнаружение в сети | mDNS (multicast DNS) |

## Установка и запуск

### Требования

- Go 1.21+
- [Wails CLI](https://wails.io/docs/gettingstarted/installation): `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- Node.js 18+

### Запуск в режиме разработки

```bash
git clone https://github.com/hitprim/File-Exchange-App
cd File-Exchange-App
wails dev
```

### Сборка

```bash
wails build
```

Готовый `.exe` будет в папке `build/bin/`.

## Мобильный клиент

Репозиторий мобильного приложения (Flutter/Android): [File_Exchange_App_Mobile](https://github.com/hitprim/File_Exchange_App_Mobile)

## Зачем это нужно

Стандартные способы передачи текста с ПК на телефон неудобны: Telegram/WhatsApp требуют входа, AirDrop только на Apple, Bluetooth медленный. Это приложение решает задачу в одно нажатие — пока оба устройства в одной Wi-Fi сети.
