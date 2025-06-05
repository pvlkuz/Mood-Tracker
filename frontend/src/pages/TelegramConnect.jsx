import { useState } from "react";
import api from "../api/axios";

export default function TelegramConnect() {
  const [chatId, setChatId] = useState("");
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");

  const handleSubmit = async (e) => {
    e.preventDefault();
    setMessage("");
    setError("");

    // Перевіримо, що chatId — непорожній та є числом
    if (!chatId || isNaN(Number(chatId))) {
      setError("Введіть коректний числовий chat_id");
      return;
    }

    try {
      await api.post("/user/telegram/register", { chat_id: Number(chatId) });
      setMessage("Telegram-чат успішно підключено! Тепер ви отримуватимете нотифікації.");
      setChatId("");
    } catch (err) {
      console.error(err);
      if (err.response && err.response.data) {
        setError(err.response.data);
      } else {
        setError("Не вдалося підключити Telegram. Спробуйте пізніше.");
      }
    }
  };

  return (
    <div className="app-container">
      <h2 className="page-title">Підключити Telegram-бот</h2>

      <div className="form-card">
        <p>
            Щоб отримувати щоденні й тижневі сповіщення у Telegram, необхідно пов’язати свій
            чат з ботом. Для цього виконайте два кроки:
        </p>
        <p>
          1. У Telegram знайдіть  бота @pvlkuz_moodtracker_bot, напишіть йому /start.
        </p>
        <p>
          2. Скопіюйте свій chat_id і вставте його в поле нижче.
            (Щоб дізнатися chat_id скористайтеся готовими ботами, наприклад @userinfobot, надішліть йому повідомлення, і він покаже ваш ID.)
        </p>

        <form onSubmit={handleSubmit} style={{ marginTop: "1rem", maxWidth: "400px" }}>
            <div>
            <label>Введіть chat_id:</label>
            <input
                type="text"
                value={chatId}
                onChange={(e) => setChatId(e.target.value.trim())}
                placeholder="Наприклад, 123456789"
                style={{ width: "100%", padding: "0.5rem", marginTop: "0.5rem" }}
            />
            </div>
            {error && <div style={{ color: "red", marginTop: "0.5rem" }}>{error}</div>}
            {message && <div style={{ color: "green", marginTop: "0.5rem" }}>{message}</div>}
            <button type="submit" style={{ marginTop: "1rem", padding: "0.5rem 1rem" }}>
            Підключити
            </button>
        </form>
      </div>
    </div>
  );
}
