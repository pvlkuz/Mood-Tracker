import { useState, useEffect } from "react";
import api from "../api/axios";
import IconPicker from "../components/IconPicker";
import { format } from "date-fns";

export default function MoodEntry() {
  const today = format(new Date(), "yyyy-MM-dd");

  const [icon, setIcon] = useState("");
  const [comment, setComment] = useState("");
  const [message, setMessage] = useState("");

  const [loading, setLoading] = useState(true);
  const [exists, setExists] = useState(false);
  const [existingMood, setExistingMood] = useState(null);

  useEffect(() => {
    // перевіряємо чи є сьогоднішній запис
    const checkToday = async () => {
      try {
        const resp = await api.get(`/mood?from=${today}&to=${today}`);
        if (Array.isArray(resp.data) && resp.data.length > 0) {
          setExists(true);
          setExistingMood(resp.data[0]);
        }
      } catch (err) {
        console.error(err);
        setMessage("Не вдалося перевірити сьогоднішній настрій");
      } finally {
        setLoading(false);
      }
    };
    checkToday();
  }, [today]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!icon) {
      setMessage("Оберіть іконку настрою");
      return;
    }
    try {
      await api.post("/mood", { icon, comment });
      setMessage("Сьогоднішній настрій успішно збережено!");
      setExists(true);
      setExistingMood({ icon, comment, date: today });
      setIcon("");
      setComment("");
    } catch (err) {
      console.error(err);
      setMessage("Не вдалося зберегти сьогоднішній настрій");
    }
  };

  if (loading) {
    return (
      <div className="app-container">
        <h2 className="page-title">Сьогоднішній настрій</h2>
        <p>Перевіряємо, чи є запис за сьогодні...</p>
      </div>
    );
  }

  return (
    <div className="app-container">
      <h2 className="page-title">Сьогоднішній настрій</h2>

      {exists ? (
        // Якщо вже є запис за сьогодні, показуємо
        <div className="form-card">
          <p className="message-error">Сьогоднішній настрій уже задано!</p>
          <div style={{ marginTop: "1rem" }}>
            <p>
              <strong>Дата:</strong> {today}
            </p>
            <p>
              <strong>Іконка:</strong>{" "}
              <span style={{ fontSize: "1.5rem" }}>{existingMood.icon}</span>
            </p>
            <p>
              <strong>Коментар:</strong> {existingMood.comment || "(немає)"}
            </p>
          </div>
        </div>
      ) : (
        // Якщо запису ще немає, показуємо форму
        <div className="form-card">
          <form onSubmit={handleSubmit}>
            <div>
              <label>Оберіть іконку:</label>
              <IconPicker onSelect={setIcon} />
            </div>

            <div style={{ marginTop: "1rem" }}>
              <label>Коментар:</label>
              <textarea
                rows="3"
                value={comment}
                onChange={(e) => setComment(e.target.value)}
              ></textarea>
            </div>

            {message && (
              <div
                className={
                  message.includes("успішно") ? "message-success" : "message-error"
                }
              >
                {message}
              </div>
            )}

            <button type="submit" style={{ marginTop: "1rem" }}>
              Зберегти
            </button>
          </form>
        </div>
      )}
    </div>
  );
}
