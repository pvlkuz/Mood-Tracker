import { useState, useEffect } from "react";
import api from "../api/axios";
import Calendar from "react-calendar";
import "react-calendar/dist/Calendar.css";
import { format, parseISO, startOfMonth, endOfMonth, startOfWeek, endOfWeek} from "date-fns";
import IconPicker from "../components/IconPicker";

export default function MoodHistory() {
  const [selectedDate, setSelectedDate] = useState(null);// Обране значення дати
  const [moodMap, setMoodMap] = useState({});// Мапа всієї історії
  const [viewDate, setViewDate] = useState(new Date());// Поточний місяць, який показуємо в календарі
  const [showModal, setShowModal] = useState(false);// Стан для модалки (див. нижче)
  const [modalData, setModalData] = useState(null);// дані для модалки

  useEffect(() => {
    const loadMoods = async () => {
      // Визначаємо перше/останнє число поточного місяця:
      const monthStart = startOfMonth(viewDate); 
      const monthEnd = endOfMonth(viewDate);

      // Визначаємо, з якого понеділка починається календарний ряд (з попереднього місяця)
      const calendarStart = startOfWeek(monthStart, { weekStartsOn: 1 });
      // Визначаємо, до якої неділі показує календар (включно з клітинками наступного місяця)
      const calendarEnd = endOfWeek(monthEnd, { weekStartsOn: 1 });

      const from = format(calendarStart, "yyyy-MM-dd");
      const to = format(calendarEnd, "yyyy-MM-dd");

      try {
        const resp = await api.get(`/mood?from=${from}&to=${to}`);
        // Формуємо мапу: ключ — "YYYY-MM-DD", значення — обʼєкт mood
        const mMap = {};
        resp.data.forEach((m) => {
          const key = format(parseISO(m.date), "yyyy-MM-dd");
          mMap[key] = { id: m.id, icon: m.icon, comment: m.comment, date: key };
        });
        setMoodMap(mMap);
      } catch (err) {
        console.error("Не вдалося завантажити записи настрою:", err);
      }
    };

    loadMoods();
  }, [viewDate]);

  // Обробник кліку по дню:
  const handleDayClick = (date) => {
    const key = format(date, "yyyy-MM-dd");
    const existing = moodMap[key] || null;

    setSelectedDate(key);
    setModalData(existing);
    setShowModal(true);
  };

  // Збереження або оновлення запису в модалці
  const handleSave = async ({ id, icon, comment, date }) => {
    try {
      if (id) {
        // Оновлюємо
        await api.put(`/mood/${id}`, { icon, comment, date });
      } else {
        // Створюємо новий
        await api.post("/mood", { icon, comment, date });
      }
      // Після успіху перезавантажуємо записи для поточного місяця
      const monthStart = startOfMonth(viewDate); 
      const monthEnd = endOfMonth(viewDate);
      const calendarStart = startOfWeek(monthStart, { weekStartsOn: 1 });
      const calendarEnd = endOfWeek(monthEnd, { weekStartsOn: 1 });
      const from = format(calendarStart, "yyyy-MM-dd");
      const to = format(calendarEnd, "yyyy-MM-dd");

      const resp = await api.get(`/mood?from=${from}&to=${to}`);
      // Формуємо мапу: ключ — "YYYY-MM-DD", значення — обʼєкт mood
      const mMap = {};
      resp.data.forEach((m) => {
        const key = format(parseISO(m.date), "yyyy-MM-dd");
        mMap[key] = { id: m.id, icon: m.icon, comment: m.comment, date: key };
      });
      setMoodMap(mMap);
      setShowModal(false);
    } catch (err) {
      console.error("Не вдалося зберегти запис:", err);
      alert("Помилка при збереженні. Перевірте дані.");
    }
  };

  // Видалення запису в модалці
  const handleDelete = async (id) => {
    if (!window.confirm("Ви дійсно хочете видалити цей запис?")) return;
    try {
      await api.delete(`/mood/${id}`);
      // Перезавантажуємо після
      const monthStart = startOfMonth(viewDate); 
      const monthEnd = endOfMonth(viewDate);
      const calendarStart = startOfWeek(monthStart, { weekStartsOn: 1 });
      const calendarEnd = endOfWeek(monthEnd, { weekStartsOn: 1 });
      const from = format(calendarStart, "yyyy-MM-dd");
      const to = format(calendarEnd, "yyyy-MM-dd");

      const resp = await api.get(`/mood?from=${from}&to=${to}`);
      // Формуємо мапу: ключ — "YYYY-MM-DD", значення — обʼєкт mood
      const mMap = {};
      resp.data.forEach((m) => {
        const key = format(parseISO(m.date), "yyyy-MM-dd");
        mMap[key] = { id: m.id, icon: m.icon, comment: m.comment, date: key };
      });
      setMoodMap(mMap);
      setShowModal(false);
    } catch (err) {
      console.error("Не вдалося видалити запис:", err);
      alert("Помилка при видаленні.");
    }
  };

  return (
    <div className="app-container">
      <h2 className="page-title">Календар настроїв</h2>

      <div style={{ width: "100%" }}>
        <Calendar
          locale="uk"
          value={viewDate}
          //onActiveStartDateChange={({ activeStartDate }) => setViewDate(activeStartDate)} // Користувач перемикає місяці — оновлюємо viewDate, тригерить useEffect
          maxDate={new Date()}    // не даємо вибрати майбутні дні
          tileContent={({ date, view }) => {
            // Для view="month" додаємо іконку, якщо є запис
            if (view === "month") {
              const key = format(date, "yyyy-MM-dd");
              const rec = moodMap[key];
              if (rec) {
                return (
                  <div
                    style={{ textAlign: "center", marginTop: "2px", fontSize: "1.2rem" }}>
                    {rec.icon}
                  </div>
                );
              }
            }
            return null;
          }}
          onClickDay={handleDayClick}
          className="react-calendar"
        />
      </div>

      {/* Модалка для перегляду/редагування/видалення */}
      {showModal && (
        <div className="modal-backdrop" onClick={() => setShowModal(false)}>
          {/* Зупиняємо “закриття”, якщо клік по самому вікну модалки */}
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <h3>Деталі дня</h3>
            <ModalContent
              record={modalData}
              date={selectedDate}
              onSave={handleSave}
              onDelete={handleDelete}
              onCancel={() => setShowModal(false)}
            />
          </div>
        </div>
      )}
    </div>
  );
}

// Компонент для вмісту модалки (форма створення/редагування/видалення)
function ModalContent({ record, date, onSave, onDelete, onCancel }) {
  const [icon, setIcon] = useState(record ? record.icon : "");
  const [comment, setComment] = useState(record ? record.comment : "");
  const isEditing = Boolean(record && record.id);

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: "1rem" }}>
      <div>
        <label>Дата:</label>
        <div style={{ marginTop: "4px", fontWeight: "bold" }}>{date}</div>
      </div>

      <div>
        <label>Іконка:</label>
        <IconPicker onSelect={setIcon} defaultIcon={icon} />
      </div>

      <div>
        <label>Коментар:</label>
        <textarea
          rows="3"
          style={{ width: "100%", padding: "0.5rem", borderRadius: "4px", border: "1px solid #ccc" }}
          value={comment}
          onChange={(e) => setComment(e.target.value)}
        />
      </div>

      <div style={{ display: "flex", justifyContent: "flex-end", gap: "0.5rem" }}>
        {isEditing && (
          <button
            className="btn-small"
            style={{ backgroundColor: "#c0392b" }}
            onClick={() => onDelete(record.id)}
          >
            Видалити
          </button>
        )}
        <button
          className="btn-small"
          style={{ backgroundColor: "#7f8c8d" }}
          onClick={onCancel}
        >
          Скасувати
        </button>
        <button
          className="btn-small"
          style={{ backgroundColor: "#27ae60" }}
          onClick={() => onSave({ id: record?.id, icon, comment, date })}
        >
          {isEditing ? "Зберегти" : "Створити"}
        </button>
      </div>
    </div>
  );
}
