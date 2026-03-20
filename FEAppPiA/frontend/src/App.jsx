import { useEffect, useState } from 'react';
import './App.css';

// 👇 Импорт функций из Go (авто-генерируются при сборке Wails)
import {
    GetClipboardHistory,
    RestoreFromHistory,
    GetPhoneAuthToken,
    GetServerInfo
} from '../wailsjs/go/main/App';
import { EventsOn, EventsOff } from '../wailsjs/runtime/runtime';

function App() {
    const [history, setHistory] = useState([]);
    const [serverInfo, setServerInfo] = useState({ ip: '', port: '', token: '' });
    const [loading, setLoading] = useState(true);
    const [copiedIndex, setCopiedIndex] = useState(null);

    // Загрузка данных при старте
    useEffect(() => {
        loadData();

        // Подписка на события от Go (например, новое копирование в буфер)
        EventsOn('clipboard-updated', (newHistory) => {
            setHistory(newHistory);
        });

        return () => {
            EventsOff('clipboard-updated');
        };
    }, []);

    const loadData = async () => {
        setLoading(true);
        try {
            // Загружаем историю
            const hist = await GetClipboardHistory();
            setHistory(hist);

            // Загружаем инфо сервера
            const info = await GetServerInfo();
            setServerInfo(info);
        } catch (err) {
            console.error('Ошибка загрузки:', err);
        } finally {
            setLoading(false);
        }
    };

    // Восстановить запись в буфер ПК
    const handleRestore = async (index) => {
        await RestoreFromHistory(index);

        // Визуальный фидбек
        setCopiedIndex(index);
        setTimeout(() => setCopiedIndex(null), 1500);
    };

    // Копировать токен в буфер (для удобства)
    const copyToken = () => {
        navigator.clipboard.writeText(serverInfo.token);
        alert('✅ Токен скопирован!');
    };

    // Форматирование даты
    const formatDate = (ts) => {
        if (!ts) return '';
        const d = new Date(ts);
        return d.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' }) +
            ' • ' +
            d.toLocaleDateString('ru-RU');
    };

    return (
        <div className="app-container">
            {/* Шапка */}
            <header className="header">
                <h1>📋 PC Clipboard Exchange</h1>
                <button onClick={loadData} className="btn-refresh">⟳ Обновить</button>
            </header>

            <main className="main-content">
                {/* Блок: Информация для телефона */}
                <section className="card phone-connection">
                    <h2>📱 Подключение телефона</h2>

                    {loading ? (
                        <p>Загрузка...</p>
                    ) : (
                        <div className="connection-info">
                            <div className="info-row">
                                <label>IP адрес:</label>
                                <code>{serverInfo.ip || '—'}</code>
                            </div>
                            <div className="info-row">
                                <label>Порт:</label>
                                <code>{serverInfo.port || '—'}</code>
                            </div>
                            <div className="info-row">
                                <label>Токен:</label>
                                <div className="token-row">
                                    <code className="token">{serverInfo.token || '—'}</code>
                                    <button onClick={copyToken} className="btn-small">📋</button>
                                </div>
                            </div>

                            <div className="qr-hint">
                                <small>💡 Откройте в браузере телефона:</small>
                                <code className="url">
                                    http://{serverInfo.ip}:{serverInfo.port}/api/clipboard?token={serverInfo.token}
                                </code>
                            </div>
                        </div>
                    )}
                </section>

                {/* Блок: История буфера */}
                <section className="card clipboard-history">
                    <h2>🗂️ История буфера ({history.length})</h2>

                    {loading ? (
                        <div className="loading">Загрузка истории...</div>
                    ) : history.length === 0 ? (
                        <div className="empty-state">
                            <p>📭 Буфер пуст</p>
                            <small>Скопируйте что-нибудь на ПК или отправьте с телефона</small>
                        </div>
                    ) : (
                        <ul className="history-list">
                            {history.map((item, index) => (
                                <li key={index} className="history-item">
                                    <div className="item-content">
                                        <p className="content-text" title={item.content}>
                                            {item.content.length > 80
                                                ? item.content.substring(0, 80) + '…'
                                                : item.content}
                                        </p>
                                        <div className="item-meta">
                                            <span className="type-badge">{item.type}</span>
                                            <span className="timestamp">{formatDate(item.timestamp)}</span>
                                        </div>
                                    </div>

                                    <button
                                        onClick={() => handleRestore(index)}
                                        className={`btn-restore ${copiedIndex === index ? 'copied' : ''}`}
                                        title="Восстановить в буфер ПК"
                                    >
                                        {copiedIndex === index ? '✓ Скопировано!' : '📋 В буфер'}
                                    </button>
                                </li>
                            ))}
                        </ul>
                    )}
                </section>
            </main>

            {/* Футер */}
            <footer className="footer">
                <small>Wails Clipboard Exchange • v1.0</small>
            </footer>
        </div>
    );
}

export default App;