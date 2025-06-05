import { useEffect, useState } from "react";
import api from "../api/axios";
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from "recharts";
import { format, parseISO, subDays, subMonths } from "date-fns";

const iconToValue = (icon) => {
  switch (icon) {
    case "😡":
        return 0;
    case "😢":
        return 1;
    case "😞":
        return 2;
    case "😐":
        return 3;
    case "😊":
        return 4;
    case "😃":
        return 5;
    default:
      return null;
  }
};

const valueToIcon = (val) => {
  switch (val) {
    case 0:
        return "😡";
    case 1:
        return "😢";
    case 2:
        return "😞";
    case 3:
        return "😐";
    case 4:
        return "😊";
    case 5:
        return "😃";
    default:
        return "";
  }
};

export default function MoodChart() {
  const todayStr = format(new Date(), "yyyy-MM-dd");
  const oneWeekAgoStr = format(subDays(new Date(), 6), "yyyy-MM-dd");
  const oneMonthAgoStr = format(subMonths(new Date(), 1), "yyyy-MM-dd");

  // Початковий стан: останні 7 днів
  const [fromDate, setFromDate] = useState(oneWeekAgoStr);
  const [toDate, setToDate] = useState(todayStr);
  const [data, setData] = useState([]);
  const [error, setError] = useState("");

  useEffect(() => {
    loadChartData(oneWeekAgoStr, todayStr);
  }, []);

  const loadChartData = async (from, to) => {
    try {
      const resp = await api.get(`/mood?from=${from}&to=${to}`);
      const moods = resp.data;

      // Обчислюємо кількість днів у проміжку (включно)
      const start = parseISO(from);
      const end = parseISO(to);
      const diffMs = end.getTime() - start.getTime();
      const dayCount = Math.floor(diffMs / (1000 * 60 * 60 * 24)) + 1;

      // Ініціалізуємо обʼєкт із ключами "yyyy-MM-dd" і значенням null
      const temp = {};
      for (let i = 0; i < dayCount; i++) {
        const d = new Date(start.getTime() + i * 86400000);
        const key = format(d, "yyyy-MM-dd");
        temp[key] = null;
      }

      // Заповнюємо temp з отриманих записів
      moods.forEach((m) => {
        const key = format(parseISO(m.date), "yyyy-MM-dd");
        if (temp[key] !== undefined) {
          temp[key] = iconToValue(m.icon);
        }
      });

      // Перетворюємо temp у масив для Recharts
      const chartData = Object.entries(temp).map(([day, val]) => ({
        date: format(parseISO(day), "dd.MM"),
        value: val,
      }));

      setData(chartData);
      setError("");
    } catch (err) {
      console.error(err);
      setError("Не вдалося завантажити дані для графіка");
      setData([]);
    }
  };

  // Кнопка "Застосувати" перезапитує дані за вказаний діапазон
  const handleApply = () => {
    loadChartData(fromDate, toDate);
  };

  return (
    <div className="app-container">
      <h2 className="page-title">Графік настрою</h2>

      {/* Блок вибору діапазону */}
      <div style={{ marginBottom: "1rem", display: "flex", alignItems: "flex-end", gap: "1rem" }}>
        <div>
          <label>Від:</label>
          <br />
          <input
            type="date"
            value={fromDate}
            // Щоб користувач не міг обрати дату раніше, ніж за місяць тому
            min={oneMonthAgoStr} 
            max={toDate}       
            onChange={(e) => setFromDate(e.target.value)}
          />
        </div>

        <div>
          <label>До:</label>
          <br />
          <input
            type="date"
            value={toDate}
            // Щоб користувач не міг обрати дату пізніше за сьогодні
            min={fromDate}    
            max={todayStr}    
            onChange={(e) => setToDate(e.target.value)}
          />
        </div>

        <button onClick={handleApply} style={{ height: "2.5rem" }}>
          Застосувати
        </button>
      </div>

      {error && <div className="message-error">{error}</div>}

      <div className="chart-container">
        <ResponsiveContainer width="100%" height={500}>
          <LineChart data={data} margin={{ top: 20, right: 30, bottom: 60, left: 5 }}>
            <CartesianGrid strokeDasharray="3 3" />

            <XAxis
              dataKey="date"
              tick={{ fontSize: 12 }}
              interval={0}
              angle={-45}
              textAnchor="end"
              height={60}
            />

            <YAxis
              type="number"
              domain={[0, 5]}
              ticks={[0, 1, 2, 3, 4, 5]}
              tickFormatter={(val) => valueToIcon(val)}
              tick={{ fontSize: 18 }}
            />

            <Line
              type="monotone"
              dataKey="value"
              connectNulls
              stroke="#8884d8"
              strokeWidth={2}
              dot={{ r: 5 }}
              isAnimationActive={false}
            />
          </LineChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}
