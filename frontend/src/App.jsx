import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { AuthProvider } from "./contexts/AuthContext";
import Login from "./pages/Login";
import MoodEntry from "./pages/MoodEntry";
import MoodHistory from "./pages/MoodHistory";
import MoodChart from "./pages/MoodChart";
import TelegramConnect from "./pages/TelegramConnect";
import Navbar from "./components/Navbar";

function ProtectedRoute({ children }) {
  const token = localStorage.getItem("token");
  return token ? children : <Navigate to="/login" />;
}

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Navbar /> {/*дає посилання на сторінки й кнопку*/}
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route
            path="/entry"
            element={
              <ProtectedRoute> {/*відправляє на /login, якщо token відсутній*/}
                <MoodEntry />
              </ProtectedRoute>
            }
          />
          <Route
            path="/history"
            element={
              <ProtectedRoute>
                <MoodHistory />
              </ProtectedRoute>
            }
          />
          <Route
            path="/chart"
            element={
              <ProtectedRoute>
                <MoodChart />
              </ProtectedRoute>
            }
          />
          <Route
            path="/telegram"
            element={
              <ProtectedRoute>
                <TelegramConnect />
              </ProtectedRoute>
            }
          />
          <Route path="*" element={<Navigate to="/login" />} />
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  );
}
